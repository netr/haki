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

func runTTS(apiKey, text, outPath string) error {
	ttsService := ai.NewTTSService(apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	bytes, err := ttsService.Generate(ctx, text, openai.VoiceAlloy, openai.SpeechResponseFormatMp3)
	if err != nil {
		return fmt.Errorf("generate tts: %w", err)
	}

	if outPath == "" {
		outPath = fmt.Sprintf("data/%s.mp3", text)
	}
	if err = lib.SaveFile(outPath, bytes); err != nil {
		return fmt.Errorf("save tts file: %w", err)
	}

	slog.Info("tts created", slog.String("word", text), slog.String("output", outPath))
	return nil
}
