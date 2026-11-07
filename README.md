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

| Option            | Default         | Description                                                      |
|-------------------|-----------------|------------------------------------------------------------------|
| `-directory`      | `.`             | The directory to scan for image and video files.                 |
| `-width`          | `1920`          | Width of the output video in pixels.                             |
| `-height`         | `1080`          | Height of the output video in pixels.                            |
| `-codec`          | `libx264`       | Video encoder to use. See [Codecs](#codecs).                     |
| `-entry-duration` | `5`             | Duration (in seconds) each image is displayed.                   |
| `-randomize`      | `false`         | Randomize the order of files in the slideshow.                   |
| `-recursive`      | `false`         | Recursively scan subdirectories for image and video files.       |
| `-fps`            | `30`            | Frames per second of the output video.                           |
| `-crf`            | codec default   | CRF value for the codec (`0` is lossless).                       |
| `-preset`         | codec default   | Preset for the codec.                                            |
| `-concurrency`    | `CPU_COUNT/2`   | Number of concurrent processing goroutines (at least 1).         |

### Codecs

`-crf` and `-preset` are only passed to encoders that support them:

| Codec        | Default CRF | CRF range | `-preset`               |
|--------------|-------------|-----------|-------------------------|
| `libx264`    | `23`        | 0–51      | e.g. `ultrafast`–`veryslow` |
| `libx265`    | `28`        | 0–51      | e.g. `ultrafast`–`veryslow` |
| `libvpx-vp9` | `31`        | 0–63      | not supported           |
| `libaom-av1` | `30`        | 0–63      | not supported           |
| `libsvtav1`  | `35`        | 0–63      | `0`–`13`                |

Any other FFmpeg encoder (e.g. `h264_nvenc`) can be used with its own defaults; passing `-crf` or `-preset` with it is an error.

## How it works

1. **Scan**: The tool scans the specified directory for image and video files (by file extension, falling back to the system's MIME types for unknown extensions). By default, only the top-level directory is scanned; pass `-recursive` to include subdirectories. Slideshows generated earlier from the same directory are skipped.
2. **Process**: Each file is converted into a short clip with the same codec settings, resolution, frame rate and audio format.
   - Images are shown for `-entry-duration` seconds; videos keep their own length.
   - Everything is scaled to fit the target resolution while maintaining the aspect ratio, with black padding added if necessary.
   - Images, and videos without audio, get a silent audio track to ensure smooth concatenation.
   - Files that fail to convert are logged and skipped.
3. **Assemble**: All intermediate clips are stored in a temporary directory in the system temp location.
4. **Concatenate**: The tool uses FFmpeg's `concat` demuxer to join all clips into a single MKV file without re-encoding them.
5. **Cleanup**: The temporary directory is removed after the final video is generated.

The output file is written to the current working directory and named after the input directory and a timestamp, e.g., `my-photos-2025-12-26_05-05-00.mkv`.

## License

[MIT](LICENSE) © [Stefan Wintergerst](https://stefanwintergerst.com)
