package main

import (
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/netr/haki/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("failed loading .env file")
	}

	if err := initLogger(newLoggerConfig("haki")); err != nil {
		log.Fatalf("failed initializing logger: %v", err)
	}

	app := setupApp(registerCommands(cli.NewApp()))
	if err := app.Run(os.Args); err != nil {
		slog.Error("run app", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// registerCommands registers all the commands for the app.
func registerCommands(app *cli.App) *cli.App {
	app.Commands = []*cli.Command{
		cmd.NewTTSCommand(),
		cmd.NewVocabCommand(),
		cmd.NewCardTestCommand(),
	}
	return app
}

// setupApp sets up the app metadata.
func setupApp(app *cli.App) *cli.App {
	app.Name = "haki"
	app.Version = "0.0.1"
	app.Usage = "haki is a tool to help you create anki cards using AI and AnkiConnect"
	app.Authors = []*cli.Author{
		{
			Name:  "Corey Jackson (netr)",
			Email: "programmatical@gmail.com",
		},
	}
	app.Compiled = time.Now()
	app.EnableBashCompletion = true
	return app
}
