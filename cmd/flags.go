package cmd

import "github.com/urfave/cli/v2"

func newWordFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:     "word",
		Aliases:  []string{"w"},
		Value:    "",
		Required: true,
		Usage:    "word to create a card for",
	}
}

func newOutFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:    "out",
		Aliases: []string{"o"},
		Value:   "",
		Usage:   "output file",
	}
}

func newServiceFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:    "service",
		Aliases: []string{"svc"},
		Value:   "openai",
		Usage:   "ai api service",
	}
}

func newModelFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:    "model",
		Aliases: []string{"m"},
		Value:   "gpt-4o-2024-11-20",
		Usage:   "ai model",
	}
}
