package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/netr/haki/ai"
	"github.com/netr/haki/anki"
	"github.com/netr/haki/lib"
)

type CardCreatorEntity interface {
	CreateCards(ctx context.Context, query string, prompt string, skipSave bool) error
}

type BaseEntity struct {
	ankiClient  anki.AnkiClienter
	cardCreator ai.CardCreator
	deckName    string
	cards       []ai.AnkiCard
}

func NewBaseEntity(cardCreator ai.CardCreator) *BaseEntity {
	return &BaseEntity{
		ankiClient:  anki.NewClient(lib.GetEnv("ANKI_CONNECT_URL", "http://localhost:8765")),
		cardCreator: cardCreator,
	}
}

func (t *BaseEntity) getFilteredDeckNames(filter ...string) ([]string, error) {
	deckNames, err := t.listBaseDeckNames()
	if err != nil {
		return nil, fmt.Errorf("get deck names: %w", err)
	}

	decks, err := t.filterDecksByName(deckNames, filter...)
	if err != nil {
		return nil, fmt.Errorf("get filtered deck names: %w", err)
	}
	return decks, nil
}

// listBaseDeckNames - this is an very big assumption that people who use Anki have sub folders.
// This returns all decks that don't have `::` in the name.
func (t *BaseEntity) listBaseDeckNames() ([]string, error) {
	deckNames, err := t.ankiClient.DeckNames().GetNames()
	if err != nil {
		return nil, fmt.Errorf("get deck names: %w", err)
	}
	return anki.FilterDecksByHierarchy(deckNames), nil
}

func (t *BaseEntity) filterDecksByName(deckNames []string, filter ...string) ([]string, error) {
	if len(filter) == 0 {
		return deckNames, nil
	}

	// only use deck names that have [filter] in them
	var decks []string
	for _, d := range deckNames {
		if strings.Contains(d, filter[0]) {
			decks = append(decks, d)
		}
	}
	return decks, nil
}

// TODO: FIXME: make sure this model is created before using the app. Preferably something like "Haki::VocabularyWithAudio"
func (t *BaseEntity) saveCardsToAnki() error {
	const modelName = "Basic"
	for _, c := range t.cards {
		data := map[string]interface{}{
			"Front": c.Front,
			"Back":  formatBack(c.Back),
		}

		note := anki.NewNoteBuilder(t.deckName, modelName, data).Build()

		id, err := t.ankiClient.Notes().Add(note)
		if err != nil {
			return fmt.Errorf("add note: %w", err)
		}
		slog.Info(
			"note added",
			slog.String("deck", t.deckName),
			slog.String("model", modelName),
			slog.String("id", fmt.Sprintf("%.f", id)),
		)
	}
	return nil
}

func (t *BaseEntity) createAnkiCards(ctx context.Context, query string, prompt string) error {
	cards, err := t.cardCreator.Create(
		ctx,
		t.deckName,
		query,
		prompt,
	)
	if err != nil {
		return fmt.Errorf("create anki cards: %w", err)
	}

	t.cards = cards
	return nil
}

func (t *BaseEntity) chooseDeck(ctx context.Context, query string, decks []string, createIfNotExists bool) error {
	deckName, err := t.cardCreator.ChooseDeck(ctx, decks, fmt.Sprintf("Which deck should I use for the topic: %s", query))
	if err != nil {
		return fmt.Errorf("choose deck: %w", err)
	}

	t.deckName = deckName

	if createIfNotExists {
		if err = t.CreateDeck(decks); err != nil {
			return fmt.Errorf("create deck: %w", err)
		}
	}
	return nil
}

func (t *BaseEntity) PrintCards(padding bool) {
	if padding {
		fmt.Println("")
	}
	for _, c := range t.cards {
		fmt.Print(colors.BeautifyCard(c))
	}
	if padding {
		fmt.Println("")
	}
}

func (t *BaseEntity) CreateDeck(decks []string) error {
	if !slices.Contains(decks, t.deckName) {
		if err := t.ankiClient.DeckNames().Create(t.deckName); err != nil {
			return fmt.Errorf("create deck (%s): %w", t.deckName, err)
		}
	}
	return nil
}

