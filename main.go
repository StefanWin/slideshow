package main

import (
	"fmt"
	"log"
	"math/rand/v2"
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
	Err   error
}

func generateIntermediateVideo(jobChannel <-chan IntermediateVideoJob, resultChannel chan<- IntermediateVideoResult, wg *sync.WaitGroup, outputDirectory string, options *SlideshowOptions) {

	defer wg.Done()

	for job := range jobChannel {
		generatedVideo, err := GenerateIntermediateVideo(job.Path, outputDirectory, job.Index, options)
		resultChannel <- IntermediateVideoResult{Path: generatedVideo, Index: job.Index, Err: err}
	}
}

func run() error {

	options, err := parseSlideshowOptions()
	if err != nil {
		return err
	}

	if err := ensureBinaryExists("ffmpeg"); err != nil {
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

	tmpDir, err := os.MkdirTemp("", "slideshow-*")
	if err != nil {
		return fmt.Errorf("failed to create tmp dir: %v", err)
	}

	defer func(path string) {
		if err := os.RemoveAll(path); err != nil {
			log.Printf("failed to cleanup tmp dir: %v", err)
			return
		}
		log.Printf("cleaned up temp directory %s\n", path)
	}(tmpDir)

	log.Printf("using %d concurrent workers\n", options.Concurrency)

	results := make([]IntermediateVideoResult, len(filesToProcess))

	log.Printf("generating intermediate videos...\n")
	start := time.Now()

	var wg sync.WaitGroup
	jobChannel := make(chan IntermediateVideoJob, len(filesToProcess))
	resultChannel := make(chan IntermediateVideoResult, len(filesToProcess))

	for i := 0; i < options.Concurrency; i++ {
		wg.Add(1)
		go generateIntermediateVideo(jobChannel, resultChannel, &wg, tmpDir, options)
	}

	for i, file := range filesToProcess {
		jobChannel <- IntermediateVideoJob{Path: file, Index: i}
	}

	close(jobChannel)

	go func() {
		wg.Wait()
		close(resultChannel)
	}()

	for result := range resultChannel {
		results[result.Index] = result
	}

	// keep the original order and skip clips that failed, so one bad file doesn't abort the run
	var intermediateFiles []string
	for i, result := range results {
		if result.Err != nil {
			log.Printf("skipping %s: %v\n", filesToProcess[i], result.Err)
			continue
		}
		intermediateFiles = append(intermediateFiles, result.Path)
	}

	elapsed := time.Since(start)
	log.Printf("generated %d of %d intermediate videos in %s\n", len(intermediateFiles), len(filesToProcess), elapsed)

	if len(intermediateFiles) == 0 {
		return fmt.Errorf("failed to generate any intermediate videos")
	}

	absDirectory, err := filepath.Abs(options.Directory)
	if err != nil {
		return err
	}
	dirName := filepath.Base(absDirectory)

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
