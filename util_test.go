package main

import (
	"testing"
	"time"
)

func TestGetTimestamp(t *testing.T) {
	t.Run("FormatCheck", func(t *testing.T) {
		timestamp := GetTimestamp()
		expectedFormat := "2006-01-02_15-04-05" // Go's reference time format
		parsedTime, err := time.Parse(expectedFormat, timestamp)
		if err != nil || parsedTime.IsZero() {
			t.Errorf("GetTimestamp() produced an invalid timestamp. Got: %s, Error: %v", timestamp, err)
		}
	})
}

func TestConvertDurationToTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		{
			name:     "ZeroDuration",
			input:    0,
			expected: "00:00:00.000",
		},
		{
			name:     "OneSecond",
			input:    time.Second,
			expected: "00:00:01.000",
		},
		{
			name:     "OneMinute",
			input:    time.Minute,
			expected: "00:01:00.000",
		},
		{
			name:     "OneHour",
			input:    time.Hour,
			expected: "01:00:00.000",
		},
		{
			name:     "ComplexDuration",
			input:    3*time.Hour + 15*time.Minute + 42*time.Second + 500*time.Millisecond,
			expected: "03:15:42.500",
		},
		{
			name:     "MillisecondsOnly",
			input:    900 * time.Millisecond,
			expected: "00:00:00.900",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertDurationToTimestamp(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertDurationToTimestamp(%v) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsPreviousOutput(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{path: OutputFileName("my photos", "2025-12-26_05-05-00"), want: true},
		{path: "sub/my photos-2025-12-26_05-05-00.mkv", want: true},
		{path: "my photos-2025-12-26_05-05-00.mp4"},
		{path: "my photos-holiday.mkv"},
		{path: "other-2025-12-26_05-05-00.mkv"},
		{path: "my photos.mkv"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsPreviousOutput(tt.path, "my photos"); got != tt.want {
				t.Errorf("IsPreviousOutput(%q) = %t, want %t", tt.path, got, tt.want)
			}
		})
	}
}
