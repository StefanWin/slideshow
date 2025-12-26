package main

import (
	"fmt"
	"io/fs"
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

func hasMimeTypePrefix(filePath, prefix string) bool {
	ext := filepath.Ext(filePath)
	mimeType := mime.TypeByExtension(ext)
	return strings.HasPrefix(mimeType, prefix)
}

func IsImage(filePath string) bool {
	return hasMimeTypePrefix(filePath, "image/")
}

func IsVideo(filePath string) bool {
	return hasMimeTypePrefix(filePath, "video/")
}

func ListFiles(dir string, recursive bool) ([]string, error) {
	var filePaths []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == dir {
			return nil
		}

		if d.IsDir() {
			if !recursive {
				return filepath.SkipDir
			}
			return nil
		}

		filePaths = append(filePaths, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return filePaths, nil
}
