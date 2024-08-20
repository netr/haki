package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/netr/haki/ai"
	"github.com/netr/haki/anki"
	"github.com/netr/haki/lib"
	"github.com/urfave/cli/v2"
)

func NewVocabCommand(apiKey string) *cli.Command {
	return &cli.Command{
		Name:      "vocab",
		Usage:     "Create a vocabulary Anki card using the specified word.",
		ArgsUsage: "--word <word>",
		Flags:     []cli.Flag{newWordFlag()},
		Action:    actionVocab(apiKey),
	}
}

func actionVocab(apiKey string) func(cCtx *cli.Context) error {
	return func(cCtx *cli.Context) error {
		word := cCtx.String("word")
		if word == "" {
			return fmt.Errorf("word is required --word <word>")
		}

		if err := runVocab(apiKey, word); err != nil {
			slog.Error("run", slog.String("action", "vocab"), slog.String("error", err.Error()))
			return err
		}
		return nil
	}
}

func runVocab(apiKey, word string) error {
	ankiClient := anki.NewClient(lib.GetEnv("ANKI_CONNECT_URL", "http://localhost:8765"))
	cardCreator, err := ai.NewAICardCreator(ai.OpenAI, apiKey)
	if err != nil {
		return fmt.Errorf("new openai api provider: %w", err)
	}
	ttsService := ai.NewTTSService(apiKey)
	vocabEntity := newVocabularyEntity(ankiClient, cardCreator, ttsService)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := vocabEntity.Create(ctx, word); err != nil {
		return fmt.Errorf("create vocab entity: %w", err)
	}

	return nil
}

type VocabularyEntity struct {
	ankiClient   anki.AnkiClienter
	cardCreator  ai.AICardCreator
	ttsService   ai.TTS
	word         string
	deckName     string
	ttsAudioPath string
	cards        []ai.AnkiCard
}

func newVocabularyEntity(ankiClient anki.AnkiClienter, cardCreator ai.AICardCreator, ttsService ai.TTS) *VocabularyEntity {
	v := &VocabularyEntity{
		ankiClient:  ankiClient,
		cardCreator: cardCreator,
		ttsService:  ttsService,
	}

	return v
}

func (v *VocabularyEntity) Create(ctx context.Context, word string) error {
	slog.Info("creating vocabulary card", slog.String("word", word))

	if word == "" {
		return fmt.Errorf("word is required")
	}
	v.word = word

	if err := v.chooseDeck(ctx); err != nil {
		return fmt.Errorf("choose deck: %w", err)
	}
	slog.Info("deck name chosen", slog.String("deck", v.deckName))

	if err := v.createAnkiCards(ctx); err != nil {
		return fmt.Errorf("create anki cards: %w", err)
	}
	slog.Info("anki card(s) created", slog.Int("count", len(v.cards)))

	if err := v.createTTS(ctx); err != nil {
		return fmt.Errorf("create tts: %w", err)
	}
	slog.Info("tts created", slog.String("output_path", v.ttsAudioPath))

	// TODO: FIXME: make sure this model is created before using the app. Preferably something like "Haki::VocabularyWithAudio"
	const modelName = "VocabularyWithAudio"
	for _, c := range v.cards {
		data := map[string]interface{}{
			"Question":   c.Front,
			"Definition": c.Back,
			"Audio":      fmt.Sprintf("[sound:%s.mp3]", v.word),
		}

		note := anki.NewNoteBuilder(v.deckName, modelName, data).
			WithAudio(
				v.ttsAudioPath,
				fmt.Sprintf("%s.mp3", v.word),
				"Front",
			).Build()

		id, err := v.ankiClient.Notes().Add(note)
		if err != nil {
			return fmt.Errorf("add note: %w", err)
		}
		slog.Info(
			"note added",
			slog.String("deck", v.deckName),
			slog.String("model", modelName),
			slog.String("id", fmt.Sprintf("%.f", id)),
		)
	}

	for _, c := range v.cards {
		colors.BeautifyCard(c)
	}
	return nil
}

func (v *VocabularyEntity) createTTS(ctx context.Context) error {
	mp3, err := v.ttsService.GenerateMP3(ctx, v.word)
	if err != nil {
		return fmt.Errorf("generate mp3: %w", err)
	}
	err = lib.SaveFile(fmt.Sprintf("data/%s.mp3", v.word), mp3)
	if err != nil {
		return fmt.Errorf("save file: %w", err)
	}
	path, err := filepath.Abs(fmt.Sprintf("data/%s.mp3", v.word))
	if err != nil {
		return fmt.Errorf("get absolute path: %w", err)
	}
	v.ttsAudioPath = path
	return nil
}

func (v *VocabularyEntity) createAnkiCards(ctx context.Context) error {
	cards, err := v.cardCreator.Create(
		ctx,
		v.deckName,
		"Create a vocabulary card (with parts of speech ONLY on front) for the word: "+v.word+".",
	)
	if err != nil {
		return fmt.Errorf("create anki cards: %w", err)
	}

	v.cards = cards
	return nil
}

func (v *VocabularyEntity) chooseDeck(ctx context.Context) error {
	decks, err := v.getVocabDecks("Vocabulary")
	if err != nil {
		return fmt.Errorf("get vocabulary deck names: %w", err)
	}

	deckName, err := v.cardCreator.ChooseDeck(ctx, decks, fmt.Sprintf("Which vocabulary deck should I use for the word: %s", v.word))
	if err != nil {
		return fmt.Errorf("choose deck: %w", err)
	}

	v.deckName = deckName
	return nil
}

func (v *VocabularyEntity) getDecks() ([]string, error) {
	deckNames, err := v.ankiClient.DeckNames().GetNames()
	if err != nil {
		return nil, fmt.Errorf("get deck names: %w", err)
	}
	return anki.RemoveParentDecks(deckNames), nil
}

func (v *VocabularyEntity) getVocabDecks(filter ...string) ([]string, error) {
	deckNames, err := v.getDecks()
	if err != nil {
		return nil, fmt.Errorf("get deck names: %w", err)
	}

	if len(filter) == 0 {
		return deckNames, nil
	}

	// only use deck names that have Vocabulary in them
	var decks []string
	for _, d := range deckNames {
		if strings.Contains(d, filter[0]) {
			decks = append(decks, d)
		}
	}
	return decks, nil
}
