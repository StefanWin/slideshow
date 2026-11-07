package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ConcatVideos(videoFiles []string, outputPath string, options *SlideshowOptions) error {

	// Create a temporary file list for ffmpeg
	tempList, err := os.CreateTemp("", "ffmpeg_filelist_*.txt")
	if err != nil {
		return err
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			log.Println(err)
		}
	}(tempList.Name())

	for _, videoFile := range videoFiles {
		absPath, err := filepath.Abs(videoFile)
		if err != nil {
			return err
		}
		// the concat demuxer requires single quotes inside quoted paths to be escaped as '\''
		escaped := strings.ReplaceAll(absPath, "'", `'\''`)
		if _, err := fmt.Fprintf(tempList, "file '%s'\n", escaped); err != nil {
			return err
		}
	}
	err = tempList.Close()
	if err != nil {
		return err
	}

	args := []string{
		"-hide_banner",
		"-f", "concat",
		"-safe", "0",
		"-i", tempList.Name(),
		"-c:v", options.Codec,
		"-crf", fmt.Sprintf("%d", options.CRF),
		"-preset", options.Preset,
		"-tune", "stillimage",
		"-movflags", "+faststart",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", "128k",
		"-ar", "48000",
		"-ac", "2",
		outputPath,
		"-y",
	}

	concatCmd := exec.Command("ffmpeg", args...)
	output, err := concatCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg concat failed: %v\nOutput: %s", err, string(output))
	}

	return nil
}
