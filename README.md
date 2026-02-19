# MediaStream

> Stream any image or video file as a virtual webcam feed over HTTP — with a cross-platform GUI and CLI.

MediaStream turns a static image, animated GIF, or video file into a continuous [MJPEG](https://en.wikipedia.org/wiki/Motion_JPEG) stream that any software expecting a camera feed can consume — OBS Studio, VLC, browser `<img>` tags, IP camera viewers, and more.

---

## Features

| Feature | Details |
|---|---|
| **Static images** | JPEG, PNG, WebP, BMP — re-streamed at your chosen FPS |
| **Animated GIFs** | Each frame replayed at its native delay, looping forever |
| **Video files** | MP4, MKV, MOV, AVI, WebM, FLV, and anything else FFmpeg handles |
| **Native GUI** | Cross-platform window (Windows · macOS · Linux) via [Fyne](https://fyne.io) |
| **Headless CLI** | `--headless` flag for scripting, containers, and servers |
| **Configurable** | Port and frame rate adjustable at runtime |
| **Auto-loop** | Videos and GIFs restart seamlessly when they reach the end |
| **Health check** | `GET /health` endpoint for uptime monitoring |

---

## Quick Start

### Download a pre-built binary

Head to the [Releases](https://github.com/idevakk/mediastream/releases) page and grab the binary for your platform. No installation required — just run it.

### Run from source

```bash
git clone https://github.com/idevakk/mediastream.git
cd mediastream
make run
```

---

## Prerequisites

| Requirement | Required for |
|---|---|
| **Go 1.21+** | Building from source |
| **FFmpeg** (in `PATH`) | Video file streaming (GIF and image streaming works without it) |
| **C compiler** | Building Fyne GUI from source (see [Fyne docs](https://docs.fyne.io/started/)) |

### Installing FFmpeg

```bash
# macOS
brew install ffmpeg

# Ubuntu / Debian
sudo apt install ffmpeg

# Windows (via Chocolatey)
choco install ffmpeg

# Windows (via Winget)
winget install ffmpeg
```

---

## Usage

### GUI mode (default)

Just double-click the binary (or run `./mediastream`). The window lets you:

1. **Browse** for any supported file — or drag and drop it onto the field
2. Set a **port** (default `8080`) and **frame rate** (default `30`)
3. Click **Start Streaming**
4. Copy the stream URL and paste it into OBS, VLC, or any browser

### CLI / headless mode

```bash
# Stream a JPEG on port 9000 at 24 FPS
./mediastream --headless --file /path/to/photo.jpg --port 9000

# Stream an MP4 video
./mediastream --headless --file /path/to/video.mp4 --port 8080

# All flags
./mediastream --help
```

---

## Stream URL

Once running, the stream is available at:

```
http://localhost:<port>/stream
```

### OBS Studio setup

1. Add a **Browser Source**
2. Set the URL to `http://localhost:8080/stream`
3. Match width/height to your source resolution
4. Click **OK**

### VLC

```bash
vlc http://localhost:8080/stream
```

### In a browser

```html
<img src="http://localhost:8080/stream" />
```

---

## Supported Formats

| Type | Extensions |
|---|---|
| Static image | `.jpg` `.jpeg` `.png` `.webp` `.bmp` |
| Animated GIF | `.gif` |
| Video | `.mp4` `.mkv` `.mov` `.avi` `.webm` `.flv` `.ts` `.m4v` |

> Any container/codec that FFmpeg can decode is supported for video. The list above is not exhaustive.

---

## Building from Source

```bash
# Install dependencies (Linux only — Fyne needs system graphics libs)
sudo apt install libgl1-mesa-dev xorg-dev libxkbcommon-dev

# Build for current platform
make build

# Cross-compile for all platforms
make release

# Run tests
make test
```

Binaries are placed in `dist/`.

---

## Architecture

```
cmd/mediastream/       Entry point — CLI flag parsing, GUI vs headless dispatch
internal/
  server/              HTTP server, MJPEG frame loop, /health endpoint
  media/               Source interface + per-format implementations
    media.go           Format detection and dispatcher
    image.go           Static image source (JPEG, PNG, WebP, BMP)
    gif.go             Animated GIF source — native per-frame delays
    video.go           FFmpeg-backed video source — any format, auto-loop
  gui/                 Fyne cross-platform window
```

The `media.Source` interface is intentionally simple:

```go
type Source interface {
    NextFrame() ([]byte, error)  // returns the next JPEG-encoded frame
    Close() error                // releases resources
}
```

Adding a new format is as easy as implementing those two methods and registering the extension in `media.go`.

---

## Endpoints

| Endpoint | Description |
|---|---|
| `GET /stream` | MJPEG stream — connect any compatible viewer here |
| `GET /health` | Returns `{"status":"ok","port":<n>}` — useful for health checks |

---

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/my-feature`)
3. Commit your changes (`git commit -m 'feat: add my feature'`)
4. Open a Pull Request

Run `make test` and `make lint` before submitting.

---

## License

MIT — see [LICENSE](LICENSE) for details.
