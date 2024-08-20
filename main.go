package main

import (
	"bufio"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/netr/haki/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	cfg, err := initConfig("haki.json")
	if err != nil {
		log.Fatalf("failed initializing config: %v", err)
	}

	if err := initLogger(cfg); err != nil {
		log.Fatalf("failed initializing logger: %v", err)
	}

	app := newApplication(cfg)
	if err := app.run(os.Args); err != nil {
		slog.Error("run app", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

type application struct {
	config *Config
	app    *cli.App
}

func newApplication(cfg *Config) *application {
	app := &application{
		config: cfg,
		app:    &cli.App{},
	}
	app.setupAppMetadata()
	app.registerCommands()
	return app
}

// registerCommands registers all the commands for the app.
func (a *application) registerCommands() *cli.App {
	a.app.Commands = []*cli.Command{
		cmd.NewTTSCommand(a.config.APIKeys.OpenAI),
		cmd.NewVocabCommand(a.config.APIKeys.OpenAI),
		cmd.NewCardTestCommand(a.config.APIKeys.OpenAI),
	}
	return a.app
}

// setupAppMetadata sets up the app metadata.
func (a *application) setupAppMetadata() *cli.App {
	a.app.Name = "haki"
	a.app.Version = "0.0.1"
	a.app.Usage = "haki is a tool to help you create anki cards using AI and AnkiConnect"
	a.app.Authors = []*cli.Author{
		{
			Name:  "Corey Jackson (netr)",
			Email: "programmatical@gmail.com",
		},
	}
	a.app.Compiled = time.Now()
	a.app.EnableBashCompletion = true
	a.app.Before = a.beforeAppWithConfig()
	return a.app
}

// run runs the application.
func (a *application) run(args []string) error {
	return a.app.Run(args)
}

func (a *application) beforeAppWithConfig() cli.BeforeFunc {
	return func(cCtx *cli.Context) error {
		if a.config.APIKeys.OpenAI == "" {
			fmt.Printf("OpenAI API Key is not set.\nHaki needs the OpenAI API to generate cards and automatically place them in respective decks.\nIf you don't have an API key, you can learn how to get one here: https://platform.openai.com/docs/api-reference/introduction\n\n")
			apiKey, err := askUserFor("Please enter your OpenAI API Key: ")
			if err != nil {
				log.Fatalf("failed asking user for OpenAI API Key: %v", err)
			}
			a.config.APIKeys.OpenAI = strings.TrimSpace(apiKey)
			if err := a.config.Save(); err != nil {
				log.Fatalf("failed saving config: %v", err)
			}
		}

		// TODO: Add Anthropic API Key check here
		// TODO: Add AnkiConnect Model check here
		return nil
	}
}

func askUserFor(input string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(input)
	output, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed reading input: %w", err)
	}
	return output, nil
}
