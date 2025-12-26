package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func GenerateImageVideo(image, outputDir string, duration time.Duration, width, height int) (string, error) {
	base := filepath.Base(image)
	baseWithoutExt := strings.TrimSuffix(base, filepath.Ext(base))
	fn := baseWithoutExt + ".mkv"
	fp := filepath.Join(outputDir, fn)

	cmd := exec.Command(
		"ffmpeg",
		"-hide_banner",
		"-f", "lavfi",
		"-i", "anullsrc=channel_layout=stereo:sample_rate=48000",
		"-loop", "1",
		"-i", image,
		"-avoid_negative_ts", "make_zero",
		//"-filter:v", "fps=30",
		"-r", "30",
		"-frames:v", fmt.Sprintf("%d", int(duration.Seconds()*30)),
		"-c:v", "libx264",
		"-crf", "20",
		"-preset", "medium",
		"-c:a", "aac",
		"-b:a", "128k",
		"-ar", "48000",
		"-ac", "2",
		"-shortest",
		"-g", "30",
		"-keyint_min", "1",
		"-t", ConvertDurationToTimestamp(duration),
		"-pix_fmt", "yuv420p",
		"-vf", fmt.Sprintf(
			"scale=%d:%d:force_original_aspect_ratio=decrease:eval=frame,pad=%d:%d:-1:-1:color=black,setsar=1",
			width,
			height,
			width,
			height,
		),
		fp,
		"-y",
	)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("ffmpeg command failed for %s: %v\nOutput: %s", image, err, string(output))
	}

	return fp, nil
}
