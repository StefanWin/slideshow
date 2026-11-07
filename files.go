package main

import (
	"fmt"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"slices"
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

// Known extensions are checked first so detection doesn't depend on the OS MIME table:
// Go's built-in table has no video types, and minimal systems ship without one.
// .ts is ignored because it is far more often TypeScript than MPEG-TS, even though some
// MIME tables (e.g. Windows) list it as video.
var (
	imageExtensions   = []string{".jpg", ".jpeg", ".jfif", ".png", ".webp", ".gif", ".bmp", ".tif", ".tiff", ".avif", ".heic", ".heif"}
	videoExtensions   = []string{".mp4", ".m4v", ".mov", ".mkv", ".webm", ".avi", ".wmv", ".flv", ".mpg", ".mpeg", ".mts", ".m2ts", ".3gp", ".ogv"}
	ignoredExtensions = []string{".ts"}
)

func hasMimeTypePrefix(filePath, prefix string) bool {
	ext := filepath.Ext(filePath)
	mimeType := mime.TypeByExtension(ext)
	return strings.HasPrefix(mimeType, prefix)
}

func extension(filePath string) string {
	return strings.ToLower(filepath.Ext(filePath))
}

func IsImage(filePath string) bool {
	ext := extension(filePath)
	if slices.Contains(imageExtensions, ext) {
		return true
	}
	return !slices.Contains(videoExtensions, ext) && !slices.Contains(ignoredExtensions, ext) && hasMimeTypePrefix(filePath, "image/")
}

func IsVideo(filePath string) bool {
	ext := extension(filePath)
	if slices.Contains(videoExtensions, ext) {
		return true
	}
	return !slices.Contains(imageExtensions, ext) && !slices.Contains(ignoredExtensions, ext) && hasMimeTypePrefix(filePath, "video/")
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
