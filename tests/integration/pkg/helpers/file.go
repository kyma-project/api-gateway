package helpers

import (
	"fmt"
	"io"
	"log"
	"os"
)

func LoadFile(filePath string) ([]byte, error) {
	if filePath == "" {
		return nil, fmt.Errorf("path is empty %s", filePath)
	}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("can't open file %s because of %w", filePath, err)
	}
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			log.Printf("error while closing file %s: %s", filePath, err.Error())
		}
	}(file)

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("can't read file %s because of %w: err", filePath, err)
	}

	return byteValue, nil
}
