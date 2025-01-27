package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/netr/haki/ai"
	"github.com/netr/haki/lib"
)

func NewPluginCommand(apiKey, outputDir string) *cli.Command {
	return &cli.Command{
		Name:      "plugin",
		Usage:     "Generate Anki cards using a specified plugin",
		ArgsUsage: "--plugin <plugin_name> --query <query> --model <model> --debug",
		Flags: []cli.Flag{
			newPluginFlag(),
			newQueryFlag(),
			newModelFlag(),
			newDebugFlag(),
		},
		Action: actionFn(
			NewPluginAction(
				apiKey,
				"plugin",
				[]string{"plugin", "query", "model", "debug"},
				outputDir,
			)),
	}
}

type PluginAction struct {
	Action
	outputDir string
}

func NewPluginAction(apiKey, name string, flags []string, outputDir string) *PluginAction {
	return &PluginAction{
		Action: Action{
			flags:  flags,
			apiKey: apiKey,
			name:   name,
		},
		outputDir: outputDir,
	}
}

func (a PluginAction) Run(args ...interface{}) error {
	if len(args) < 2 {
		return fmt.Errorf("action run: %w", lib.ErrQueryRequired)
	}

	pluginName := args[0].(string)
	query := args[1].(string)
	model := args[2].(string)
	debug := args[3].(string)

	skipSave := debug == "true"

	slog.Info("action",
		slog.String("action", "plugin"),
		slog.String("plugin", pluginName),
		slog.String("query", query),
		slog.String("model", model),
		slog.String("debug", debug),
	)

	// Load plugin configuration
	path := filepath.Join(a.outputDir, fmt.Sprintf("plugins/%s/plugin.xml", pluginName))
	config, err := lib.NewPluginConfigFrom(a.outputDir, path)
	if err != nil {
		return fmt.Errorf("loading plugin config: %w", err)
	}

	queries := []string{query}
	if config.Generation.Mode == "single" {
		queries = lib.SplitQuery(query)
	}
	for _, query := range queries {
		if err := runPlugin(config, a.apiKey, query, model, skipSave); err != nil {
			slog.Error("failed creating card",
				slog.String("error", err.Error()),
				slog.String("plugin", pluginName),
				slog.String("query", query),
				slog.String("model", model),
			)
		}
	}
	return nil
}

func runPlugin(config *lib.PluginConfig, apiKey, query, model string, skipSave bool) error {
	timeStart := time.Now()

	// Create base services
	cardCreator, err := ai.NewOpenAICardCreator(apiKey)
	if err != nil {
		return fmt.Errorf("new openai api provider: %w", err)
	}

	// Initialize optional services based on plugin configuration
	var ttsService ai.TTS
	var imageGenService ai.ImageGen

	if config.Services.TTS {
		ttsService = ai.NewTTSService(apiKey)
	}
	if config.Services.ImageGen {
		imageGenService = ai.NewImageGenService(apiKey)
	}

	plugin, err := lib.NewDerivedPlugin(config, cardCreator, ttsService, imageGenService)
	if err != nil {
		return fmt.Errorf("creating plugin: %w", err)
	}

	prompt, err := config.GetPromptContentFrom(config.OutputDir)
	if err != nil {
		return fmt.Errorf("get prompt content: %w", err)
	}

	// Choose appropriate deck
	deckName, err := plugin.ChooseDeck(context.Background(), query)
	if err != nil {
		return fmt.Errorf("choosing deck: %w", err)
	}

	// Generate Anki cards
	cards, err := plugin.GenerateAnkiCards(context.Background(), prompt, query)
	if err != nil {
		return fmt.Errorf("generating cards: %w", err)
	}

	// Store cards if not in debug mode
	if !skipSave {
		if err := plugin.StoreAnkiCards(deckName, query, cards, false); err != nil {
			return fmt.Errorf("storing cards: %w", err)
		}
	}

	lib.PrintCards(cards, true)

	slog.Info("plugin",
		slog.String("plugin", config.Identifier),
		slog.String("query", query),
		slog.String("model", model),
		slog.String("deck", deckName),
		slog.String("duration", time.Since(timeStart).String()),
	)
	return nil
}
