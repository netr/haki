package lib

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// SaveFile saves data to a file
func SaveFile(fileName string, data []byte) error {
	fd, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer func() { _ = fd.Close() }()

	_, err = fd.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// FileExists checks if a file exists at the given path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ValidateOutputPath validates the output path to ensure it is a valid file path and writable.
func ValidateOutputPath(output string) error {
	if output == "" {
		return errors.New("output path is empty")
	}

	// Check if the path is a directory
	info, err := os.Stat(output)
	if err == nil && info.IsDir() {
		return errors.New("output path is a directory")
	}

	// Check if the directory exists and is writable
	dir := filepath.Dir(output)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create output directory: %w", err)
	}

	// Check if the file is writable
	fd, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("cannot write to output file: %w", err)
	}
	defer func() { _ = fd.Close() }()

	return nil
}

// GetEnv returns the value of an environment variable or a default value if the environment variable is not set.
func GetEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetEnvInt returns the value of an environment variable as an integer or a default value if the environment variable is not set or is not a valid integer.
func GetEnvInt(key string, defaultValue int) int {
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
