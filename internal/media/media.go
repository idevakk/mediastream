// Package media provides a unified interface for reading JPEG frames
// from different media sources: static images, GIFs, and video files.
package media

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Source is the common interface for all media types.
// Implementations must be safe for concurrent calls to NextFrame.
type Source interface {
	// NextFrame returns the next JPEG-encoded frame.
	// For static images it always returns the same bytes.
	// For videos/GIFs it advances the playback position, looping at the end.
	NextFrame() ([]byte, error)

	// Close releases any resources (file handles, FFmpeg processes) held by the source.
	Close() error
}

// SupportedExtensions lists every file extension Open can handle.
var SupportedExtensions = []string{
	".jpg", ".jpeg", ".png", ".webp", ".bmp", // static images
	".gif",                               // animated GIF
	".mp4", ".mkv", ".mov", ".avi",       // common video containers
	".webm", ".flv", ".ts", ".m4v",       // additional video formats
}

// Open inspects the file extension and returns the appropriate Source.
// frameRate is only used for video sources; it is ignored for images.
func Open(path string, frameRate int) (Source, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".bmp":
		return newImageSource(path)
	case ".gif":
		return newGIFSource(path, frameRate)
	case ".mp4", ".mkv", ".mov", ".avi", ".webm", ".flv", ".ts", ".m4v":
		return newVideoSource(path, frameRate)
	default:
		return nil, fmt.Errorf(
			"unsupported file type %q — supported formats: %s",
			ext, strings.Join(SupportedExtensions, ", "),
		)
	}
}
