package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/netr/haki/ai"
)

func NewVocabCommand(apiKey, outputDir string) *cli.Command {
	return &cli.Command{
		Name:      "vocab",
		Usage:     "Create a vocabulary Anki card using the specified word.",
		ArgsUsage: "--words <word,word> --service <service> --model <model> --debug",
		Flags: []cli.Flag{
			newWordsFlag(),
			newServiceFlag(),
			newModelFlag(),
			newDebugFlag(),
		},
		Action: actionFn(
			NewVocabAction(
				apiKey,
				"vocab",
				outputDir,
				[]string{"words", "service", "model", "debug"},
			)),
	}
}

type VocabAction struct {
	Action
	outputDir string
}

func NewVocabAction(apiKey, name, outputDir string, flags []string) *VocabAction {
	return &VocabAction{
		Action: Action{
			flags:  flags,
			apiKey: apiKey,
			name:   name,
		},
		outputDir: outputDir,
	}
}

func (a VocabAction) Run(args ...interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("action run: %w", ErrQueryRequired)
	}
	words := args[0].(string)
	if words == "" {
		return ErrWordFlagRequired
	}

	for _, word := range a.splitWords(words) {
		if err := runVocab(a.apiKey, word, a.outputDir); err != nil {
			return err
		}
	}
	return nil
}

func (a VocabAction) splitWords(word string) []string {
	words := []string{word}
	if strings.Contains(word, ",") {
		words = strings.Split(word, ",")
		for i, w := range words {
			words[i] = strings.TrimSpace(w)
		}
	}
	return words
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
