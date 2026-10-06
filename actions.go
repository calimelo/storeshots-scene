package scene

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// EnvActions names the directory storeshots uses to send video actions.
const EnvActions = "STORESHOTS_ACTIONS"

// OnAction runs fn for every named action storeshots sends while recording a
// video (scenario steps like {"action": "load-sample"}). fn runs on a
// background goroutine: UI changes must go through the toolkit's UI thread
// (fyne.Do in Fyne). Return an error for unknown actions; the video then fails
// with that message. OnAction is a no-op when the app was not launched by
// storeshots video.
func OnAction(fn func(name string) error) {
	dir := os.Getenv(EnvActions)
	if dir == "" {
		return
	}
	go watchActions(dir, fn, 50*time.Millisecond, nil)
}

// watchActions implements the protocol: storeshots writes <n>.action
// (atomically, via rename) containing the action name; the app answers with
// <n>.done or <n>.error holding the error message.
func watchActions(dir string, fn func(string) error, every time.Duration, stop <-chan struct{}) {
	seen := map[string]bool{}
	tick := time.NewTicker(every)
	defer tick.Stop()
	for {
		entries, _ := os.ReadDir(dir)
		var names []string
		for _, e := range entries {
			if n := e.Name(); strings.HasSuffix(n, ".action") && !seen[n] {
				names = append(names, n)
			}
		}
		sort.Strings(names)
		for _, n := range names {
			seen[n] = true
			base := filepath.Join(dir, strings.TrimSuffix(n, ".action"))
			b, err := os.ReadFile(filepath.Join(dir, n))
			if err == nil {
				err = fn(strings.TrimSpace(string(b)))
			}
			if err != nil {
				os.WriteFile(base+".error", []byte(err.Error()), 0o644)
				continue
			}
			os.WriteFile(base+".done", nil, 0o644)
		}
		select {
		case <-stop:
			return
		case <-tick.C:
		}
	}
}
