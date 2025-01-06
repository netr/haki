package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/netr/haki/ai"
)

func NewTopicCommand(apiKey, outputDir string) *cli.Command {
	return &cli.Command{
		Name:      "topic",
		Usage:     "GenerateAnkiCards a topical Anki card using the specified topic.",
		ArgsUsage: "--topic <topic> --service <service> --model <model> --debug",
		Flags: []cli.Flag{
			newTopicFlag(),
			newServiceFlag(),
			newModelFlag(),
			newDebugFlag(),
		},
		Action: actionFn(
			NewTopicAction(
				apiKey,
				"topic",
				[]string{"topic", "service", "model", "debug"},
			)),
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

type TopicAction struct {
	Action
}

func NewTopicAction(apiKey, name string, flags []string) *TopicAction {
	return &TopicAction{
		Action{
			flags:  flags,
			apiKey: apiKey,
			name:   name,
		},
	}
}

func (a TopicAction) Run(args ...interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("action run: %w", ErrQueryRequired)
	}
	topic := args[0].(string)
	service := args[1].(string)
	model := args[2].(string)
	debug := args[3].(string)

	skipSave := false
	if debug == "true" {
		skipSave = true
	}

	slog.Info("action",
		slog.String("action", "topic"),
		slog.String("topic", topic),
		slog.String("service", service),
		slog.String("model", model),
		slog.String("debug", debug),
	)

	if err := runTopic(a.apiKey, topic, model, skipSave); err != nil {
		return err
	}
	return nil
}

// runTopic creates an anki client, card creator and builds the anki card.
// doesn't need to be part of the action topic struct because the problem terminates after finishing.
// if we make this a long running program, we should put this in the struct and hold references to the client/creator.
func runTopic(apiKey, word, model string, skipSave bool) error {
	cardCreator, err := ai.NewCardCreator(ai.OpenAI, apiKey, ai.OpenAIModelName(model))
	if err != nil {
		return fmt.Errorf("new openai card creator (%s): %w", model, err)
	}
	plugin := newTopicPlugin(cardCreator)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	deckName, err := plugin.ChooseDeck(ctx, word)
	if err != nil {
		return fmt.Errorf("plugin: %w", err)
	}

	cards, err := plugin.GenerateAnkiCards(ctx, word)
	if err != nil {
		return fmt.Errorf("plugin: %w", err)
	}

	if !skipSave {
		if err := plugin.StoreAnkiCards(deckName, cards); err != nil {
			return fmt.Errorf("plugin: %w", err)
		}
	}

	PrintCards(cards, true)
	return nil
}
