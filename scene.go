// Package scene implements the app side of the storeshots screenshot protocol.
//
// storeshots launches the app with STORESHOTS_* environment variables. The app
// reads them with FromEnv, opens the requested screen with demo data, sizes its
// window, and calls Ready once the screen is fully rendered.
package scene

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Environment variables set by storeshots when it launches the app.
const (
	EnvScene = "STORESHOTS_SCENE" // scene name, e.g. "settings"
	EnvLang  = "STORESHOTS_LANG"  // language code, e.g. "tr"
	EnvSize  = "STORESHOTS_SIZE"  // logical window size, e.g. "1440x900"
	EnvReady = "STORESHOTS_READY" // file to create when the scene is ready
)

// Scene is the screen storeshots asked the app to show.
type Scene struct {
	Name   string
	Lang   string
	Width  int // logical (DPI-independent) window width; 0 if not set
	Height int // logical window height; 0 if not set
}

// FromEnv reports the requested scene. ok is false when the app was not
// launched by storeshots.
func FromEnv() (s Scene, ok bool) {
	name := os.Getenv(EnvScene)
	if name == "" {
		return Scene{}, false
	}
	s = Scene{Name: name, Lang: os.Getenv(EnvLang)}
	if w, h, err := ParseSize(os.Getenv(EnvSize)); err == nil {
		s.Width, s.Height = w, h
	}
	return s, true
}

// Ready tells storeshots the scene is rendered and can be captured.
// It is a no-op when the app was not launched by storeshots.
func Ready() error {
	p := os.Getenv(EnvReady)
	if p == "" {
		return nil
	}
	return os.WriteFile(p, []byte("ready\n"), 0o644)
}

// ParseSize parses "<width>x<height>", e.g. "1440x900".
func ParseSize(v string) (w, h int, err error) {
	ws, hs, found := strings.Cut(strings.ToLower(strings.TrimSpace(v)), "x")
	if found {
		w, err = strconv.Atoi(ws)
		if err == nil {
			h, err = strconv.Atoi(hs)
		}
	}
	if !found || err != nil || w <= 0 || h <= 0 {
		return 0, 0, fmt.Errorf("invalid size %q (want WIDTHxHEIGHT, e.g. 1440x900)", v)
	}
	return w, h, nil
}

// FormatSize formats a size as "<width>x<height>".
func FormatSize(w, h int) string {
	return strconv.Itoa(w) + "x" + strconv.Itoa(h)
}
