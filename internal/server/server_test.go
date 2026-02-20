package server_test

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/idevakk/mediastream/internal/server"
)

// writeTestJPEG creates a minimal valid JPEG programmatically.
func writeTestJPEG(t *testing.T) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.Black)

	dir := t.TempDir()
	path := filepath.Join(dir, "test.jpg")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("writing test JPEG: %v", err)
	}
	defer f.Close()

	if err := jpeg.Encode(f, img, nil); err != nil {
		t.Fatalf("encoding test JPEG: %v", err)
	}
	return path
}

func TestServerStartStop(t *testing.T) {
	jpg := writeTestJPEG(t)
	cfg := server.Config{FilePath: jpg, Port: 19870}

	srv, err := server.New(cfg)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	go srv.Start()                    //nolint:errcheck
	time.Sleep(80 * time.Millisecond) // give the server time to bind

	if !srv.IsRunning() {
		t.Fatal("expected server to be running")
	}

	if err := srv.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}

func TestHealthEndpoint(t *testing.T) {
	jpg := writeTestJPEG(t)
	cfg := server.Config{FilePath: jpg, Port: 19871}

	srv, err := server.New(cfg)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}
	go srv.Start() //nolint:errcheck
	time.Sleep(80 * time.Millisecond)
	defer srv.Stop() //nolint:errcheck

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", cfg.Port))
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Fatal("expected non-empty health response")
	}
}

func TestStreamEndpointHeaders(t *testing.T) {
	jpg := writeTestJPEG(t)
	cfg := server.Config{FilePath: jpg, Port: 19872, FrameRate: 5}

	srv, err := server.New(cfg)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}
	go srv.Start() //nolint:errcheck
	time.Sleep(80 * time.Millisecond)
	defer srv.Stop() //nolint:errcheck

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/stream", cfg.Port))
	if err != nil {
		t.Fatalf("GET /stream: %v", err)
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		t.Fatal("expected Content-Type header")
	}
}
