package media

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// videoSource pipes frames from an FFmpeg subprocess as raw JPEG images.
// It works with any container/codec that FFmpeg supports, and loops automatically.
type videoSource struct {
	mu        sync.Mutex
	path      string
	frameRate int
	cmd       *exec.Cmd
	stdout    io.ReadCloser
	buf       bytes.Buffer
}

// newVideoSource verifies that FFmpeg is available, then spawns the decoding
// subprocess. FFmpeg outputs one JPEG per frame separated by JPEG EOI markers.
func newVideoSource(path string, frameRate int) (*videoSource, error) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, fmt.Errorf(
			"ffmpeg not found in PATH — please install FFmpeg to stream video files: %w", err,
		)
	}

	s := &videoSource{path: path, frameRate: frameRate}
	if err := s.spawn(); err != nil {
		return nil, err
	}
	return s, nil
}

// spawn starts (or restarts) the FFmpeg process. Called on init and on loop.
func (s *videoSource) spawn() error {
	// -stream_loop -1 tells FFmpeg to loop the input indefinitely.
	// image2pipe + mjpeg output gives us a raw stream of back-to-back JPEGs.
	cmd := exec.Command("ffmpeg",
		"-stream_loop", "-1",
		"-re",                   // read at native frame rate
		"-i", s.path,
		"-vf", fmt.Sprintf("fps=%d", s.frameRate),
		"-q:v", "3",             // JPEG quality (2=best, 31=worst)
		"-f", "image2pipe",
		"-vcodec", "mjpeg",
		"-",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("creating ffmpeg stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting ffmpeg: %w", err)
	}

	s.cmd = cmd
	s.stdout = stdout
	return nil
}

// NextFrame reads the next JPEG frame from the FFmpeg output pipe.
// It scans for JPEG SOI (0xFFD8) and EOI (0xFFD9) markers to extract
// each complete frame from the continuous byte stream.
func (s *videoSource) NextFrame() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Read until we find a complete JPEG (SOI...EOI).
	for {
		b, err := readByte(s.stdout)
		if err != nil {
			return nil, fmt.Errorf("reading from ffmpeg: %w", err)
		}
		// Look for JPEG SOI marker: 0xFF 0xD8
		if b != 0xFF {
			continue
		}
		b2, err := readByte(s.stdout)
		if err != nil {
			return nil, err
		}
		if b2 != 0xD8 {
			continue
		}

		// We're at the start of a JPEG. Read until EOI (0xFF 0xD9).
		frame := []byte{0xFF, 0xD8}
		for {
			chunk := make([]byte, 4096)
			n, err := s.stdout.Read(chunk)
			if n > 0 {
				frame = append(frame, chunk[:n]...)
				// Check if EOI appeared in this chunk
				if idx := bytes.Index(frame, []byte{0xFF, 0xD9}); idx >= 0 {
					return frame[:idx+2], nil
				}
			}
			if err != nil {
				return nil, fmt.Errorf("reading ffmpeg frame body: %w", err)
			}
		}
	}
}

// Close terminates the FFmpeg subprocess and closes the pipe.
func (s *videoSource) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stdout != nil {
		s.stdout.Close()
	}
	if s.cmd != nil && s.cmd.Process != nil {
		return s.cmd.Process.Kill()
	}
	return nil
}

// readByte is a helper that reads exactly one byte from r.
func readByte(r io.Reader) (byte, error) {
	var b [1]byte
	_, err := io.ReadFull(r, b[:])
	// suppress "unused import" for binary — it's available for future use
	_ = binary.LittleEndian
	return b[0], err
}
