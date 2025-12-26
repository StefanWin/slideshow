package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

type IntermediateVideoOptions struct {
	OutputDirectory    string
	Codec              string
	EntryDuration      time.Duration
	Width, Height, FPS int
}

type IntermediateVideoJob struct {
	Path  string
	Index int
}

type IntermediateVideoResult struct {
	Path  string
	Index int
}

func generateIntermediateVideo(jobChannel <-chan IntermediateVideoJob, resultChannel chan<- IntermediateVideoResult, wg *sync.WaitGroup, options *IntermediateVideoOptions) {

	defer wg.Done()

	for imageFile := range jobChannel {
		generatedVideo, err := GenerateImageVideo(imageFile.Path, options.OutputDirectory, options.Codec, options.EntryDuration, options.Width, options.Height, options.FPS)
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

	var directory string
	var width int
	var height int
	var fps int
	var codec string
	var entryDuration int
	var randomize bool
	var recursive bool
	var concurrency int

	flag.StringVar(&directory, "directory", ".", "directory to scan")
	flag.IntVar(&width, "width", 1920, "width of output video")
	flag.IntVar(&height, "height", 1080, "height of output video")
	flag.IntVar(&fps, "fps", 30, "frames per second of output video")
	flag.StringVar(&codec, "codec", "libx264", "codec to use for output video")
	flag.IntVar(&entryDuration, "entry-duration", 5, "duration of each entry in seconds")
	flag.BoolVar(&randomize, "randomize", false, "randomize order of files")
	flag.BoolVar(&recursive, "recursive", false, "recursively scan subdirectories for image files")
	flag.IntVar(&concurrency, "concurrency", runtime.NumCPU()/2, "number of concurrent workers")

	flag.Parse()

	log.Printf("slideshow@%s\n", VERSION)
	log.Printf("directory: %s\n", directory)
	log.Printf("%ds per image, randomize order: %t\n", entryDuration, randomize)
	log.Printf("recursive scanning: %t\n", recursive)
	log.Printf("output: %dx%d@%d (%s)\n", width, height, fps, codec)

	files, err := ListFiles(directory, recursive)
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

	if recursive {
		log.Printf("found %d image files in %s and subdirectories\n", len(imageFiles), directory)
	} else {
		log.Printf("found %d image files in %s\n", len(imageFiles), directory)
	}

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
		log.Printf("cleaned up temp directory %s\n", path)
	}(tmpDir)

	log.Printf("using %d concurrent workers\n", concurrency)

	intermediateFiles := make([]string, len(imageFiles))

	log.Printf("generating intermediate videos...\n")
	start := time.Now()

	var wg sync.WaitGroup
	jobChannel := make(chan IntermediateVideoJob, len(imageFiles))
	resultChannel := make(chan IntermediateVideoResult, len(imageFiles))

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go generateIntermediateVideo(jobChannel, resultChannel, &wg, &IntermediateVideoOptions{
			OutputDirectory: tmpDir,
			Codec:           codec,
			EntryDuration:   time.Second * time.Duration(entryDuration),
			Width:           width,
			Height:          height,
			FPS:             fps,
		})
	}

	for i, imageFile := range imageFiles {
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
	log.Printf("generated %d intermediate videos in %dms\n", len(intermediateFiles), elapsed.Milliseconds())

	dirName := filepath.Base(directory)

	outputPath := fmt.Sprintf("%s-%s.mkv", dirName, GetTimestamp())

	log.Printf("concatenating intermediate videos...\n")
	startTime := time.Now()
	if err := ConcatVideos(intermediateFiles, outputPath, codec); err != nil {
		return fmt.Errorf("failed to concat videos: %v", err)
	}
	elapsed = time.Since(startTime)

	log.Printf("output video written to %s in %dms", outputPath, elapsed.Milliseconds())
	return nil
}