// ==================================================
// TopicEntity
// ==================================================

var (
	ErrQueryRequired     = fmt.Errorf("query is required")
	ErrInvalidOutputPath = fmt.Errorf("invalid output path")
)

type TopicEntity struct {
	*BaseEntity
	topic string
}

func newTopicEntity(cardCreator ai.CardCreator) CardCreatorEntity {
	t := &TopicEntity{
		BaseEntity: NewBaseEntity(cardCreator),
	}
	return t
}

func (t *TopicEntity) CreateCards(ctx context.Context, query string, prompt string, skipSave bool) error {
	slog.Info("creating topic card", slog.String("topic", query))

	if query == "" {
		return ErrQueryRequired
	}
	t.topic = query

	decks, err := t.getFilteredDeckNames("Haki")
	if err != nil {
		return err
	}

	// We're using AI for both of these steps. The context is being passed into the OpenAI structured ouput struct.
	// It's a single shot program currently, so this should be adequate enough error handling and reporting.
	if err := t.chooseDeck(ctx, t.topic, decks, true); err != nil {
		return fmt.Errorf("choose deck: %w", err)
	}
	slog.Info("deck name chosen", slog.String("deck", t.deckName))

	if err := t.createAnkiCards(ctx, t.topic, prompt); err != nil {
		return fmt.Errorf("create anki cards: %w", err)
	}
	slog.Info("anki card(s) created", slog.Int("count", len(t.cards)))

	if !skipSave {
		if err := t.saveCardsToAnki(); err != nil {
			return fmt.Errorf("create anki cards: %w", err)
		}
	}

	t.PrintCards(true)
	return nil
}

// ==================================================
// VocabEntity
// ==================================================

var (
	ErrWordRequired = fmt.Errorf("word is required")
)

type VocabEntity struct {
	*BaseEntity
	ttsService      ai.TTS
	imageGenService ai.ImageGen
	word            string
	ttsFilePath     string
	imageFilePath   string
	outputDir       string
}

func newVocabEntity(cardCreator ai.CardCreator, ttsService ai.TTS, imageGenService ai.ImageGen, outputDir string) CardCreatorEntity {
	e := &VocabEntity{
		BaseEntity:      NewBaseEntity(cardCreator),
		ttsService:      ttsService,
		imageGenService: imageGenService,
		outputDir:       outputDir,
	}
	return e
}

func (v *VocabEntity) CreateCards(ctx context.Context, query string, prompt string, skipSave bool) error {
	slog.Info("creating vocabulary card", slog.String("query", query))

	if query == "" {
		return ErrWordRequired
	}
	v.word = query

	decks, err := v.getFilteredDeckNames("Vocabulary")
	if err != nil {
		return err
	}

	if err := v.chooseDeck(ctx, v.word, decks, true); err != nil {
		return fmt.Errorf("choose deck: %w", err)
	}
	slog.Info("deck name chosen", slog.String("deck", v.deckName))

	if err := v.createAnkiCards(ctx, v.word, prompt); err != nil {
		return fmt.Errorf("create anki cards: %w", err)
	}
	slog.Info("anki card(s) created", slog.Int("count", len(v.cards)))

	if err := v.createTTS(ctx); err != nil {
		slog.Error("create tts", slog.String("error", err.Error()))
	} else {
		slog.Info("tts created", slog.String("file_path", v.ttsFilePath))
	}

	if err := v.createImage(ctx); err != nil {
		slog.Error("create image", slog.String("error", err.Error()))
	} else {
		slog.Info("image created", slog.String("file_path", v.imageFilePath))
	}

	if !skipSave {
		const modelName = "VocabularyWithAudio"
		for _, c := range v.cards {
			note, err := v.createNote(modelName, c)
			if err != nil {
				slog.Error("failed adding note",
					slog.String("deck", v.deckName),
					slog.String("model", modelName),
					slog.String("error", err.Error()),
				)
				continue
			}

			id, err := v.ankiClient.Notes().Add(note)
			if err != nil {
				return fmt.Errorf("add vocab note: %w", err)
			}
			slog.Info(
				"note added",
				slog.String("deck", v.deckName),
				slog.String("model", modelName),
				slog.String("id", fmt.Sprintf("%.f", id)),
			)
		}
	}

	v.PrintCards(true)
	return nil
}

