package cmd

import (
	"context"
	"fmt"
	"log/slog"
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
