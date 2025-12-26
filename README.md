# slideshow

A simple CLI tool to generate a slideshow video from a directory of images using FFmpeg.

## Prerequisites

This tool requires `ffmpeg` and `ffprobe` to be installed and available in your `PATH`.

## Installation

```bash
go build -o slideshow
```

## Usage

Run the tool by pointing it to a directory containing images:

```bashx
./slideshow -directory ./my-photos -entry-duration 3
```

### Options

| Option | Default | Description |
|--------|---------|-------------|
| `-directory` | `.` | The directory to scan for image files. |
| `-width` | `1920` | Width of the output video in pixels. |
| `-height` | `1080` | Height of the output video in pixels. |
| `-entry-duration` | `5` | Duration (in seconds) each image is displayed. |

## How it works

1. **Scan**: The tool scans the specified directory for image files (based on MIME types).
2. **Process**: For each image found, it uses FFmpeg to create a short video clip with the specified duration and resolution.
   - Images are scaled to fit the target resolution while maintaining an aspect ratio, with black padding added if necessary.
   - A silent audio track is added to each clip to ensure smooth concatenation.
3. **Assemble**: All intermediate video clips are stored in a temporary `.tmp/` directory.
4. **Concatenate**: The tool uses FFmpeg's `concat` demuxer to join all clips into a single MP4 file.
5. **Cleanup**: The temporary directory is removed after the final video is generated.

The output file is named based on the input directory and a timestamp, e.g., `my-photos-2025-12-26_05-05-00.mp4`.
