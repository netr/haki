package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/netr/haki/ai"
)

func NewVocabCommand(apiKey, outputDir string) *cli.Command {
	return &cli.Command{
		Name:      "vocab",
		Usage:     "Create a vocabulary Anki card using the specified word.",
		ArgsUsage: "--word <word> --service <service> --model <model> --debug",
		Flags: []cli.Flag{
			newWordFlag(),
			newServiceFlag(),
			newModelFlag(),
			newDebugFlag(),
		},
		Action: actionFn(
			NewVocabAction(
				apiKey,
				"vocab",
				outputDir,
				[]string{"word", "service", "model", "debug"},
			)),
	}
}

type VocabAction struct {
	flags     []string
	apiKey    string
	name      string
	outputDir string
}

func NewVocabAction(apiKey, name, outputDir string, flags []string) *VocabAction {
	return &VocabAction{
		flags:     flags,
		apiKey:    apiKey,
		name:      name,
		outputDir: outputDir,
	}
}

func (a VocabAction) Flags() []string {
	return a.flags
}

func (a VocabAction) Name() string {
	return a.name
}

func (a VocabAction) Run(args ...interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("action run: %w", ErrQueryRequired)
	}
	word := args[0].(string)
	if word == "" {
		return ErrWordFlagRequired
	}

	words := []string{word}
	if strings.Contains(word, ",") {
		words = strings.Split(word, ",")
		for i, w := range words {
			words[i] = strings.TrimSpace(w)
		}
	}

	for _, word := range words {
		if err := runVocab(a.apiKey, word, a.outputDir); err != nil {
			slog.Error("run", slog.String("action", "vocab"), slog.String("error", err.Error()))
			return err
		}
	}
	return nil
}

func runVocab(apiKey, word, outputDir string) error {
	cardCreator, err := ai.NewCardCreator(ai.OpenAI, apiKey)
	if err != nil {
		return fmt.Errorf("new openai api provider: %w", err)
	}
	ttsService := ai.NewTTSService(apiKey)
	imageGenService := ai.NewImageGenService(apiKey)
	vocabEntity := newVocabEntity(cardCreator, ttsService, imageGenService, outputDir)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := vocabEntity.CreateCards(ctx, word, generateAnkiCardPrompt(), false); err != nil {
		return fmt.Errorf("create vocab cards: %w", err)
	}

	return nil
}
