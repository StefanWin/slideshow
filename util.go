package main

import (
	"fmt"
	"time"
)

func GetTimestamp() string {
	now := time.Now()

	return fmt.Sprintf("%02d-%02d-%02d_%02d-%02d-%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
}

func ConvertDurationToTimestamp(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	milliseconds := int(d.Milliseconds()) % 1000

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
}
