package main

import (
	"flag"
	"fmt"
	"runtime"
	"time"
)

type SlideshowOptions struct {
	Directory     string
	Width, Height int
	Codec         string
	FPS           int
	EntryDuration time.Duration
	Randomize     bool
	Recursive     bool
	Concurrency   int
}

func parseSlideshowOptions() (*SlideshowOptions, error) {
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

	if width <= 0 {
		return nil, fmt.Errorf("width must be greater than 0")
	}
	if height <= 0 {
		return nil, fmt.Errorf("height must be greater than 0")
	}
	if fps <= 0 {
		return nil, fmt.Errorf("fps must be greater than 0")
	}
	if concurrency <= 0 {
		return nil, fmt.Errorf("concurrency must be greater than 0")
	}
	if entryDuration <= 0 {
		return nil, fmt.Errorf("entry-duration must be greater than 0")
	}

	return &SlideshowOptions{
		Directory:     directory,
		Width:         width,
		Height:        height,
		Codec:         codec,
		FPS:           fps,
		EntryDuration: time.Duration(entryDuration) * time.Second,
		Randomize:     randomize,
		Recursive:     recursive,
		Concurrency:   concurrency,
	}, nil
}
