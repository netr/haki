package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/lmittmann/tint"
)

type LoggerConfig struct {
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

func newLoggerConfig(command string) *LoggerConfig {
	return &LoggerConfig{
		Level:     getEnv("LOG_LEVEL", "info"),
		Format:    getEnv("LOG_FORMAT", "text"),
		Output:    getEnv("LOG_OUTPUT", "stdout"),
		DebugMode: getEnv("DEBUG_MODE", "false") == "true",
		Command:   command,
		File: LoggerFile{
			Path:       getEnv("LOG_FILE_PATH", "./logs"),
			Name:       getEnv("LOG_FILE_NAME", "app.log"),
			MaxSize:    getEnvInt("LOG_FILE_MAX_SIZE", 100),
			MaxBackups: getEnvInt("LOG_FILE_MAX_BACKUPS", 3),
			MaxAge:     getEnvInt("LOG_FILE_MAX_AGE", 28),
		},
	}
}

func initLogger(cfg *LoggerConfig) error {
	var (
		logWriter  io.Writer
		logHandler slog.Handler
		err        error
	)
	sLvl := slog.LevelDebug
	switch cfg.Level {
	case "debug":
		sLvl = slog.LevelDebug
	case "info":
		sLvl = slog.LevelInfo
	case "warn":
		sLvl = slog.LevelWarn
	case "error":
		sLvl = slog.LevelError
	}

	switch cfg.Output {
	case "stdout":
		logWriter = os.Stdout
	case "stderr":
		logWriter = os.Stderr
	case "file":
		cfg.File.Name, err = createLogFilename(cfg)
		if err != nil {
			return fmt.Errorf("create log filename: %w", err)
		}
		file, err := os.OpenFile(path.Join(cfg.File.Path, cfg.File.Name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("open log file: %w", err)
		}
		logWriter = file
	}

	opts := &slog.HandlerOptions{
		Level:     sLvl,
		AddSource: true,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.SourceKey {
				source := attr.Value.Any().(*slog.Source)
				source.File = filepath.Base(source.File)
			}
			return attr
		},
	}

	logHandler = slog.NewTextHandler(logWriter, opts)
	if cfg.Format == "json" {
		logHandler = slog.NewJSONHandler(logWriter, opts)
	} else if cfg.Format == "tint" {
		logHandler = tint.NewHandler(logWriter, &tint.Options{
			Level:      sLvl,
			TimeFormat: time.Kitchen,
			AddSource:  true,
			ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
				if attr.Key == slog.SourceKey {
					source := attr.Value.Any().(*slog.Source)
					source.File = filepath.Base(source.File)
				}
				return attr
			},
		})
	}

	slog.SetDefault(slog.New(logHandler))
	return nil
}

var (
	ErrInvalidLogFilename = errors.New("invalid log filename")
)

func createLogFilename(cfg *LoggerConfig) (string, error) {
	fnSplit := strings.Split(cfg.File.Name, ".")
	fName := fnSplit[0]
	switch len(fnSplit) {
	case 1:
		fName += cfg.Command + ".log"
	case 2:
		fName += "_" + cfg.Command + "." + fnSplit[1]
	default:
		return "", ErrInvalidLogFilename
	}
	return fName, nil
}
