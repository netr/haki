package cmd

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/urfave/cli/v2"

	"github.com/netr/haki/lib"

	"github.com/netr/haki/ai"
)

const TrueString = "true"

func NewAWSCommand(apiKey, outputDir string) *cli.Command {
	return &cli.Command{
		Name:      "aws",
		Usage:     "GenerateAnkiCards a topical Anki card using the specified topic.",
		ArgsUsage: "--topic <topic> --service <service> --model <model> --debug",
		Flags: []cli.Flag{
			newTopicFlag(),
			newServiceFlag(),
			newModelFlag(),
			newDebugFlag(),
		},
		Action: actionFn(
			NewAWSAction(
				apiKey,
				"aws",
				[]string{"topic", "service", "model", "debug"},
			)),
	}
}

type AWSAction struct {
	Action
}

func NewAWSAction(apiKey, name string, flags []string) *AWSAction {
	return &AWSAction{
		Action{
			flags:  flags,
			apiKey: apiKey,
			name:   name,
		},
	}
}

func (a AWSAction) Run(args ...interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("action run: %w", lib.ErrQueryRequired)
	}
	topic := args[0].(string)
	service := args[1].(string)
	model := args[2].(string)
	debug := args[3].(string)

	skipSave := false
	if debug == TrueString {
		skipSave = true
	}

	slog.Info("action",
		slog.String("action", "aws"),
		slog.String("topic", topic),
		slog.String("service", service),
		slog.String("model", model),
		slog.String("debug", debug),
	)

	if err := runAWS(a.apiKey, topic, model, skipSave); err != nil {
		return err
	}
	return nil
}

// runTopic creates an anki client, card creator and builds the anki card.
// doesn't need to be part of the action topic struct because the problem terminates after finishing.
// if we make this a long running program, we should put this in the struct and hold references to the client/creator.
func runAWS(apiKey, query, model string, skipSave bool) error {
	config, err := lib.NewPluginConfigFrom("plugins/aws/plugin.xml")
	if err != nil {
		log.Fatal(err)
	}

	// Create services
	cardCreator, err := ai.NewOpenAICardCreator(apiKey)
	if err != nil {
		return fmt.Errorf("new openai api provider: %w", err)
	}
	ttsService := ai.NewTTSService(apiKey)
	imageGenService := ai.NewImageGenService(apiKey)

	// Create plugin
	plugin, err := lib.NewDerivedPlugin(config, cardCreator, ttsService, imageGenService)
	if err != nil {
		log.Fatal(err)
	}

	deckName, err := plugin.ChooseDeck(context.Background(), query)
	if err != nil {
		return fmt.Errorf("run vocab: %w", err)
	}

	// Use the plugin
	cards, err := plugin.GenerateAnkiCards(context.Background(), query)
	if err != nil {
		return fmt.Errorf("run vocab: %w", err)
	}

	if err := plugin.StoreAnkiCards(deckName, query, cards); err != nil {
		return fmt.Errorf("run vocab: %w", err)
	}

	lib.PrintCards(cards, true)
	return nil
}
