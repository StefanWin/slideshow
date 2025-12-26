package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var (
	VERSION = "dev"
)

func ensureBinaryExists(binary string) error {
	_, err := exec.LookPath(binary)
	if err != nil {
		return fmt.Errorf("%s not found in PATH", binary)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {

	if err := ensureBinaryExists("ffmpeg"); err != nil {
		return err
	}

	var directory string
	var width int
	var height int
	var codec string
	var entryDuration int
	var randomize bool

	flag.StringVar(&directory, "directory", ".", "directory to scan")
	flag.IntVar(&width, "width", 1920, "width of output video")
	flag.IntVar(&height, "height", 1080, "height of output video")
	flag.StringVar(&codec, "codec", "libx264", "codec to use for output video")
	flag.IntVar(&entryDuration, "entry-duration", 5, "duration of each entry in seconds")
	flag.BoolVar(&randomize, "randomize", false, "randomize order of files")

	flag.Parse()

	log.Printf("slideshow@%s\n", VERSION)
	log.Printf("directory: %s\n", directory)
	log.Printf("resolution: %dx%d\n", width, height)
	log.Printf("entry duration: %d\n", entryDuration)
	log.Printf("codec: %s\n", codec)

	files, err := ListFiles(directory)
	if err != nil {
		return err
	}

	log.Printf("found %d files in %s\n", len(files), directory)

	var imageFiles []string

	for _, file := range files {
		if IsImage(file) {
			imageFiles = append(imageFiles, file)
		}
	}

	log.Printf("found %d image files in %s\n", len(imageFiles), directory)

	if len(imageFiles) == 0 {
		return fmt.Errorf("no image files found in %s", directory)
	}

	if randomize {
		rand.Shuffle(len(imageFiles), func(i, j int) {
			imageFiles[i], imageFiles[j] = imageFiles[j], imageFiles[i]
		})
		log.Printf("randomized order of files")
	}

	tmpDir := ".tmp/"

	if err := EnsureDir(tmpDir); err != nil {
		return fmt.Errorf("failed to create tmp dir: %v", err)
	}

	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			log.Printf("failed to cleanup tmp dir: %v", err)
		}
	}(tmpDir)

	intermediateFiles := make([]string, len(imageFiles))

	for i, imageFile := range imageFiles {
		log.Printf("processing image %s (%d/%d)", imageFile, i+1, len(imageFiles))

		generatedVideo, err := GenerateImageVideo(imageFile, tmpDir, codec, time.Second*time.Duration(entryDuration), width, height)
		if err != nil {
			return fmt.Errorf("failed to generate video from image %s: %v", imageFile, err)
		}

		intermediateFiles[i] = generatedVideo
	}

	dirName := filepath.Base(directory)

	outputPath := fmt.Sprintf("%s-%s.mp4", dirName, GetTimestamp())

	if err := ConcatVideos(intermediateFiles, outputPath, codec); err != nil {
		return fmt.Errorf("failed to concat videos: %v", err)
	}

	log.Printf("output video written to %s", outputPath)
	return nil
}
