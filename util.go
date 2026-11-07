package main

import (
	"fmt"
	"time"
)

func GetTimestamp() string {
	return time.Now().Format("2006-01-02_15-04-05")
}

func ConvertDurationToTimestamp(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	milliseconds := int(d.Milliseconds()) % 1000

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
}
