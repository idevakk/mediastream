// Package gui implements the cross-platform native window using the Fyne toolkit.
package gui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/idevakk/mediastream/internal/media"
	"github.com/idevakk/mediastream/internal/server"
)

// Run creates and displays the main application window. It blocks until
// the window is closed.
func Run() {
	a := app.NewWithID("com.idevakk.mediastream")
	a.SetIcon(theme.MediaVideoIcon()) // placeholder; swap for a real icon asset

	w := a.NewWindow("MediaStream")
	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(480, 320))

	ui := buildUI(w)
	w.SetContent(ui)
	w.CenterOnScreen()
	w.ShowAndRun()
}

// state holds mutable UI runtime state.
type state struct {
	srv      *server.Server
	filePath string
}

func buildUI(w fyne.Window) fyne.CanvasObject {
	st := &state{}

	// ── File selection ──────────────────────────────────────────────────────
	fileLabel := widget.NewLabel("No file selected")
	fileLabel.Truncation = fyne.TextTruncateEllipsis

	browseBtn := widget.NewButtonWithIcon("Browse…", theme.FolderOpenIcon(), func() {
		fd := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
			if err != nil || uc == nil {
				return
			}
			uc.Close()
			st.filePath = uc.URI().Path()
			fileLabel.SetText(uc.URI().Name())
		}, w)

		// Build the filter from supported extensions
		exts := make([]string, 0, len(media.SupportedExtensions))
		for _, e := range media.SupportedExtensions {
			exts = append(exts, strings.TrimPrefix(e, "."))
		}
		fd.SetFilter(storage.NewExtensionFileFilter(media.SupportedExtensions))
		fd.Show()
		_ = exts
	})

	fileRow := container.NewBorder(nil, nil, nil, browseBtn, fileLabel)

	// ── Port input ──────────────────────────────────────────────────────────
	portEntry := widget.NewEntry()
	portEntry.SetText("8080")
	portEntry.SetPlaceHolder("e.g. 8080")
	portEntry.Validator = func(s string) error {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 || n > 65535 {
			return fmt.Errorf("must be a number between 1 and 65535")
		}
		return nil
	}

	// ── Frame rate input ────────────────────────────────────────────────────
	fpsEntry := widget.NewEntry()
	fpsEntry.SetText("30")
	fpsEntry.SetPlaceHolder("e.g. 30")
	fpsEntry.Validator = func(s string) error {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 || n > 120 {
			return fmt.Errorf("must be a number between 1 and 120")
		}
		return nil
	}

	form := widget.NewForm(
		widget.NewFormItem("Port", portEntry),
		widget.NewFormItem("Frame Rate (FPS)", fpsEntry),
	)

	// ── Status & URL ────────────────────────────────────────────────────────
	statusLabel := widget.NewLabelWithStyle("Stopped", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	urlLabel := widget.NewHyperlink("", nil)
	urlLabel.Hidden = true

	// ── Start / Stop ────────────────────────────────────────────────────────
	var startBtn, stopBtn *widget.Button

	stopBtn = widget.NewButtonWithIcon("Stop", theme.MediaStopIcon(), func() {
		if st.srv == nil {
			return
		}
		if err := st.srv.Stop(); err != nil {
			dialog.ShowError(err, w)
		}
		st.srv = nil

		statusLabel.SetText("Stopped")
		statusLabel.TextStyle = fyne.TextStyle{Bold: true}
		urlLabel.Hidden = true
		startBtn.Enable()
		stopBtn.Disable()
	})
	stopBtn.Importance = widget.DangerImportance
	stopBtn.Disable()

	startBtn = widget.NewButtonWithIcon("Start Streaming", theme.MediaPlayIcon(), func() {
		if st.filePath == "" {
			dialog.ShowInformation("No File", "Please select an image or video file first.", w)
			return
		}
		if portEntry.Validate() != nil || fpsEntry.Validate() != nil {
			dialog.ShowInformation("Invalid Settings", "Please fix the highlighted fields.", w)
			return
		}

		port, _ := strconv.Atoi(portEntry.Text)
		fps, _ := strconv.Atoi(fpsEntry.Text)

		cfg := server.Config{
			FilePath:  st.filePath,
			Port:      port,
			FrameRate: fps,
		}

		srv, err := server.New(cfg)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to open file:\n%v", err), w)
			return
		}
		st.srv = srv

		go func() {
			if err := srv.Start(); err != nil && srv.IsRunning() {
				dialog.ShowError(fmt.Errorf("server error: %v", err), w)
			}
		}()

		streamURL := srv.StreamURL()
		u, _ := fyne.ParseURI("http://localhost:" + strconv.Itoa(port) + "/stream")
		urlLabel.SetText(streamURL)
		urlLabel.SetURL(u)
		urlLabel.Hidden = false

		statusLabel.SetText("● Streaming")
		statusLabel.TextStyle = fyne.TextStyle{Bold: true}

		startBtn.Disable()
		stopBtn.Enable()
	})
	startBtn.Importance = widget.HighImportance

	// ── Copy URL button ─────────────────────────────────────────────────────
	copyBtn := widget.NewButtonWithIcon("Copy URL", theme.ContentCopyIcon(), func() {
		if urlLabel.Hidden {
			return
		}
		w.Clipboard().SetContent(urlLabel.Text)
	})

	urlRow := container.NewBorder(nil, nil, nil, copyBtn, urlLabel)

	// ── Layout ──────────────────────────────────────────────────────────────
	content := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Media File", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		fileRow,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		form,
		widget.NewSeparator(),
		container.NewGridWithColumns(2, startBtn, stopBtn),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Stream URL", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		urlRow,
		statusLabel,
	)

	return container.NewPadded(content)
}
