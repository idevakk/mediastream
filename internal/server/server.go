// Package server implements the HTTP MJPEG streaming server.
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/idevakk/mediastream/internal/media"
)

// Config holds all configuration needed to start a stream.
type Config struct {
	FilePath string
	Port     int
	// FrameRate is the target frames-per-second for the stream.
	// Defaults to 30 if zero.
	FrameRate int
}

// Server manages the HTTP server and the active media source.
type Server struct {
	cfg     Config
	source  media.Source
	httpSrv *http.Server
	mu      sync.RWMutex
	started bool
	cancel  context.CancelFunc
}

// New creates and validates a new Server from the given Config.
// It detects the media type from the file path and prepares the source.
func New(cfg Config) (*Server, error) {
	if cfg.FrameRate == 0 {
		cfg.FrameRate = 30
	}

	src, err := media.Open(cfg.FilePath, cfg.FrameRate)
	if err != nil {
		return nil, fmt.Errorf("opening media: %w", err)
	}

	return &Server{cfg: cfg, source: src}, nil
}

// Start begins serving the MJPEG stream. It blocks until the server
// is stopped via Stop or the context is cancelled.
func (s *Server) Start() error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return fmt.Errorf("server is already running")
	}
	s.started = true

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	mux := http.NewServeMux()
	mux.HandleFunc("/stream", s.handleStream)
	mux.HandleFunc("/health", s.handleHealth)

	s.httpSrv = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.Port),
		Handler: mux,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}
	s.mu.Unlock()

	return s.httpSrv.ListenAndServe()
}

// Stop gracefully shuts down the HTTP server and closes the media source.
func (s *Server) Stop() error {
	s.mu.Lock()

	if !s.started {
		s.mu.Unlock()
		return nil
	}
	s.started = false

	if s.cancel != nil {
		s.cancel()
	}

	httpSrv := s.httpSrv
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.source.Close(); err != nil {
		return fmt.Errorf("closing media source: %w", err)
	}
	return httpSrv.Shutdown(ctx)
}

// IsRunning reports whether the server is currently active.
func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.started
}

// StreamURL returns the full URL of the MJPEG stream endpoint.
func (s *Server) StreamURL() string {
	return fmt.Sprintf("http://localhost:%d/stream", s.cfg.Port)
}

// handleStream is the HTTP handler that outputs an MJPEG stream.
func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=mjpegframe")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported by this client", http.StatusInternalServerError)
		return
	}

	interval := time.Duration(float64(time.Second) / float64(s.cfg.FrameRate))
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			frame, err := s.source.NextFrame()
			if err != nil {
				return
			}

			fmt.Fprintf(w, "--mjpegframe\r\n")
			fmt.Fprintf(w, "Content-Type: image/jpeg\r\n")
			fmt.Fprintf(w, "Content-Length: %d\r\n\r\n", len(frame))
			w.Write(frame) //nolint:errcheck
			fmt.Fprintf(w, "\r\n")
			flusher.Flush()
		}
	}
}

// handleHealth returns a simple 200 OK for health checks.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","port":%d}`, s.cfg.Port)
}
