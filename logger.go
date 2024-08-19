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

func initLogger(cfg *Config) error {
	var (
		logWriter  io.Writer
		logHandler slog.Handler
		err        error
	)
	sLvl := slog.LevelDebug
	switch cfg.Logger.Level {
	case "debug":
		sLvl = slog.LevelDebug
	case "info":
		sLvl = slog.LevelInfo
	case "warn":
		sLvl = slog.LevelWarn
	case "error":
		sLvl = slog.LevelError
	}

	switch cfg.Logger.Output {
	case "stdout":
		logWriter = os.Stdout
	case "stderr":
		logWriter = os.Stderr
	case "file":
		cfg.Logger.File.Name, err = createLogFilename(cfg.Logger)
		if err != nil {
			return fmt.Errorf("create log filename: %w", err)
		}
		file, err := os.OpenFile(path.Join(cfg.Logger.File.Path, cfg.Logger.File.Name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
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
	if cfg.Logger.Format == "json" {
		logHandler = slog.NewJSONHandler(logWriter, opts)
	} else if cfg.Logger.Format == "tint" {
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

func createLogFilename(cfg *ConfigLogger) (string, error) {
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
