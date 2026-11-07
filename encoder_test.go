package main

import (
	"slices"
	"testing"
)

func TestResolveEncoderOptions(t *testing.T) {
	tests := []struct {
		name    string
		codec   string
		crf     int
		preset  string
		wantCRF int
		wantErr bool
	}{
		{name: "x264Defaults", codec: "libx264", crf: crfUnset, wantCRF: 23},
		{name: "x265Defaults", codec: "libx265", crf: crfUnset, wantCRF: 28},
		{name: "vp9Defaults", codec: "libvpx-vp9", crf: crfUnset, wantCRF: 31},
		{name: "LosslessAllowed", codec: "libx264", crf: 0, wantCRF: 0},
		{name: "x264CRFTooHigh", codec: "libx264", crf: 52, wantErr: true},
		{name: "AV1AllowsHigherCRF", codec: "libsvtav1", crf: 63, preset: "8", wantCRF: 63},
		{name: "NegativeCRF", codec: "libx264", crf: -5, wantErr: true},
		{name: "PresetNotSupported", codec: "libvpx-vp9", crf: crfUnset, preset: "slow", wantErr: true},
		{name: "UnknownCodecDefaults", codec: "h264_nvenc", crf: crfUnset, wantCRF: crfUnset},
		{name: "UnknownCodecWithCRF", codec: "h264_nvenc", crf: 23, wantErr: true},
		{name: "UnknownCodecWithPreset", codec: "h264_nvenc", crf: crfUnset, preset: "p5", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := &SlideshowOptions{Codec: tt.codec, CRF: tt.crf, Preset: tt.preset}
			err := resolveEncoderOptions(options)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveEncoderOptions() error = %v, wantErr %t", err, tt.wantErr)
			}
			if !tt.wantErr && options.CRF != tt.wantCRF {
				t.Errorf("CRF = %d, want %d", options.CRF, tt.wantCRF)
			}
		})
	}
}

func TestVideoEncoderArgs(t *testing.T) {
	tests := []struct {
		name    string
		options SlideshowOptions
		kind    clipKind
		want    []string
	}{
		{
			name:    "x264Image",
			options: SlideshowOptions{Codec: "libx264", CRF: 23, Preset: "slow"},
			kind:    imageClip,
			want:    []string{"-c:v", "libx264", "-crf", "23", "-preset", "slow", "-tune", "stillimage"},
		},
		{
			name:    "x264VideoNoPreset",
			options: SlideshowOptions{Codec: "libx264", CRF: 23},
			kind:    videoClip,
			want:    []string{"-c:v", "libx264", "-crf", "23", "-tune", "film"},
		},
		{
			name:    "x265HasNoTune",
			options: SlideshowOptions{Codec: "libx265", CRF: 28},
			kind:    imageClip,
			want:    []string{"-c:v", "libx265", "-crf", "28"},
		},
		{
			name:    "vp9ConstantQuality",
			options: SlideshowOptions{Codec: "libvpx-vp9", CRF: 31},
			kind:    videoClip,
			want:    []string{"-c:v", "libvpx-vp9", "-crf", "31", "-b:v", "0"},
		},
		{
			name:    "UnknownCodec",
			options: SlideshowOptions{Codec: "h264_nvenc", CRF: crfUnset},
			kind:    videoClip,
			want:    []string{"-c:v", "h264_nvenc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := videoEncoderArgs(&tt.options, tt.kind)
			if !slices.Equal(got, tt.want) {
				t.Errorf("videoEncoderArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
