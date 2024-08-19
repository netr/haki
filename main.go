package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/netr/haki/ai"
	"github.com/netr/haki/anki"
	"github.com/sashabaranov/go-openai"
	"github.com/urfave/cli/v2"
)

func ttsAction(cCtx *cli.Context) error {
	word := cCtx.String("word")
	if word == "" {
		return fmt.Errorf("word is required --word <word>")
	}

	if err := runTTS(word); err != nil {
		slog.Error("run", slog.String("error", err.Error()))
		return err
	}
	return nil
}

func runTTS(word string) error {
	apiToken := os.Getenv("OPENAI_API_KEY")
	if apiToken == "" {
		return fmt.Errorf("OPENAI_API_KEY is not set")
	}

	ttsService := ai.NewTTSService(apiToken)
	res, err := ttsService.Generate(word, openai.VoiceAlloy, openai.SpeechResponseFormatWav)
	if err != nil {
		return fmt.Errorf("generate wav: %w", err)
	}

	var bytes []byte
	bytes, err = io.ReadAll(res.ReadCloser)
	if err != nil {
		return fmt.Errorf("read tts response: %w", err)
	}

	// save wav to file
	f, err := os.Create(fmt.Sprintf("data/%s.wav", word))
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

func runVocab(word string) error {
	apiToken := os.Getenv("OPENAI_API_KEY")
	if apiToken == "" {
		return fmt.Errorf("OPENAI_API_KEY is not set")
	}

	ankiClient := anki.NewClient(getEnv("ANKI_CONNECT_URL", "http://localhost:8765"))
	aiService, err := ai.NewAPIProvider(ai.OpenAI, apiToken)
	if err != nil {
		return fmt.Errorf("new openai api provider: %w", err)
	}

	deckNames, err := ankiClient.DeckNames.GetNames()
	if err != nil {
		return fmt.Errorf("get deck names: %w", err)
	}
	// only use deck names that have Vocabulary in them
	var decks []string
	for _, d := range anki.RemoveParentDecks(deckNames) {
		if strings.Contains(d, "Vocabulary") {
			decks = append(decks, d)
		}
	}

	deckName, err := aiService.Action().ChooseDeck(decks, fmt.Sprintf("Which vocabulary deck should I use for the word: %s", word))
	if err != nil {
		return fmt.Errorf("choose deck: %w", err)
	}
	slog.Info("deck chosen", slog.String("deck", deckName))

	card, err := aiService.Action().CreateAnkiCards(deckName, "Create a vocabulary card (with parts of speech ONLY on front) for the word: "+word+".")
	if err != nil {
		return fmt.Errorf("create anki cards: %w", err)
	}
	slog.Info("anki card created", slog.String("word", word), slog.Any("card", card))

	ttsService := ai.NewTTSService(apiToken)
	res, err := ttsService.Generate(word, openai.VoiceAlloy, openai.SpeechResponseFormatMp3)
	if err != nil {
		return fmt.Errorf("generate mp3: %w", err)
	}

	var bytes []byte
	bytes, err = io.ReadAll(res.ReadCloser)
	if err != nil {
		return fmt.Errorf("read tts response: %w", err)
	}

	// save mp3 to file
	f, err := os.Create(fmt.Sprintf("data/%s.mp3", word))
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	_, err = f.Write(bytes)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	slog.Info("tts created", slog.String("word", word))

	for _, c := range card {
		data := map[string]interface{}{
			"Question":   c.Front,
			"Definition": c.Back,
			"Audio":      fmt.Sprintf("[sound:%s.mp3]", word),
		}

		// how to get the absolute path of the mp3 file
		mp3Path, err := filepath.Abs(fmt.Sprintf("data/%s.mp3", word))
		if err != nil {
			return fmt.Errorf("get absolute path: %w", err)
		}

		note := anki.
			NewNoteBuilder("Vocabulary::English", "VocabularyWithAudio", data).
			WithAudio(
				mp3Path,
				fmt.Sprintf("%s.mp3", word),
				"Front",
			).Build()

		id, err := ankiClient.Notes.Add(note)
		if err != nil {
			return fmt.Errorf("add note: %w", err)
		}
		slog.Info(
			"note added",
			slog.String("word", word),
			slog.String("id", fmt.Sprintf("%.f", id)),
		)
	}
	return nil
}

func vocabAction(cCtx *cli.Context) error {
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
			{
				Name:  "tts",
				Usage: "create a tts mp3 for a word",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "word",
						Aliases: []string{"w"},
						Value:   "",
						Usage:   "word to create a card for",
					},
				},
				Action: ttsAction,
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
