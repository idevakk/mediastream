# ============================================================
#  MediaStream — Makefile
# ============================================================
#
#  Common targets:
#    make run          — build and run the GUI locally
#    make build        — build for the current OS
#    make release      — build for Windows, macOS, and Linux
#    make test         — run all tests
#    make lint         — run golangci-lint
#    make clean        — remove build artifacts
#
# ============================================================

BINARY    := mediastream
PKG       := ./cmd/mediastream
DIST      := dist
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -ldflags "-s -w -X main.Version=$(VERSION)"

.PHONY: run build release test lint clean

# ── Development ─────────────────────────────────────────────

run:
	go run $(PKG)

# ── Single-platform build ────────────────────────────────────

build:
	go build $(LDFLAGS) -o $(DIST)/$(BINARY) $(PKG)

# ── Cross-platform release builds ───────────────────────────

release: release-linux release-darwin release-windows

release-linux:
	GOOS=linux GOARCH=amd64 \
	go build $(LDFLAGS) -o $(DIST)/$(BINARY)-linux-amd64 $(PKG)

	GOOS=linux GOARCH=arm64 \
	go build $(LDFLAGS) -o $(DIST)/$(BINARY)-linux-arm64 $(PKG)

release-darwin:
	GOOS=darwin GOARCH=amd64 \
	go build $(LDFLAGS) -o $(DIST)/$(BINARY)-darwin-amd64 $(PKG)

	GOOS=darwin GOARCH=arm64 \
	go build $(LDFLAGS) -o $(DIST)/$(BINARY)-darwin-arm64 $(PKG)

release-windows:
	# -H=windowsgui hides the console window on Windows
	GOOS=windows GOARCH=amd64 \
	go build $(LDFLAGS) -H=windowsgui -o $(DIST)/$(BINARY)-windows-amd64.exe $(PKG)

# ── Quality ──────────────────────────────────────────────────

test:
	go test ./... -race -coverprofile=coverage.out
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...

# ── Housekeeping ─────────────────────────────────────────────

clean:
	rm -rf $(DIST) coverage.out
