package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

var (
	VERSION = "dev"
)

func ensureBinaryExists(binary string) error {
	_, err := exec.LookPath(binary)
	if err != nil {
		return fmt.Errorf("%s not found in $PATH", binary)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

type IntermediateVideoJob struct {
	Path  string
	Index int
}

type IntermediateVideoResult struct {
	Path  string
	Index int
}

func generateIntermediateVideo(jobChannel <-chan IntermediateVideoJob, resultChannel chan<- IntermediateVideoResult, wg *sync.WaitGroup, outputDirectory string, options *SlideshowOptions) {

	defer wg.Done()

	for imageFile := range jobChannel {
		generatedVideo, err := GenerateIntermediateVideo(imageFile.Path, outputDirectory, options)
		if err != nil {
			log.Printf("failed to generate video from image %v: %v\n", imageFile, err)
		}

		resultChannel <- IntermediateVideoResult{Path: generatedVideo, Index: imageFile.Index}
	}
}

func run() error {

	if err := ensureBinaryExists("ffmpeg"); err != nil {
		return err
	}

	options, err := parseSlideshowOptions()
	if err != nil {
		return err
	}

	log.Printf("slideshow@%s\n", VERSION)
	log.Printf("directory: %s\n", options.Directory)
	log.Printf("%fs per image, randomize order: %t\n", options.EntryDuration.Seconds(), options.Randomize)
	log.Printf("recursive scanning: %t\n", options.Recursive)
	log.Printf("output: %dx%d@%d (%s, crf: %d, preset: %s)\n", options.Width, options.Height, options.FPS, options.Codec, options.CRF, options.Preset)

	files, err := ListFiles(options.Directory, options.Recursive)
	if err != nil {
		return err
	}

	log.Printf("found %d files in %s\n", len(files), options.Directory)

	var filesToProcess []string

	for _, file := range files {
		if IsImage(file) || IsVideo(file) {
			filesToProcess = append(filesToProcess, file)
		}
	}

	if options.Recursive {
		log.Printf("found %d image/video files in %s and subdirectories\n", len(filesToProcess), options.Directory)
	} else {
		log.Printf("found %d image/video files in %s\n", len(filesToProcess), options.Directory)
	}

	if len(filesToProcess) == 0 {
		return fmt.Errorf("no image/video files found in %s", options.Directory)
	}

	if options.Randomize {
		rand.Shuffle(len(filesToProcess), func(i, j int) {
			filesToProcess[i], filesToProcess[j] = filesToProcess[j], filesToProcess[i]
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
		log.Printf("cleaned up temp directory %s\n", path)
	}(tmpDir)

	log.Printf("using %d concurrent workers\n", options.Concurrency)

	intermediateFiles := make([]string, len(filesToProcess))

	log.Printf("generating intermediate videos...\n")
	start := time.Now()

	var wg sync.WaitGroup
	jobChannel := make(chan IntermediateVideoJob, len(filesToProcess))
	resultChannel := make(chan IntermediateVideoResult, len(filesToProcess))

	for i := 0; i < options.Concurrency; i++ {
		wg.Add(1)
		go generateIntermediateVideo(jobChannel, resultChannel, &wg, tmpDir, options)
	}

	for i, imageFile := range filesToProcess {
		jobChannel <- IntermediateVideoJob{Path: imageFile, Index: i}
	}

	close(jobChannel)

	go func() {
		wg.Wait()
		close(resultChannel)
	}()

	i := 0
	for result := range resultChannel {
		intermediateFiles[result.Index] = result.Path
		i++
	}

	elapsed := time.Since(start)
	log.Printf("generated %d intermediate videos in %s\n", len(intermediateFiles), elapsed)

	dirName := filepath.Base(options.Directory)

	outputPath := fmt.Sprintf("%s-%s.mkv", dirName, GetTimestamp())

	log.Printf("concatenating intermediate videos...\n")
	startTime := time.Now()
	if err := ConcatVideos(intermediateFiles, outputPath, options); err != nil {
		return fmt.Errorf("failed to concat videos: %v", err)
	}
	elapsed = time.Since(startTime)

	log.Printf("output video written to %s in %s", outputPath, elapsed)
	return nil
}
