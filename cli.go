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
	Preset        string
	FPS           int
	CRF           int
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
	var crf int
	var codec string
	var preset string
	var entryDuration int
	var randomize bool
	var recursive bool
	var concurrency int

	flag.StringVar(&directory, "directory", ".", "directory to scan")
	flag.IntVar(&width, "width", 1920, "width of output video")
	flag.IntVar(&height, "height", 1080, "height of output video")
	flag.IntVar(&fps, "fps", 30, "frames per second of output video")
	flag.IntVar(&crf, "crf", crfUnset, "constant rate factor for codec (default: the codec's default)")
	flag.StringVar(&preset, "preset", "", "preset to use for codec (default: the codec's default)")
	flag.StringVar(&codec, "codec", "libx264", "codec to use for output video ("+supportedCodecs()+", or any other ffmpeg encoder with its defaults)")
	flag.IntVar(&entryDuration, "entry-duration", 5, "duration of each entry in seconds")
	flag.BoolVar(&randomize, "randomize", false, "randomize order of files")
	flag.BoolVar(&recursive, "recursive", false, "recursively scan subdirectories for image and video files")
	flag.IntVar(&concurrency, "concurrency", max(1, runtime.NumCPU()/2), "number of concurrent workers")

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

	options := &SlideshowOptions{
		Directory:     directory,
		Width:         width,
		Height:        height,
		Preset:        preset,
		Codec:         codec,
		FPS:           fps,
		CRF:           crf,
		EntryDuration: time.Duration(entryDuration) * time.Second,
		Randomize:     randomize,
		Recursive:     recursive,
		Concurrency:   concurrency,
	}

	if err := resolveEncoderOptions(options); err != nil {
		return nil, err
	}

	return options, nil
}
