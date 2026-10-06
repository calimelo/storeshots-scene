package scene

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFromEnvAbsent(t *testing.T) {
	t.Setenv(EnvScene, "")
	if _, ok := FromEnv(); ok {
		t.Fatal("FromEnv() ok = true without STORESHOTS_SCENE")
	}
}

func TestFromEnvFull(t *testing.T) {
	t.Setenv(EnvScene, "settings")
	t.Setenv(EnvLang, "tr")
	t.Setenv(EnvSize, "1440x900")
	s, ok := FromEnv()
	want := Scene{Name: "settings", Lang: "tr", Width: 1440, Height: 900}
	if !ok || s != want {
		t.Fatalf("FromEnv() = %+v, %v; want %+v, true", s, ok, want)
	}
}

func TestFromEnvBadSizeIgnored(t *testing.T) {
	t.Setenv(EnvScene, "settings")
	t.Setenv(EnvSize, "big")
	s, ok := FromEnv()
	if !ok || s.Width != 0 || s.Height != 0 {
		t.Fatalf("FromEnv() = %+v, %v; want zero size, true", s, ok)
	}
}

func TestParseSize(t *testing.T) {
	for _, tc := range []struct {
		in   string
		w, h int
		ok   bool
	}{
		{"1440x900", 1440, 900, true},
		{" 1920X1080 ", 1920, 1080, true},
		{"1440", 0, 0, false},
		{"0x900", 0, 0, false},
		{"-1x2", 0, 0, false},
		{"axb", 0, 0, false},
		{"", 0, 0, false},
	} {
		w, h, err := ParseSize(tc.in)
		if (err == nil) != tc.ok || w != tc.w || h != tc.h {
			t.Errorf("ParseSize(%q) = %d, %d, %v", tc.in, w, h, err)
		}
	}
	if got := FormatSize(1440, 900); got != "1440x900" {
		t.Errorf("FormatSize = %q", got)
	}
}

func TestReadyWritesFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "ready")
	t.Setenv(EnvReady, p)
	if err := Ready(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("ready file not written: %v", err)
	}
}

func TestReadyWithoutEnvIsNoop(t *testing.T) {
	t.Setenv(EnvReady, "")
	if err := Ready(); err != nil {
		t.Fatal(err)
	}
}
