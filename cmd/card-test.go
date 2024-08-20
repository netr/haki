package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/netr/haki/ai"
	"github.com/urfave/cli/v2"
)

func NewCardTestCommand(apiKey string) *cli.Command {
	return &cli.Command{
		Name:      "cardtest",
		Usage:     "Test creating a card for the specified word.",
		ArgsUsage: "--word <word>",
		Flags:     []cli.Flag{newWordFlag()},
		Action:    actionCardTest(apiKey),
		Aliases:   []string{"test"},
	}
}

func actionCardTest(apiKey string) func(cCtx *cli.Context) error {
	return func(cCtx *cli.Context) error {
		word := cCtx.String("word")
		if word == "" {
			return fmt.Errorf("word is required --word <word>")
		}

		if err := runCardTest(apiKey, word); err != nil {
			slog.Error("run", slog.String("action", "card_test"), slog.String("error", err.Error()))
			return err
		}
		return nil
	}
}

func runCardTest(apiKey, word string) error {
	cardCreator, err := ai.NewOpenAICardCreator(apiKey)
	if err != nil {
		return fmt.Errorf("new openai card creator: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ans, err := cardCreator.Create(ctx, "test", word)
	if err != nil {
		return fmt.Errorf("create card: %w", err)
	}

	for _, c := range ans {
		fmt.Print(colors.BeautifyCard(c))
	}
	return nil
}
