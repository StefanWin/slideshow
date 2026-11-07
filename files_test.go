package main

import "testing"

func TestIsImageAndIsVideo(t *testing.T) {
	tests := []struct {
		path               string
		wantImage, wantVid bool
	}{
		{path: "a.jpg", wantImage: true},
		{path: "a.JPEG", wantImage: true},
		{path: "dir/a.heic", wantImage: true},
		{path: "a.mp4", wantVid: true},
		{path: "a.MOV", wantVid: true},
		{path: "a.mkv", wantVid: true},
		{path: "a.webm", wantVid: true},
		{path: "a.ts"},
		{path: "a.txt"},
		{path: "noext"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsImage(tt.path); got != tt.wantImage {
				t.Errorf("IsImage(%q) = %t, want %t", tt.path, got, tt.wantImage)
			}
			if got := IsVideo(tt.path); got != tt.wantVid {
				t.Errorf("IsVideo(%q) = %t, want %t", tt.path, got, tt.wantVid)
			}
		})
	}
}
