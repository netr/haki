package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/netr/haki/ai"
	"github.com/netr/haki/lib"
	"github.com/sashabaranov/go-openai"
	"github.com/urfave/cli/v2"
)

func NewTTSCommand() *cli.Command {
	return &cli.Command{
		Name:      "tts",
		Usage:     "Create a text-to-speech audio file for the specified word.",
		ArgsUsage: "--word <word> [--out <output file>]",
		Flags:     []cli.Flag{newWordFlag(), newOutFlag()},
		Action:    actionTTS,
	}
}

func actionTTS(cCtx *cli.Context) error {
	word := cCtx.String("word")
	if word == "" {
		return fmt.Errorf("word is required --word <word>")
	}
	// optional output file
	output := cCtx.String("out")
	if output != "" {
		if err := lib.ValidateOutputPath(output); err != nil {
			return fmt.Errorf("validate output path: %w", err)
		}
	}

	if err := runTTS(word, output); err != nil {
		slog.Error("run", slog.String("action", "tts"), slog.String("error", err.Error()))
		return err
	}
	return nil
}

func runTTS(word string, output string) error {
	apiToken := os.Getenv("OPENAI_API_KEY")
	if apiToken == "" {
		return fmt.Errorf("OPENAI_API_KEY is not set")
	}

	ttsService := ai.NewTTSService(apiToken)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	bytes, err := ttsService.Generate(ctx, word, openai.VoiceAlloy, openai.SpeechResponseFormatMp3)
	if err != nil {
		return fmt.Errorf("generate mp3: %w", err)
	}

	if output == "" {
		output = fmt.Sprintf("data/%s.mp3", word)
	}
	if err = lib.SaveFile(output, bytes); err != nil {
		return fmt.Errorf("save file: %w", err)
	}

	slog.Info("tts created", slog.String("word", word), slog.String("output", output))
	return nil
}
