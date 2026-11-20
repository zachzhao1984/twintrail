package main

// Renders whole-window shots from the headless tester, so looking at the
// window needs no window: nothing is launched, no window takes the focus and
// nothing is clicked outside this process. Refresh them with:
//
//	go test -run TestCaptureWindow -count=1 .
import (
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestCaptureWindow(t *testing.T) {
	const dir = "design/shots"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	shots := []struct {
		name  string
		light bool
		w, h  int
		setup func(*testing.T, *app, *ui.Tester)
	}{
		{"window-light", true, 1240, 760, nil},
		{"window-dark", false, 1240, 760, nil},
		{"window-dark-terminal", false, 1240, 760, func(t *testing.T, a *app, tt *ui.Tester) {
			tt.Key(ui.Cmd, ui.KeyBackquote)
		}},
	}
	for _, shot := range shots {
		a := fixture(t)
		a.settingsWriter.save = func(session) {}
		a.light = shot.light
		tt := ui.NewTester(a.view, shot.w, shot.h)
		// Two pixels a point, as the display draws the window.
		tt.SetScale(2)
		tt.Frame()
		if shot.setup != nil {
			shot.setup(t, a, tt)
		}
		tt.Advance(300 * time.Millisecond)
		img := tt.Image()
		file, err := os.Create(dir + "/" + shot.name + ".png")
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(file, img); err != nil {
			file.Close()
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
