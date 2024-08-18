package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/netr/haki/ai"
	"github.com/urfave/cli/v2"
)

func runVocab(word string) error {
	apiToken := os.Getenv("OPENAI_API_KEY")
	if apiToken == "" {
		return fmt.Errorf("OPENAI_API_KEY is not set")
	}

	ttsService := ai.NewTTSService(apiToken)
	res, err := ttsService.GenerateMP3(word)
	if err != nil {
		return fmt.Errorf("generate mp3: %w", err)
	}

	var bytes []byte
	bytes, err = io.ReadAll(res.ReadCloser)
	if err != nil {
		return fmt.Errorf("read tts response: %w", err)
	}

	// save mp3 to file
	f, err := os.Create(fmt.Sprintf("%s.mp3", word))
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	_, err = f.Write(bytes)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	slog.Info("tts created", slog.String("word", word))
	return nil
}

func vocabAction(cCtx *cli.Context) error {
	fmt.Println(cCtx.Command.VisibleFlags())
	word := cCtx.String("word")
	if word == "" {
		return fmt.Errorf("word is required --word <word>")
	}

	if err := runVocab(word); err != nil {
		slog.Error("run", slog.String("error", err.Error()))
		return err
	}
	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("failed loading .env file")
	}

	if err := initLogger(newLoggerConfig("haki")); err != nil {
		log.Fatalf("failed initializing logger: %v", err)
	}

	app := &cli.App{
		Name:  "haki",
		Usage: "haki is a tool to help you create anki cards using AI and AnkiConnect",
		Authors: []*cli.Author{
			{
				Name:  "Corey Jackson (netr)",
				Email: "programmatical@gmail.com",
			},
		},
		Version:  "0.0.1",
		Compiled: time.Now(),
		Commands: []*cli.Command{
			{
				Name:  "vocab",
				Usage: "create a vocabulary anki card",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "word",
						Aliases: []string{"w"},
						Value:   "",
						Usage:   "word to create a card for",
					},
				},
				Action: vocabAction,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("run app", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	ivalue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return ivalue
}
