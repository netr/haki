package cmd

import (
	"github.com/urfave/cli/v2"
)

func newDebugFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:    "debug",
		Aliases: []string{"d"},
		Value:   false,
		Usage:   "debug mode",
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

func newPluginFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:     "plugin",
		Aliases:  []string{"p"},
		Usage:    "name of the plugin to use",
		Required: true,
	}
}

func newQueryFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:     "query",
		Aliases:  []string{"q"},
		Usage:    "comma-separated list of search queries QUERY[,QUERY...]",
		Required: true,
	}
}
