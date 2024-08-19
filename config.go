package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

var (
	ErrConfigNotFound = fmt.Errorf("config not found")
)

type Config struct {
	Logger   *ConfigLogger  `json:"logger"`
	APIKeys  *ConfigApiKeys `json:"api_keys"`
	fileName string
}

func (c *Config) Save() error {
	return saveConfig(c.fileName, c)
}

type ConfigApiKeys struct {
	OpenAI    string `json:"openai"`
	Anthropic string `json:"anthropic"`
}

type ConfigLogger struct {
	Level     string     `json:"level"`
	Format    string     `json:"format"`
	Output    string     `json:"output"`
	File      LoggerFile `json:"file"`
	Command   string     `json:"command"`
	DebugMode bool       `json:"debug_mode"`
}

type LoggerFile struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
}

// initConfig initializes the config file.
// If the config file does not exist, it will create a default config file.
// We don't need to worry about API keys here, because it will be caught in the command?
// We need to allow the user to paste in their API keys....
func initConfig(path string) (*Config, error) {
	cfg, err := readConfig(path)
	if err != nil {
		if errors.Is(err, ErrConfigNotFound) {
			return createDefaultConfig(path)
		}
		return nil, err
	}
	return cfg, nil
}

func createDefaultConfig(path string) (*Config, error) {
	cfg := &Config{
		Logger: createDefaultLoggerConfig(),
		APIKeys: &ConfigApiKeys{
			OpenAI:    "",
			Anthropic: "",
		},
		fileName: path,
	}

	err := saveConfig(path, cfg)
	if err != nil {
		return nil, fmt.Errorf("save config: %w", err)
	}
	return cfg, nil
}

func saveConfig(path string, cfg *Config) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cfg); err != nil {
		return err
	}
	return nil
}

func readConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, ErrConfigNotFound
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}

	config.fileName = path
	return &config, nil
}

func createDefaultLoggerConfig() *ConfigLogger {
	return &ConfigLogger{
		Level:     "info",
		Format:    "text",
		Output:    "stdout",
		DebugMode: false,
		Command:   "haki",
		File: LoggerFile{
			Path:       "./logs",
			Name:       "app.log",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
		},
	}
}
