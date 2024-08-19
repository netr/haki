package lib

import (
	"fmt"
	"os"
)

// SaveFile saves data to a file
func SaveFile(fileName string, data []byte) error {
	f, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}
