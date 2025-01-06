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

type AnkiCardGeneratorPlugin interface {
	GenerateAnkiCards(ctx context.Context, query string) ([]ai.AnkiCard, error)
	StoreAnkiCards(deckName string, cards []ai.AnkiCard) error
	ChooseDeck(ctx context.Context, query string) (string, error)
}

type BasePlugin struct {
	ankiClient anki.AnkiClienter
	ankiAI     ai.AnkiController
	deckName   string
	ankiCards  []ai.AnkiCard
}

func NewBasePlugin(c ai.AnkiController) *BasePlugin {
	return &BasePlugin{
		ankiClient: anki.NewClient(lib.GetEnv("ANKI_CONNECT_URL", "http://localhost:8765")),
		ankiAI:     c,
	}
}

func (t *BasePlugin) getFilteredDeckNames(filter ...string) ([]string, error) {
	deckNames, err := t.listBaseDeckNames()
	if err != nil {
		return nil, fmt.Errorf("get filtered deck names: %w", err)
	}

	decks, err := t.filterDecksByName(deckNames, filter...)
	if err != nil {
		return nil, fmt.Errorf("get filtered deck names: %w", err)
	}
	return decks, nil
}

// listBaseDeckNames - this is an very big assumption that people who use Anki have sub folders.
// This returns all decks that don't have `::` in the name.
func (t *BasePlugin) listBaseDeckNames() ([]string, error) {
	deckNames, err := t.ankiClient.DeckNames().GetNames()
	if err != nil {
		return nil, fmt.Errorf("list base deck names: %w", err)
	}
	return anki.FilterDecksByHierarchy(deckNames), nil
}

