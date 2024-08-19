package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/netr/haki/ai"
	"github.com/urfave/cli/v2"
)

func NewCardTestCommand() *cli.Command {
	return &cli.Command{
		Name:      "cardtest",
		Usage:     "Test creating a card for the specified word.",
		ArgsUsage: "--word <word>",
		Flags:     []cli.Flag{newWordFlag()},
		Action:    actionCardTest,
		Aliases:   []string{"test"},
	}
}

func actionCardTest(cCtx *cli.Context) error {
	word := cCtx.String("word")
	if word == "" {
		return fmt.Errorf("word is required --word <word>")
	}

	if err := runCardTest(word); err != nil {
		slog.Error("run", slog.String("action", "card_test"), slog.String("error", err.Error()))
		return err
	}
	return nil
}

func runCardTest(word string) error {
	apiToken := os.Getenv("OPENAI_API_KEY")
	if apiToken == "" {
		return fmt.Errorf("OPENAI_API_KEY is not set")
	}

	oa, err := ai.NewOpenAICardCreator(apiToken)
	if err != nil {
		return fmt.Errorf("new openai card creator: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ans, err := oa.Create(ctx, "test", word)
	if err != nil {
		return fmt.Errorf("create card: %w", err)
	}

	fmt.Println(ans)
	return nil
}
