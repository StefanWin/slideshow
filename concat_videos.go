package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ConcatVideos joins the intermediate clips without re-encoding them. This relies on every
// clip having been encoded with the same codec settings, frame rate, resolution, pixel
// format and audio format (see GenerateIntermediateVideo).
func ConcatVideos(videoFiles []string, outputPath string) error {

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
		"-c", "copy",
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