func (t *BasePlugin) filterDecksByName(deckNames []string, filter ...string) ([]string, error) {
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

func (t *BasePlugin) generateAnkiCards(ctx context.Context, query string, prompt string) ([]ai.AnkiCard, error) {
	cards, err := t.ankiAI.GenerateAnkiCards(
		ctx,
		t.deckName,
		query,
		prompt,
	)
	if err != nil {
		return nil, fmt.Errorf("generate anki cards: %w", err)
	}
	return cards, nil
}

func (t *BasePlugin) chooseDeck(ctx context.Context, query string, decks []string, createIfNotExists bool) (string, error) {
	deckName, err := t.ankiAI.ChooseDeck(ctx, decks, fmt.Sprintf("Which deck should I use for the topic: %s", query))
	if err != nil {
		return "", fmt.Errorf("choose deck: %w", err)
	}

	if createIfNotExists {
		if !slices.Contains(decks, deckName) {
			if err := t.ankiClient.DeckNames().Create(deckName); err != nil {
				return "", fmt.Errorf("choose deck (%s): %w", deckName, err)
			}
		}
	}

	t.deckName = deckName
	return deckName, nil
}

func (t *BasePlugin) chooseFilteredDeck(ctx context.Context, deckFilter string, query string) (string, error) {
	if query == "" {
		return "", ErrQueryRequired
	}

	decks, err := t.getFilteredDeckNames(deckFilter)
	if err != nil {
		return "", err
	}

	// We're using AI for both of these steps. The context is being passed into the OpenAI structured ouput struct.
	// It's a single shot program currently, so this should be adequate enough error handling and reporting.
	deckName, err := t.chooseDeck(ctx, query, decks, true)
	if err != nil {
		return "", fmt.Errorf("choose filtered deck: %w", err)
	}

	return deckName, nil
}

func PrintCards(ac []ai.AnkiCard, padding bool) {
	if padding {
		fmt.Println("")
	}
	for _, c := range ac {
		fmt.Print(colors.BeautifyCard(c))
	}
	if padding {
		fmt.Println("")
	}
}

// ==================================================
// TopicPlugin
// ==================================================

var (
	ErrQueryRequired = fmt.Errorf("query is required")
)

type TopicPlugin struct {
	*BasePlugin
}

func newTopicPlugin(c ai.AnkiController) AnkiCardGeneratorPlugin {
	t := &TopicPlugin{
		BasePlugin: NewBasePlugin(c),
	}
	return t
}

func (t *TopicPlugin) ChooseDeck(ctx context.Context, query string) (string, error) {
	return t.BasePlugin.chooseFilteredDeck(ctx, "Haki", query)
}

func (t *TopicPlugin) GenerateAnkiCards(ctx context.Context, query string) ([]ai.AnkiCard, error) {
	cards, err := t.generateAnkiCards(ctx, query, generateAnkiCardPrompt())
	if err != nil {
		return nil, fmt.Errorf("topic: %w", err)
	}
	slog.Info("anki card(s) generated", slog.Int("count", len(cards)))

	return cards, nil
}

func (t *TopicPlugin) StoreAnkiCards(deckName string, cards []ai.AnkiCard) error {
	const modelName = "Basic"
	for _, c := range cards {
		data := map[string]interface{}{
			"Front": c.Front,
			"Back":  formatBack(c.Back),
		}

		note := anki.NewNoteBuilder(deckName, modelName, data).Build()

		id, err := t.ankiClient.Notes().Add(note)
		if err != nil {
			return fmt.Errorf("topic: %w", err)
		}
		slog.Info(
			"note added",
			slog.String("deck", deckName),
			slog.String("model", modelName),
			slog.String("id", fmt.Sprintf("%.f", id)),
		)
	}
	return nil
}

// ==================================================
// VocabPlugin
// ==================================================

type VocabPlugin struct {
	*BasePlugin
	ttsService      ai.TTS
	imageGenService ai.ImageGen
	word            string
	ttsFilePath     string
	imageFilePath   string
	outputDir       string
}

func newVocabPlugin(cardCreator ai.AnkiController, ttsService ai.TTS, imageGenService ai.ImageGen, outputDir string) AnkiCardGeneratorPlugin {
	e := &VocabPlugin{
		BasePlugin:      NewBasePlugin(cardCreator),
		ttsService:      ttsService,
		imageGenService: imageGenService,
		outputDir:       outputDir,
	}
	return e
}

func (v *VocabPlugin) ChooseDeck(ctx context.Context, query string) (string, error) {
	return v.BasePlugin.chooseFilteredDeck(ctx, "Vocabulary", query)
}

func (v *VocabPlugin) GenerateAnkiCards(ctx context.Context, query string) ([]ai.AnkiCard, error) {
	cards, err := v.generateAnkiCards(ctx, query, generateAnkiCardPrompt())
	if err != nil {
		return nil, fmt.Errorf("vocab: %w", err)
	}
	slog.Info("anki card(s) created", slog.Int("count", len(v.ankiCards)))

	if _, err := v.generateTTS(ctx, query); err != nil {
		slog.Error("vocab: create tts", slog.String("error", err.Error()))
	} else {
		slog.Info("tts created", slog.String("file_path", v.ttsFilePath))
	}

	if _, err := v.generateImage(ctx, query); err != nil {
		slog.Error("vocab: create image", slog.String("error", err.Error()))
	} else {
		slog.Info("image created", slog.String("file_path", v.imageFilePath))
	}

	v.word = query
	return cards, nil
}

func (v *VocabPlugin) StoreAnkiCards(deckName string, cards []ai.AnkiCard) error {
	const modelName = "VocabularyWithAudio"
	for _, c := range cards {
		note, err := v.buildNote(modelName, c)
		if err != nil {
			slog.Error("failed building note",
				slog.String("deck", deckName),
				slog.String("model", modelName),
				slog.String("error", err.Error()),
			)
			continue
		}

		id, err := v.ankiClient.Notes().Add(note)
		if err != nil {
			return fmt.Errorf("vocab: %w", err)
		}
		slog.Info(
			"note added",
			slog.String("deck", deckName),
			slog.String("model", modelName),
			slog.String("id", fmt.Sprintf("%.f", id)),
		)
	}
	return nil
}

func (v *VocabPlugin) hasImage() bool {
	return strings.TrimSpace(v.imageFilePath) != ""
}

func (v *VocabPlugin) hasTTS() bool {
	return strings.TrimSpace(v.ttsFilePath) != ""
}

func (v *VocabPlugin) generateTTS(ctx context.Context, query string) (string, error) {
	mp3Bytes, err := v.ttsService.GenerateMP3(ctx, query)
	if err != nil {
		return "", fmt.Errorf("generate mp3: %w", err)
	}
	path, err := makeTTSFilePath(v.outputDir, query)
	if err != nil {
		return "", fmt.Errorf("make file path: %w", err)
	}
	err = lib.SaveFile(path, mp3Bytes)
	if err != nil {
		return "", fmt.Errorf("save file: %w", err)
	}
	v.ttsFilePath = path
	return path, nil
}

func makeTTSFilePath(outputDir, word string) (string, error) {
	return filepath.Abs(fmt.Sprintf("%s/data/%s.mp3", outputDir, word))
}

func (v *VocabPlugin) generateImage(ctx context.Context, query string) (string, error) {
	imgBytes, err := v.imageGenService.Generate(ctx, query)
	if err != nil {
		return "", fmt.Errorf("generate image: %w", err)
	}
	path, err := makeImageFilePath(v.outputDir, query)
	if err != nil {
		return "", fmt.Errorf("make file path: %w", err)
	}
	err = lib.SaveFile(path, imgBytes)
	if err != nil {
		return "", fmt.Errorf("save file: %w", err)
	}
	v.imageFilePath = path
	return "", nil
}
func makeImageFilePath(outputDir, word string) (string, error) {
	return filepath.Abs(fmt.Sprintf("%s/data/%s.webp", outputDir, word))
}

func (v *VocabPlugin) buildNote(modelName string, c ai.AnkiCard) (anki.Note, error) {
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
