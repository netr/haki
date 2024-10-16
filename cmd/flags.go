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

func newTopicFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:     "topic",
		Aliases:  []string{"t"},
		Value:    "",
		Required: true,
		Usage:    "topic to create a card for",
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
