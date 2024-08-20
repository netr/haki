package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/netr/haki/ai"
	"github.com/netr/haki/lib"
	"github.com/sashabaranov/go-openai"
	"github.com/urfave/cli/v2"
)

func NewTTSCommand(apiKey string) *cli.Command {
	return &cli.Command{
		Name:      "tts",
		Usage:     "Create a text-to-speech audio file for the specified word.",
		ArgsUsage: "--word <word> [--out <output file>]",
		Flags:     []cli.Flag{newWordFlag(), newOutFlag()},
		Action:    actionTTS(apiKey),
	}
}

func actionTTS(apiKey string) func(cCtx *cli.Context) error {
	return func(cCtx *cli.Context) error {
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

		if err := runTTS(apiKey, word, output); err != nil {
			slog.Error("run", slog.String("action", "tts"), slog.String("error", err.Error()))
			return err
		}
		return nil
	}
}

func runTTS(apiKey, word, output string) error {
	ttsService := ai.NewTTSService(apiKey)
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
