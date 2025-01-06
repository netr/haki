package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/urfave/cli/v2"

	"github.com/netr/haki/ai"
	"github.com/netr/haki/lib"
)

func NewImageCommand(apiKey, outputDir string) *cli.Command {
	return &cli.Command{
		Name:      "image",
		Usage:     "Create an image for the specified text.",
		ArgsUsage: "--prompt <prompt>",
		Flags: []cli.Flag{
			newPromptFlag(),
			newDebugFlag(),
		},
		Action: actionFn(
			NewImageAction(
				apiKey,
				outputDir,
				[]string{"prompt", "debug"},
			)),
	}
}

type ImageAction struct {
	*Action
	OutputDir string
}

func NewImageAction(apiKey, outputDir string, flags []string) *ImageAction {
	return &ImageAction{
		Action:    NewAction(apiKey, "image", flags),
		OutputDir: outputDir,
	}
}

func (i ImageAction) Run(args ...interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("action run (%s): %w", i.Name(), ErrQueryRequired)
	}
	prompt := fmt.Sprintf("Please create an illustration for the word \"%s\" to help visually represent its meaning for my Anki card.", args[0].(string))
	outPath := fmt.Sprintf("%s/data/%s.webp", i.OutputDir, uuid.NewString())
	debug := args[1].(string)

	skipSave := false
	if debug == "true" {
		skipSave = true
	}

	fmt.Println("Creating image with prompt:", prompt)
	fmt.Println("Skip save?: ", skipSave)

	if err := runImage(i.apiKey, prompt, outPath, skipSave); err != nil {
		return fmt.Errorf("action run (%s): %w", i.Name(), err)
	}
	return nil
}

func runImage(apiKey, prompt, outPath string, skipSave bool) error {
	svc := ai.NewImageGenService(apiKey)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	data, err := svc.Generate(ctx, prompt)
	if err != nil {
		return fmt.Errorf("generate image: %w", err)
	}

	if !skipSave {
		if err = lib.SaveFile(outPath, data); err != nil {
			return fmt.Errorf("save tts file: %w", err)
		}
		slog.Info("saving image to file", slog.String("prompt", prompt), slog.Int("size", len(data)))
	} else {
		slog.Info("skipping save", slog.String("prompt", prompt), slog.Int("size", len(data)))
	}
	return nil
}

func newPromptFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:     "prompt",
		Aliases:  []string{"p"},
		Value:    "",
		Required: true,
		Usage:    "prompt to create image with",
	}
}
