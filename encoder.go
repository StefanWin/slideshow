package main

import (
	"fmt"
	"slices"
	"strings"
)

// codecProfile describes which quality options an encoder understands.
type codecProfile struct {
	defaultCRF int
	maxCRF     int
	presets    bool
	// tunes are encoder specific; empty means no -tune is passed
	imageTune, videoTune string
	// extra arguments always passed for this encoder
	extra []string
}

// CRF ranges and defaults follow the encoders' own documentation. libvpx-vp9 and
// libaom-av1 only use constant quality mode when the bitrate is set to 0.
var codecProfiles = map[string]codecProfile{
	"libx264":    {defaultCRF: 23, maxCRF: 51, presets: true, imageTune: "stillimage", videoTune: "film"},
	"libx265":    {defaultCRF: 28, maxCRF: 51, presets: true},
	"libvpx-vp9": {defaultCRF: 31, maxCRF: 63, extra: []string{"-b:v", "0"}},
	"libaom-av1": {defaultCRF: 30, maxCRF: 63, extra: []string{"-b:v", "0"}},
	"libsvtav1":  {defaultCRF: 35, maxCRF: 63, presets: true},
}

// crfUnset marks that -crf was not passed, so the codec default applies.
const crfUnset = -1

func supportedCodecs() string {
	codecs := make([]string, 0, len(codecProfiles))
	for codec := range codecProfiles {
		codecs = append(codecs, codec)
	}
	slices.Sort(codecs)
	return strings.Join(codecs, ", ")
}

// resolveEncoderOptions checks -crf and -preset against the codec and fills in the codec's
// default CRF. Codecs without a profile get no quality options, so passing -crf or -preset
// with them is an error rather than being silently ignored.
func resolveEncoderOptions(options *SlideshowOptions) error {
	profile, known := codecProfiles[options.Codec]

	if !known {
		if options.CRF != crfUnset || options.Preset != "" {
			return fmt.Errorf("-crf and -preset are only supported for %s, not %s", supportedCodecs(), options.Codec)
		}
		return nil
	}

	if options.CRF == crfUnset {
		options.CRF = profile.defaultCRF
	}
	if options.CRF < 0 || options.CRF > profile.maxCRF {
		return fmt.Errorf("crf for %s must be between 0 and %d", options.Codec, profile.maxCRF)
	}
	if options.Preset != "" && !profile.presets {
		return fmt.Errorf("-preset is not supported for %s", options.Codec)
	}
	return nil
}

type clipKind int

const (
	imageClip clipKind = iota
	videoClip
)

// videoEncoderArgs returns the encoder arguments for options.Codec. resolveEncoderOptions
// must have been called first.
func videoEncoderArgs(options *SlideshowOptions, kind clipKind) []string {
	args := []string{"-c:v", options.Codec}

	profile, known := codecProfiles[options.Codec]
	if !known {
		return args
	}

	args = append(args, "-crf", fmt.Sprintf("%d", options.CRF))
	if options.Preset != "" {
		args = append(args, "-preset", options.Preset)
	}

	tune := profile.videoTune
	if kind == imageClip {
		tune = profile.imageTune
	}
	if tune != "" {
		args = append(args, "-tune", tune)
	}

	return append(args, profile.extra...)
}

func describeEncoder(options *SlideshowOptions) string {
	if _, known := codecProfiles[options.Codec]; !known {
		return fmt.Sprintf("%s, encoder defaults", options.Codec)
	}
	preset := options.Preset
	if preset == "" {
		preset = "default"
	}
	return fmt.Sprintf("%s, crf: %d, preset: %s", options.Codec, options.CRF, preset)
}
