package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

const silentAudioSource = "anullsrc=channel_layout=stereo:sample_rate=48000"

// All clips share the same audio format, pixel format, frame rate and resolution so the
// concat step can join them without re-encoding.
var audioEncoderArgs = []string{
	"-c:a", "aac",
	"-b:a", "128k",
	"-ar", "48000",
	"-ac", "2",
}

func scaleAndPadFilter(options *SlideshowOptions) string {
	return fmt.Sprintf(
		"scale=%d:%d:force_original_aspect_ratio=decrease:eval=frame,pad=%d:%d:-1:-1:color=black,setsar=1",
		options.Width,
		options.Height,
		options.Width,
		options.Height,
	)
}

func getImageOptions(file string, options *SlideshowOptions) []string {
	args := []string{
		"-hide_banner",
		// with -loop 1 an undecodable image makes ffmpeg retry forever; fail instead
		"-xerror",
		"-f", "lavfi",
		"-i", silentAudioSource,
		"-loop", "1",
		"-i", file,
		"-r", fmt.Sprintf("%d", options.FPS),
		"-frames:v", fmt.Sprintf("%d", int(options.EntryDuration.Seconds()*float64(options.FPS))),
	}
	args = append(args, videoEncoderArgs(options, imageClip)...)
	args = append(args, audioEncoderArgs...)
	return append(args,
		"-shortest",
		"-g", fmt.Sprintf("%d", options.FPS),
		"-keyint_min", "1",
		"-t", ConvertDurationToTimestamp(options.EntryDuration),
		"-pix_fmt", "yuv420p",
		"-vf", scaleAndPadFilter(options),
	)
}

// hasAudioStream reports whether ffmpeg finds at least one audio stream in file.
func hasAudioStream(file string) bool {
	cmd := exec.Command("ffmpeg", "-v", "error", "-i", file, "-map", "0:a:0", "-t", "0", "-f", "null", "-")
	return cmd.Run() == nil
}

func getVideoOptions(file string, options *SlideshowOptions) []string {
	args := []string{
		"-hide_banner",
		"-i", file,
	}

	// Every clip needs exactly one video and one audio stream for the concat demuxer
	// to line them up, so videos without audio get a silent track.
	if hasAudioStream(file) {
		args = append(args, "-map", "0:v:0", "-map", "0:a:0")
	} else {
		args = append(args,
			"-f", "lavfi",
			"-i", silentAudioSource,
			"-map", "0:v:0", "-map", "1:a:0",
			"-shortest",
		)
	}

	args = append(args, "-r", fmt.Sprintf("%d", options.FPS))
	args = append(args, videoEncoderArgs(options, videoClip)...)
	args = append(args, audioEncoderArgs...)
	return append(args,
		"-pix_fmt", "yuv420p",
		"-vf", scaleAndPadFilter(options),
	)
}

// GenerateIntermediateVideo encodes file into a clip in outputDir. The index prefix keeps
// clip names unique when inputs share a base name (e.g. a/1.jpg and b/1.jpg, or 1.jpg and 1.png).
func GenerateIntermediateVideo(file, outputDir string, index int, options *SlideshowOptions) (string, error) {
	base := filepath.Base(file)
	baseWithoutExt := strings.TrimSuffix(base, filepath.Ext(base))
	fn := fmt.Sprintf("%05d-%s.mkv", index, baseWithoutExt)
	fp := filepath.Join(outputDir, fn)

	var args []string

	if IsImage(file) {
		args = getImageOptions(file, options)
	} else {
		args = getVideoOptions(file, options)
	}

	args = append(args, fp, "-y")

	cmd := exec.Command("ffmpeg", args...)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("ffmpeg command failed for %s: %v\nOutput: %s", file, err, string(output))
	}

	return fp, nil
}
