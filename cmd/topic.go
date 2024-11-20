package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
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
		Action:    actionFn(NewActionTopic(apiKey, "topic", []string{"topic"})),
		// Action: actionFn("topic", runTopic(apiKey, topic))
	}
}

func newTopicFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:     "topic",
		Aliases:  []string{"t"},
		Value:    "",
		Required: true,
		Usage:    "topic to create a card for",
	}
}

type ActionTopic struct {
	flags  []string
	apiKey string
	name   string
}

func NewActionTopic(apiKey, name string, flags []string) *ActionTopic {
	return &ActionTopic{
		flags:  flags,
		apiKey: apiKey,
		name:   name,
	}
}

func (a ActionTopic) Flags() []string {
	return a.flags
}

func (a ActionTopic) Name() string {
	return a.name
}

func (a ActionTopic) Run(args ...interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("action run: %w", ErrQueryRequired)
	}
	topic := args[0].(string)

	if err := runTopic(a.apiKey, topic); err != nil {
		return fmt.Errorf("action run: %w", err)
	}
	return nil
}

// runTopic creates an anki client, card creator and builds the anki card.
// doesn't need to be part of the action topic struct because the problem terminates after finishing.
// if we make this a long running program, we should put this in the struct and hold references to the client/creator.
func runTopic(apiKey, word string) error {
	cardCreator, err := ai.NewAICardCreator(ai.OpenAI, apiKey)
	if err != nil {
		return fmt.Errorf("new openai api provider: %w", err)
	}
	topicEntity := newTopicEntity(cardCreator)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := topicEntity.CreateCards(ctx, word, generateAnkiCardPrompt()); err != nil {
		return fmt.Errorf("create topic entity: %w", err)
	}

	return nil
}

var (
	ErrQueryRequired = fmt.Errorf("query is required")
)

type ErrFlagValueMissing struct {
	Flag string
}

func (e *ErrFlagValueMissing) Error() string {
	return fmt.Sprintf("flag '%s' is missing data", e.Flag)
}

type BaseEntity struct {
	ankiClient  anki.AnkiClienter
	cardCreator ai.AICardCreator
	deckName    string
	cards       []ai.AnkiCard
}

func NewBaseEntity(cardCreator ai.AICardCreator) *BaseEntity {
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
func (t *BaseEntity) addCardsToAnki() error {
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

func (t *BaseEntity) chooseDeck(ctx context.Context, query string, decks []string) error {
	deckName, err := t.cardCreator.ChooseDeck(ctx, decks, fmt.Sprintf("Which deck should I use for the topic: %s", query))
	if err != nil {
		return fmt.Errorf("choose deck: %w", err)
	}

	t.deckName = deckName
	return nil
}

type CardCreatorEntity interface {
	CreateCards(ctx context.Context, query string, prompt string) error
}

type TopicEntity struct {
	*BaseEntity
	topic string
}

func newTopicEntity(cardCreator ai.AICardCreator) CardCreatorEntity {
	t := &TopicEntity{
		BaseEntity: NewBaseEntity(cardCreator),
	}
	return t
}

func (t *TopicEntity) CreateCards(ctx context.Context, query string, prompt string) error {
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
	if err := t.chooseDeck(ctx, t.topic, decks); err != nil {
		return fmt.Errorf("choose deck: %w", err)
	}
	slog.Info("deck name chosen", slog.String("deck", t.deckName))

	if !slices.Contains(decks, t.deckName) {
		if err := t.ankiClient.DeckNames().Create(t.deckName); err != nil {
			return fmt.Errorf("create deck (%s): %w", t.deckName, err)
		}
	}

	if err := t.createAnkiCards(ctx, t.topic, prompt); err != nil {
		return fmt.Errorf("create anki cards: %w", err)
	}
	slog.Info("anki card(s) created", slog.Int("count", len(t.cards)))

	if err := t.addCardsToAnki(); err != nil {
		return fmt.Errorf("create anki cards: %w", err)
	}

	fmt.Println("")
	for _, c := range t.cards {
		fmt.Print(colors.BeautifyCard(c))
	}
	return nil
}
