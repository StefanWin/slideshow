# slideshow

A simple CLI tool to generate a slideshow video from a directory of images and videos using FFmpeg.

## Prerequisites

This tool requires `ffmpeg` to be installed and available in your `PATH`.

## Installation

```bash
go build -o slideshow
```

## Usage

Run the tool by pointing it to a directory containing images and/or videos:

```bash
./slideshow -directory ./my-photos -entry-duration 3
```

### Options

| Option            | Default       | Description                                                      |
|-------------------|---------------|------------------------------------------------------------------|
| `-directory`      | `.`           | The directory to scan for image and video files.                 |
| `-width`          | `1920`        | Width of the output video in pixels.                             |
| `-height`         | `1080`        | Height of the output video in pixels.                            |
| `-codec`          | `libx264`     | Codec to use to generate the slideshow (and intermediate files). |
| `-entry-duration` | `5`           | Duration (in seconds) each image is displayed.                   |
| `-randomize`      | `false`       | Randomize the order of files in the slideshow.                   |
| `-recursive`      | `false`       | Recursively scan subdirectories for image and video files.       |
| `-fps`            | `30`          | Frames per second of the output video.                           |
| `-crf`            | `23`          | CRF value for the codec (`0` is lossless).                       |
| `-preset`         | `medium`      | Preset for the codec.                                            |
| `-concurrency`    | `CPU_COUNT/2` | Number of concurrent processing goroutines (at least 1).         |

## How it works

1. **Scan**: The tool scans the specified directory for image and video files (based on MIME types). By default, only the top-level directory is scanned; pass `-recursive` to include subdirectories.
2. **Process**: Each file is converted into a short clip with the target resolution and frame rate.
   - Images are shown for `-entry-duration` seconds; videos keep their own length.
   - Everything is scaled to fit the target resolution while maintaining the aspect ratio, with black padding added if necessary.
   - Images, and videos without audio, get a silent audio track to ensure smooth concatenation.
   - Files that fail to convert are logged and skipped.
3. **Assemble**: All intermediate clips are stored in a temporary directory in the system temp location.
4. **Concatenate**: The tool uses FFmpeg's `concat` demuxer to join all clips into a single MKV file.
5. **Cleanup**: The temporary directory is removed after the final video is generated.

The output file is written to the current working directory and named after the input directory and a timestamp, e.g., `my-photos-2025-12-26_05-05-00.mkv`.

## License

[MIT](LICENSE) © [Stefan Wintergerst](https://stefanwintergerst.com)
