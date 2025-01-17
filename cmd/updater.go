package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/netr/haki/ai"
	"github.com/netr/haki/anki"
	"github.com/netr/haki/lib"
)

func NewUpdaterCommand(apiKey, outputDir string) *cli.Command {
	return &cli.Command{
		Name:      "updater",
		Usage:     "Update Vocabulary anki cards",
		ArgsUsage: "--debug",
		Flags: []cli.Flag{
			newDebugFlag(),
		},
		Action: actionFn(
			NewUpdaterAction(
				apiKey,
				"updater",
				[]string{"debug"},
				outputDir,
			)),
	}
}

type UpdaterAction struct {
	Action
	outputDir string
}

func NewUpdaterAction(apiKey, name string, flags []string, outputDir string) *UpdaterAction {
	return &UpdaterAction{
		Action: Action{
			flags:  flags,
			apiKey: apiKey,
			name:   name,
		},
		outputDir: outputDir,
	}
}

func (a UpdaterAction) Run(args ...interface{}) error {
	pluginName := args[0].(string)
	debug := "false"
	if len(args) >= 2 {
		debug = args[1].(string)
	}

	// skipSave := debug == "true"

	slog.Info("action",
		slog.String("action", pluginName),
		slog.String("debug", debug),
	)

	client := anki.NewClient(lib.GetEnv("ANKI_CONNECT_URL", "http://localhost:8765"))
	noteSvc := anki.NewNoteService(client)
	vocabUpdaterSvc := lib.NewVocabUpdaterService(client)
	notes, err := vocabUpdaterSvc.GetUpdatableNotes("deck:Vocabulary::*", "VocabularyWithAudio", []string{"Audio", "Picture"})
	if err != nil {
		return fmt.Errorf("get updatable notes: %w", err)
	}

	for _, n := range notes {
		query := ""
		switch n.ModelName {
		case "Basic":
			slog.Info(
				"updatable note found",
				slog.Int64("note_id", n.NoteID),
				slog.String("model_name", n.ModelName),
				slog.String("query", n.Fields["Front"].Value),
			)
			query = n.Fields["Front"].Value
		case "VocabularyWithAudio":
			slog.Info(
				"updatable note found",
				slog.Int64("note_id", n.NoteID),
				slog.String("model_name", n.ModelName),
				slog.String("query", n.Fields["Question"].Value),
			)
			query = n.Fields["Question"].Value
		default:
			continue
		}

		if err := runUpdater(a.apiKey, query, a.outputDir, true, true, false); err != nil {
			slog.Error("failed updating note",
				slog.Int64("note_id", n.NoteID),
				slog.String("model_name", n.ModelName),
				slog.String("query", query),
			)
			return err
		}

		// need this or else we get read errors from anki connect client
		time.Sleep(time.Second * 2)

		if err := noteSvc.DeleteNotes([]int64{n.NoteID}); err != nil {
			slog.Error("failed deleting note",
				slog.Int64("note_id", n.NoteID),
				slog.String("model_name", n.ModelName),
				slog.String("query", query),
			)
			return err
		}
	}

	return nil
}

func runUpdater(apiKey, query, outputDir string, generateTTS bool, generateImage bool, skipSave bool) error {
	timeStart := time.Now()

	// Load plugin configuration
	path := filepath.Join(outputDir, fmt.Sprintf("plugins/%s/plugin.xml", "vocab"))
	config, err := lib.NewPluginConfigFrom(outputDir, path)
	if err != nil {
		return fmt.Errorf("loading plugin config: %w", err)
	}

	// Create base services
	cardCreator, err := ai.NewOpenAICardCreator(apiKey)
	if err != nil {
		return fmt.Errorf("new openai api provider: %w", err)
	}

	vocabWord, err := cardCreator.ExtractVocabularyEntity(context.Background(), query)
	if err != nil {
		return fmt.Errorf("extract vocab entity: %w", err)
	}
	slog.Info("vocab word extracted", slog.String("word", vocabWord))

	// Initialize optional services based on plugin configuration
	var ttsService ai.TTS
	var imageGenService ai.ImageGen

	if config.Services.TTS && generateTTS {
		ttsService = ai.NewTTSService(apiKey)
	}
	if config.Services.ImageGen && generateImage {
		imageGenService = ai.NewImageGenService(apiKey)
	}

	plugin, err := lib.NewDerivedPlugin(config, cardCreator, ttsService, imageGenService)
	if err != nil {
		return fmt.Errorf("creating plugin: %w", err)
	}

	prompt, err := config.GetPromptContentFrom(outputDir)
	if err != nil {
		return fmt.Errorf("get prompt content: %w", err)
	}

	// Choose appropriate deck
	deckName, err := plugin.ChooseDeck(context.Background(), vocabWord)
	if err != nil {
		return fmt.Errorf("choosing deck: %w", err)
	}

	// Generate Anki cards
	cards, err := plugin.GenerateAnkiCards(context.Background(), prompt, vocabWord)
	if err != nil {
		return fmt.Errorf("generating cards: %w", err)
	}

	// Store cards if not in debug mode
	if !skipSave {
		if err := plugin.StoreAnkiCards(deckName, vocabWord, cards, true); err != nil {
			return fmt.Errorf("storing cards: %w", err)
		}
	}

	lib.PrintCards(cards, true)

	slog.Info("plugin",
		slog.String("query", query),
		slog.String("deck", deckName),
		slog.String("duration", time.Since(timeStart).String()),
	)
	return nil
}
