package main

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

func EnsureDir(directory string) error {
	if info, err := os.Stat(directory); os.IsNotExist(err) {
		err := os.MkdirAll(directory, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %v", directory, err)
		}
	} else if err != nil {
		return err
	} else if !info.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", directory)
	}
	return nil
}

func IsImage(filePath string) bool {
	ext := filepath.Ext(filePath)
	mimeType := mime.TypeByExtension(ext)
	return strings.HasPrefix(mimeType, "image/")
}

func ListFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var filePaths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filePaths = append(filePaths, filepath.Join(dir, entry.Name()))
	}
	return filePaths, nil
}
