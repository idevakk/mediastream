package media

import (
	"bytes"
	"fmt"
	"image/gif"
	"image/jpeg"
	"os"
	"sync"
	"time"
)

// gifFrame holds a single decoded GIF frame and the delay before the next one.
type gifFrame struct {
	data  []byte        // JPEG-encoded bytes
	delay time.Duration // original GIF frame delay
}

// gifSource decodes all frames of an animated GIF upfront and loops them.
type gifSource struct {
	mu      sync.Mutex
	frames  []gifFrame
	index   int
	lastAt  time.Time
}

// newGIFSource opens path, decodes every frame to JPEG, and returns a gifSource.
// frameRate is used only as a fallback when a GIF frame has zero delay.
func newGIFSource(path string, frameRate int) (*gifSource, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening GIF %q: %w", path, err)
	}
	defer f.Close()

	g, err := gif.DecodeAll(f)
	if err != nil {
		return nil, fmt.Errorf("decoding GIF %q: %w", path, err)
	}
	if len(g.Image) == 0 {
		return nil, fmt.Errorf("GIF %q contains no frames", path)
	}

	fallbackDelay := time.Duration(float64(time.Second) / float64(frameRate))

	frames := make([]gifFrame, 0, len(g.Image))
	for i, img := range g.Image {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
			return nil, fmt.Errorf("encoding GIF frame %d: %w", i, err)
		}

		// GIF delays are in hundredths of a second
		delay := time.Duration(g.Delay[i]) * 10 * time.Millisecond
		if delay == 0 {
			delay = fallbackDelay
		}

		frames = append(frames, gifFrame{data: buf.Bytes(), delay: delay})
	}

	return &gifSource{frames: frames, lastAt: time.Now()}, nil
}

// NextFrame returns the current frame, advancing to the next one when the
// frame's delay has elapsed. This makes the GIF play back at its native speed
// regardless of how often NextFrame is called.
func (s *gifSource) NextFrame() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	frame := s.frames[s.index]
	if time.Since(s.lastAt) >= frame.delay {
		s.index = (s.index + 1) % len(s.frames)
		s.lastAt = time.Now()
	}
	return frame.data, nil
}

func (s *gifSource) Close() error { return nil }
