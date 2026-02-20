package media_test

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/idevakk/mediastream/internal/media"
)

func TestOpenUnsupportedExtension(t *testing.T) {
	_, err := media.Open("file.xyz", 30)
	if err == nil {
		t.Fatal("expected error for unsupported extension, got nil")
	}
}

func TestOpenMissingFile(t *testing.T) {
	_, err := media.Open("/nonexistent/path/image.jpg", 30)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// writeMinimalJPEG creates a valid JPEG in dir and returns the path.
func writeMinimalJPEG(t *testing.T) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.Black)

	path := filepath.Join(t.TempDir(), "test.jpg")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("creating test JPEG: %v", err)
	}
	defer f.Close()

	if err := jpeg.Encode(f, img, nil); err != nil {
		t.Fatalf("encoding test JPEG: %v", err)
	}
	return path
}

func TestImageSourceNextFrame(t *testing.T) {
	path := writeMinimalJPEG(t)

	src, err := media.Open(path, 30)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer src.Close()

	frame1, err := src.NextFrame()
	if err != nil {
		t.Fatalf("NextFrame: %v", err)
	}
	if len(frame1) < 2 {
		t.Fatal("frame too small")
	}
	// JPEG SOI marker
	if frame1[0] != 0xFF || frame1[1] != 0xD8 {
		t.Fatalf("expected JPEG SOI, got %X %X", frame1[0], frame1[1])
	}

	// Static image must return identical bytes every call
	frame2, _ := src.NextFrame()
	if len(frame1) != len(frame2) {
		t.Fatal("static image returned different frame sizes")
	}
}
