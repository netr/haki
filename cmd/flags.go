package cmd

import (
	"github.com/urfave/cli/v2"
)

func newWordsFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:     "words",
		Aliases:  []string{"w"},
		Value:    "",
		Required: true,
		Usage:    "words to create a card for (comma separated)",
	}
}

func newOutFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:     "out",
		Aliases:  []string{"o"},
		Value:    "",
		Required: false,
		Usage:    "output file",
	}
}

func newDebugFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:    "debug",
		Aliases: []string{"d"},
		Value:   false,
		Usage:   "debug mode",
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