func (v *VocabEntity) hasImage() bool {
	return strings.TrimSpace(v.imageFilePath) != ""
}

func (v *VocabEntity) hasTTS() bool {
	return strings.TrimSpace(v.ttsFilePath) != ""
}

func (v *VocabEntity) createTTS(ctx context.Context) error {
	mp3Bytes, err := v.ttsService.GenerateMP3(ctx, v.word)
	if err != nil {
		return fmt.Errorf("generate mp3: %w", err)
	}
	path, err := makeTTSFilePath(v.outputDir, v.word)
	if err != nil {
		return fmt.Errorf("make file path: %w", err)
	}
	err = lib.SaveFile(path, mp3Bytes)
	if err != nil {
		return fmt.Errorf("save file: %w", err)
	}
	v.ttsFilePath = path
	return nil
}

func makeTTSFilePath(outputDir, word string) (string, error) {
	return filepath.Abs(fmt.Sprintf("%s/data/%s.mp3", outputDir, word))
}

func (v *VocabEntity) createImage(ctx context.Context) error {
	imgBytes, err := v.imageGenService.Generate(ctx, v.word)
	if err != nil {
		return fmt.Errorf("generate image: %w", err)
	}
	path, err := makeImageFilePath(v.outputDir, v.word)
	if err != nil {
		return fmt.Errorf("make file path: %w", err)
	}
	err = lib.SaveFile(path, imgBytes)
	if err != nil {
		return fmt.Errorf("save file: %w", err)
	}
	v.imageFilePath = path
	return nil
}
func makeImageFilePath(outputDir, word string) (string, error) {
	return filepath.Abs(fmt.Sprintf("%s/data/%s.webp", outputDir, word))
}

func (v *VocabEntity) createNote(modelName string, c ai.AnkiCard) (anki.Note, error) {
	data := map[string]interface{}{
		"Question":   c.Front,
		"Definition": formatBack(c.Back),
	}

	note := anki.NewNoteBuilder(v.deckName, modelName, data)

	if v.hasTTS() {
		note.
			SetField(
				"Audio",
				createAudioTag(makeTTSFileName(v.word)),
			).
			WithAudio(
				v.ttsFilePath,
				makeTTSFileName(v.word),
				"Front",
			)
	}

	if v.hasImage() {
		note.
			SetField(
				"Picture",
				createImageTag(makeImageFileName(v.word)),
			).
			WithPicture(
				"",
				v.imageFilePath,
				makeImageFileName(v.word),
				"Front",
			)
	}

	return note.Build(), nil
}

func createImageTag(fileName string) string {
	return fmt.Sprintf("<img src=\"%s\">", fileName)
}

func createAudioTag(fileName string) string {
	return fmt.Sprintf("[sound:%s]", fileName)
}

func makeImageFileName(name string) string {
	if strings.HasSuffix(name, ".webp") {
		return name
	}
	return fmt.Sprintf("%s.webp", name)
}

func makeTTSFileName(name string) string {
	if strings.HasSuffix(name, ".mp3") {
		return name
	}
	return fmt.Sprintf("%s.mp3", name)
}

func replaceBold(text string) string {
	re := regexp.MustCompile(`\*\*(.*?)\*\*`)
	return re.ReplaceAllString(text, `<b>$1</b>`)
}

func formatBack(data string) string {
	if !strings.Contains(strings.ToLower(data), "<") {
		data = strings.ReplaceAll(data, "\n", "<br>\n")
	}
	data = strings.ReplaceAll(data, "\\<", "<")
	data = replaceBold(data)
	return data
}
