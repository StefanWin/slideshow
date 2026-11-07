package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const timestampLayout = "2006-01-02_15-04-05"

func GetTimestamp() string {
	return time.Now().Format(timestampLayout)
}

// OutputFileName is the name of the slideshow generated from the directory named dirName.
func OutputFileName(dirName, timestamp string) string {
	return fmt.Sprintf("%s-%s.mkv", dirName, timestamp)
}

// IsPreviousOutput reports whether path is a slideshow generated earlier from the directory
// named dirName, so it isn't picked up as input when the output lands in the scanned directory.
func IsPreviousOutput(path, dirName string) bool {
	name := filepath.Base(path)
	prefix := dirName + "-"
	if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".mkv") {
		return false
	}
	timestamp := strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".mkv")
	_, err := time.Parse(timestampLayout, timestamp)
	return err == nil
}

func ConvertDurationToTimestamp(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	milliseconds := int(d.Milliseconds()) % 1000

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
}
