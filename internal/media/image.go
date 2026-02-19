package media

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"  // register PNG decoder
	_ "golang.org/x/image/webp" // register WebP decoder
	_ "golang.org/x/image/bmp"  // register BMP decoder
	"os"
	"sync"
)

// imageSource streams a single static image indefinitely.
type imageSource struct {
	mu    sync.Mutex
	frame []byte // pre-encoded JPEG bytes
}

// newImageSource reads and decodes the image at path, then re-encodes it
// as a JPEG so that downstream consumers always receive consistent bytes.
func newImageSource(path string) (*imageSource, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening image %q: %w", path, err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decoding image %q: %w", path, err)
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 92}); err != nil {
		return nil, fmt.Errorf("re-encoding image as JPEG: %w", err)
	}

	return &imageSource{frame: buf.Bytes()}, nil
}

func (s *imageSource) NextFrame() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.frame, nil
}

func (s *imageSource) Close() error { return nil }
