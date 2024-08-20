package main

import (
	"bufio"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/netr/haki/cmd"
	"github.com/netr/haki/lib"
)

func main() {
	cfg := mustSetup()
	app := newApplication(cfg)
	if err := app.run(os.Args); err != nil {
		slog.Error("run app", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func mustSetup() *Config {
	hakiDir, err := getHakiDirPath()
	if err != nil {
		log.Fatalf("failed getting haki directory path: %v", err)
	}
	cfgFile, err := getConfigPath(hakiDir)
	if err != nil {
		log.Fatalf("failed getting config path: %v", err)
	}
	cfg, err := initConfig(cfgFile)
	if err != nil {
		log.Fatalf("failed initializing config: %v", err)
	}
	if err := initLogger(cfg); err != nil {
		log.Fatalf("failed initializing logger: %v", err)
	}

	cfg.hakiDir = hakiDir
	return cfg
}

type application struct {
	config  *Config
	app     *cli.App
	hakiDir string
}

func newApplication(cfg *Config) *application {
	app := &application{
		config:  cfg,
		app:     &cli.App{},
		hakiDir: filepath.Dir(cfg.fileName),
	}
	app.setupAppMetadata()
	app.registerCommands()
	return app
}

// registerCommands registers all the commands for the app.
func (a *application) registerCommands() *cli.App {
	a.app.Commands = []*cli.Command{
		cmd.NewTTSCommand(a.config.APIKeys.OpenAI, a.config.hakiDir),
		cmd.NewVocabCommand(a.config.APIKeys.OpenAI, a.config.hakiDir),
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
	return func(_ *cli.Context) error {
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
		return "", fmt.Errorf("reading input: %w", err)
	}
	return output, nil
}

// getConfigPath returns the path to the configuration file 'haki.json' based on the OS.
func getConfigPath(hakiDir string) (string, error) {
	configPath := filepath.Join(hakiDir, "config.json")
	return configPath, nil
}

// ErrUnsupportedPlatform is returned when the platform is not supported.
var ErrUnsupportedPlatform = fmt.Errorf("unsupported platform")

// getHakiDirPath returns the path to the haki directory based on the OS.
// If the config.json file exists in the current directory, it will return the current directory. This allows for better development and testing.
func getHakiDirPath() (string, error) {
	if lib.FileExists("./config.json") {
		return ".", nil
	}

	var basePath string
	switch runtime.GOOS {
	case "windows":
		basePath = os.Getenv("LOCALAPPDATA")
	case "darwin", "linux":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if runtime.GOOS == "darwin" {
			basePath = filepath.Join(home, "Library", "Application Support")
		} else {
			basePath = home // For Linux, use the home directory
		}
	default:
		return "", ErrUnsupportedPlatform
	}

	hakiDir := filepath.Join(basePath, ".haki")
	if err := os.MkdirAll(hakiDir, 0755); err != nil {
		return "", err
	}
	dataDir := filepath.Join(hakiDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", err
	}
	return hakiDir, nil
}
