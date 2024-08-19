package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/netr/haki/ai"
	"github.com/netr/haki/anki"
	"github.com/netr/haki/lib"
	"github.com/sashabaranov/go-openai"
	"github.com/urfave/cli/v2"
)

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

func actionTTS(cCtx *cli.Context) error {
	word := cCtx.String("word")
	if word == "" {
		return fmt.Errorf("word is required --word <word>")
	}
	// optional output file
	output := cCtx.String("out")
	if output != "" {
		if err := lib.ValidateOutputPath(output); err != nil {
			return fmt.Errorf("validate output path: %w", err)
		}
	}

	if err := runTTS(word, output); err != nil {
		slog.Error("run", slog.String("action", "tts"), slog.String("error", err.Error()))
		return err
	}
	return nil
}

func runTTS(word string, output string) error {
	apiToken := os.Getenv("OPENAI_API_KEY")
	if apiToken == "" {
		return fmt.Errorf("OPENAI_API_KEY is not set")
	}

	ttsService := ai.NewTTSService(apiToken)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	bytes, err := ttsService.Generate(ctx, word, openai.VoiceAlloy, openai.SpeechResponseFormatMp3)
	if err != nil {
		return fmt.Errorf("generate mp3: %w", err)
	}

	if output == "" {
		output = fmt.Sprintf("data/%s.mp3", word)
	}
	if err = lib.SaveFile(output, bytes); err != nil {
		return fmt.Errorf("save file: %w", err)
	}

	slog.Info("tts created", slog.String("word", word), slog.String("output", output))
	return nil
}

func actionVocab(cCtx *cli.Context) error {
	word := cCtx.String("word")
	if word == "" {
		return fmt.Errorf("word is required --word <word>")
	}

	if err := runVocab(word); err != nil {
		slog.Error("run", slog.String("action", "vocab"), slog.String("error", err.Error()))
		return err
	}
	return nil
}

func runVocab(word string) error {
	apiToken := os.Getenv("OPENAI_API_KEY")
	if apiToken == "" {
		return fmt.Errorf("OPENAI_API_KEY is not set")
	}

	ankiClient := anki.NewClient(getEnv("ANKI_CONNECT_URL", "http://localhost:8765"))
	aiService, err := ai.NewAICardCreator(ai.OpenAI, apiToken)
	if err != nil {
		return fmt.Errorf("new openai api provider: %w", err)
	}
	ttsService := ai.NewTTSService(apiToken)

	vocabEntity := newVocabularyEntity(ankiClient, aiService, ttsService, word)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := vocabEntity.Create(ctx); err != nil {
		return fmt.Errorf("create vocab entity: %w", err)
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

	wordFlag := &cli.StringFlag{
		Name:     "word",
		Aliases:  []string{"w"},
		Value:    "",
		Required: true,
		Usage:    "word to create a card for",
	}

	outFlag := &cli.StringFlag{
		Name:    "out",
		Aliases: []string{"o"},
		Value:   "",
		Usage:   "output file",
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
				Name:      "vocab",
				Usage:     "Create a vocabulary Anki card using the specified word.",
				ArgsUsage: "--word <word>",
				Flags:     []cli.Flag{wordFlag},
				Action:    actionVocab,
			},
			{
				Name:      "tts",
				Usage:     "Create a text-to-speech audio file for the specified word.",
				ArgsUsage: "--word <word> [--out <output file>]",
				Flags:     []cli.Flag{wordFlag, outFlag},
				Action:    actionTTS,
			},
			{
				Name:      "cardtest",
				Usage:     "Test creating a card for the specified word.",
				ArgsUsage: "--word <word>",
				Flags:     []cli.Flag{wordFlag},
				Action:    actionCardTest,
				Aliases:   []string{"test"},
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
