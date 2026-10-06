package scene

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func waitFile(t *testing.T, p string) []byte {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if b, err := os.ReadFile(p); err == nil {
			return b
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("%s never appeared", p)
	return nil
}

func TestWatchActionsDone(t *testing.T) {
	dir := t.TempDir()
	stop := make(chan struct{})
	defer close(stop)
	got := make(chan string, 1)
	go watchActions(dir, func(n string) error { got <- n; return nil }, 5*time.Millisecond, stop)
	os.WriteFile(filepath.Join(dir, "0001.action"), []byte("load-sample\n"), 0o644)
	select {
	case n := <-got:
		if n != "load-sample" {
			t.Fatalf("name = %q", n)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("handler not called")
	}
	waitFile(t, filepath.Join(dir, "0001.done"))
}

func TestWatchActionsError(t *testing.T) {
	dir := t.TempDir()
	stop := make(chan struct{})
	defer close(stop)
	go watchActions(dir, func(n string) error { return errors.New(`unknown action "` + n + `"`) }, 5*time.Millisecond, stop)
	os.WriteFile(filepath.Join(dir, "0001.action"), []byte("nope"), 0o644)
	if b := waitFile(t, filepath.Join(dir, "0001.error")); !strings.Contains(string(b), `unknown action "nope"`) {
		t.Fatalf("error file = %q", b)
	}
}

func TestWatchActionsRunsEachOnce(t *testing.T) {
	dir := t.TempDir()
	stop := make(chan struct{})
	defer close(stop)
	var n atomic.Int32
	go watchActions(dir, func(string) error { n.Add(1); return nil }, 5*time.Millisecond, stop)
	os.WriteFile(filepath.Join(dir, "0001.action"), []byte("a"), 0o644)
	waitFile(t, filepath.Join(dir, "0001.done"))
	time.Sleep(50 * time.Millisecond)
	if n.Load() != 1 {
		t.Fatalf("handler ran %d times", n.Load())
	}
}

func TestOnActionWithoutEnvIsNoop(t *testing.T) {
	t.Setenv(EnvActions, "")
	OnAction(func(string) error { t.Error("handler must not run"); return nil })
}

func TestWatchActionsAnswersAtomically(t *testing.T) {
	dir := t.TempDir()
	stop := make(chan struct{})
	defer close(stop)
	go watchActions(dir, func(string) error { return errors.New("boom") }, 5*time.Millisecond, stop)
	os.WriteFile(filepath.Join(dir, "0001.action"), []byte("x"), 0o644)
	waitFile(t, filepath.Join(dir, "0001.error"))
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Fatalf("temp file left: %s", e.Name())
		}
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "0001.error")); string(b) != "boom" {
		t.Fatalf("error file = %q", b)
	}
}
