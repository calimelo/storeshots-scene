# storeshots-scene

App-side helper for [storeshots](https://github.com/calimelo/storeshots), which
captures Microsoft Store and Mac App Store screenshots of desktop apps.

```
go get github.com/calimelo/storeshots-scene
```

storeshots starts your app with these environment variables:

| Variable | Example | Meaning |
|---|---|---|
| `STORESHOTS_SCENE` | `settings` | screen to open |
| `STORESHOTS_LANG` | `tr` | UI language |
| `STORESHOTS_SIZE` | `1440x900` | logical window size |
| `STORESHOTS_READY` | `/tmp/…/ready` | file to create when the screen is rendered |

Your app opens the scene with demo data, sizes its window and calls `scene.Ready()`.

`STORESHOTS_SIZE` is the **content (client-area) size**, without title bar or borders —
storeshots captures only the content on Windows, and composes to 16:10 / 16:9 store sizes.

## Fyne

```go
import scene "github.com/calimelo/storeshots-scene"

a := app.New()
w := a.NewWindow("MyApp")
if s, ok := scene.FromEnv(); ok {
	setLanguage(s.Lang)
	openScene(w, s.Name) // load demo data, show the screen
	if s.Width > 0 {
		w.Resize(fyne.NewSize(float32(s.Width), float32(s.Height)))
	}
	a.Lifecycle().SetOnStarted(func() {
		go func() {
			time.Sleep(300 * time.Millisecond) // let the first frame render
			_ = scene.Ready()
		}()
	})
}
w.ShowAndRun()
```

## Wails v2

```go
import goruntime "runtime"

s, shot := scene.FromEnv()
opts := &options.App{
	Width: 1024, Height: 768,
	OnDomReady: func(ctx context.Context) {
		if shot {
			runtime.EventsEmit(ctx, "storeshots:scene", s.Name, s.Lang)
		}
	},
	Bind: []interface{}{app},
}
if shot && s.Width > 0 {
	opts.Width, opts.Height = s.Width, s.Height
	// On Windows, Wails' Width/Height include the window frame. A frameless window
	// makes the content exactly s.Width × s.Height (macOS keeps its real frame).
	if goruntime.GOOS == "windows" {
		opts.Frameless = true
	}
}

// Bound method; the frontend calls it after navigating to the scene.
func (a *App) ScreenshotReady() error { return scene.Ready() }
```

Frontend:

```js
EventsOn("storeshots:scene", async (name, lang) => {
  await i18n.changeLanguage(lang);
  await router.push("/" + name);
  requestAnimationFrame(() => window.go.main.App.ScreenshotReady());
});
```

## Video actions (v0.2.0)

`storeshots video` scenarios can send named actions to the app — for things
that are hard to click (load sample data, trigger an animation):

```go
scene.OnAction(func(name string) error {
	switch name {
	case "load-sample":
		fyne.Do(loadSample) // UI changes on the UI thread
		return nil
	}
	return fmt.Errorf("unknown action %q", name)
})
```

`OnAction` does nothing unless the app was started by `storeshots video`.
