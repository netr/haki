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
