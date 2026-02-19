package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/idevakk/mediastream/internal/gui"
	"github.com/idevakk/mediastream/internal/server"
)

func main() {
	// CLI mode flags — if provided, skip GUI and run headless
	filePath := flag.String("file", "", "Path to image or video file to stream")
	port := flag.Int("port", 8080, "Port to serve the MJPEG stream on")
	headless := flag.Bool("headless", false, "Run without GUI (requires --file)")
	flag.Parse()

	if *headless {
		if *filePath == "" {
			fmt.Fprintln(os.Stderr, "error: --file is required in headless mode")
			flag.Usage()
			os.Exit(1)
		}
		cfg := server.Config{
			FilePath: *filePath,
			Port:     *port,
		}
		s, err := server.New(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Streaming %q on http://localhost:%d/stream\n", *filePath, *port)
		if err := s.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Default: launch the GUI
	gui.Run()
}
