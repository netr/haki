package main

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/netr/haki/ai"
	"github.com/netr/haki/anki"
	"github.com/netr/haki/lib"
)

type VocabularyEntity struct {
	ankiClient   *anki.Client
	cardCreator  ai.AICardCreator
	ttsService   ai.TTS
	word         string
	deckName     string
	ttsAudioPath string
	cards        []ai.AnkiCard
}

func newVocabularyEntity(ankiClient *anki.Client, cardCreator ai.AICardCreator, ttsService ai.TTS, word string) *VocabularyEntity {
	v := &VocabularyEntity{
		ankiClient:  ankiClient,
		cardCreator: cardCreator,
		ttsService:  ttsService,
		word:        word,
	}

	return v
}

func (v *VocabularyEntity) Create() error {
	if err := v.chooseDeck(); err != nil {
		return fmt.Errorf("choose deck: %w", err)
	}
	if err := v.createAnkiCards(); err != nil {
		return fmt.Errorf("create anki cards: %w", err)
	}
	if err := v.createTTS(); err != nil {
		return fmt.Errorf("create tts: %w", err)
	}

	for _, c := range v.cards {
		data := map[string]interface{}{
			"Question":   c.Front,
			"Definition": c.Back,
			"Audio":      fmt.Sprintf("[sound:%s.mp3]", v.word),
		}

		note := anki.NewNoteBuilder(v.deckName, "VocabularyWithAudio", data).
			WithAudio(
				v.ttsAudioPath,
				fmt.Sprintf("%s.mp3", v.word),
				"Front",
			).Build()

		id, err := v.ankiClient.Notes.Add(note)
		if err != nil {
			return fmt.Errorf("add note: %w", err)
		}
		slog.Info(
			"note added",
			slog.String("front", c.Front),
			slog.String("back", c.Back),
			slog.String("id", fmt.Sprintf("%.f", id)),
		)
	}
	return nil
}

func (v *VocabularyEntity) createTTS() error {
	mp3, err := v.ttsService.GenerateMP3(v.word)
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

func (v *VocabularyEntity) createAnkiCards() error {
	cards, err := v.cardCreator.Create(
		v.deckName,
		"Create a vocabulary card (with parts of speech ONLY on front) for the word: "+v.word+".",
	)
	if err != nil {
		return fmt.Errorf("create anki cards: %w", err)
	}
	slog.Info("anki card created", slog.String("word", v.word), slog.Any("card", cards))

	v.cards = cards
	return nil
}

func (v *VocabularyEntity) chooseDeck() error {
	decks, err := v.getVocabDecks("Vocabulary")
	if err != nil {
		return fmt.Errorf("get vocabulary deck names: %w", err)
	}

	deckName, err := v.cardCreator.ChooseDeck(decks, fmt.Sprintf("Which vocabulary deck should I use for the word: %s", v.word))
	if err != nil {
		return fmt.Errorf("choose deck: %w", err)
	}
	slog.Info("deck chosen", slog.String("deck", deckName))

	v.deckName = deckName
	return nil
}

func (v *VocabularyEntity) getDecks() ([]string, error) {
	deckNames, err := v.ankiClient.DeckNames.GetNames()
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
