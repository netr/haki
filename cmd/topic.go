package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/netr/haki/ai"
	"github.com/netr/haki/anki"
	"github.com/netr/haki/lib"
)

func NewTopicCommand(apiKey, outputDir string) *cli.Command {
	return &cli.Command{
		Name:      "topic",
		Usage:     "Create a topical Anki card using the specified topic.",
		ArgsUsage: "--topic <topic>",
		Flags:     []cli.Flag{newTopicFlag()},
		Action:    actionTopic(apiKey),
	}
}

func actionTopic(apiKey string) func(cCtx *cli.Context) error {
	return func(cCtx *cli.Context) error {
		topic := cCtx.String("topic")
		if topic == "" {
			return ErrTopicRequired
		}

		if err := runTopic(apiKey, topic); err != nil {
			slog.Error("run", slog.String("action", "topic"), slog.String("error", err.Error()))
			return err
		}
		return nil
	}
}

func runTopic(apiKey, word string) error {
	ankiClient := anki.NewClient(lib.GetEnv("ANKI_CONNECT_URL", "http://localhost:8765"))
	cardCreator, err := ai.NewAICardCreator(ai.OpenAI, apiKey)
	if err != nil {
		return fmt.Errorf("new openai api provider: %w", err)
	}
	topicEntity := newTopicEntity(ankiClient, cardCreator)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := topicEntity.Create(ctx, word); err != nil {
		return fmt.Errorf("create vocab entity: %w", err)
	}

	return nil
}

var (
	ErrTopicRequired = fmt.Errorf("topic is required")
)

type TopicEntity struct {
	ankiClient  anki.AnkiClienter
	cardCreator ai.AICardCreator
	topic       string
	deckName    string
	cards       []ai.AnkiCard
}

func newTopicEntity(ankiClient anki.AnkiClienter, cardCreator ai.AICardCreator) *TopicEntity {
	t := &TopicEntity{
		ankiClient:  ankiClient,
		cardCreator: cardCreator,
	}

	return t
}

func (t *TopicEntity) Create(ctx context.Context, topic string) error {
	slog.Info("creating topic card", slog.String("topic", topic))

	if topic == "" {
		return ErrTopicRequired
	}
	t.topic = topic

	if err := t.chooseDeck(ctx); err != nil {
		return fmt.Errorf("choose deck: %w", err)
	}
	slog.Info("deck name chosen", slog.String("deck", t.deckName))

	if err := t.createAnkiCards(ctx); err != nil {
		return fmt.Errorf("create anki cards: %w", err)
	}
	slog.Info("anki card(s) created", slog.Int("count", len(t.cards)))

	// TODO: FIXME: make sure this model is created before using the app. Preferably something like "Haki::VocabularyWithAudio"
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

	fmt.Println("")
	for _, c := range t.cards {
		fmt.Print(colors.BeautifyCard(c))
	}
	return nil
}

func (t *TopicEntity) createAnkiCards(ctx context.Context) error {
	cards, err := t.cardCreator.Create(
		ctx,
		t.deckName,
		"Create a concise and informative card for the topic: "+t.topic+".",
	)
	if err != nil {
		return fmt.Errorf("create anki cards: %w", err)
	}

	t.cards = cards
	return nil
}

func (t *TopicEntity) chooseDeck(ctx context.Context) error {
	decks, err := t.getTopicDecks("Haki")
	if err != nil {
		return fmt.Errorf("get vocabulary deck names: %w", err)
	}

	deckName, err := t.cardCreator.ChooseDeck(ctx, decks, fmt.Sprintf("Which deck should I use for the topic: %s", t.topic))
	if err != nil {
		return fmt.Errorf("choose deck: %w", err)
	}

	t.deckName = deckName
	return nil
}

func (t *TopicEntity) getDecks() ([]string, error) {
	deckNames, err := t.ankiClient.DeckNames().GetNames()
	if err != nil {
		return nil, fmt.Errorf("get deck names: %w", err)
	}
	return anki.RemoveParentDecks(deckNames), nil
}

func (t *TopicEntity) getTopicDecks(filter ...string) ([]string, error) {
	deckNames, err := t.getDecks()
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
