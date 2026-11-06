package main

// Run it to refresh the Settings shots the design notes point at:
//   go test -run TestCaptureSettings -count=1 .
import (
	"image"
	"image/draw"
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestCaptureSettings(t *testing.T) {
	const dir = "design/settings/shots"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	shots := []struct {
		name  string
		light bool
		area  int
		w, h  int
	}{
		{"settings-dark-files", false, 3, 1240, 760},
		{"settings-dark-terminal", false, 8, 1240, 760},
		{"settings-light-files", true, 3, 1240, 760},
		{"settings-dark-breadcrumbs", false, 1, 1240, 760},
		{"settings-minimum", false, 3, 960, 420},
	}
	for _, shot := range shots {
		a := fixture(t)
		a.settingsWriter.save = func(session) {}
		a.light = shot.light
		a.fonts.Files.Size = 20
		tt := ui.NewTester(a.view, shot.w, shot.h)
		tt.Key(ui.Cmd, ui.KeyComma)
		if err := tt.Click("字体区域：" + areaTitle(shot.area)); err != nil {
			t.Fatal(err)
		}
		tt.Advance(300 * time.Millisecond)
		panel, ok := tt.Find("设置面板")
		if !ok {
			t.Fatal("no settings panel")
		}
		rect := image.Rect(int(panel.X), int(panel.Y), int(panel.X+panel.W), int(panel.Y+panel.H))
		img := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
		draw.Draw(img, img.Bounds(), tt.Image(), rect.Min, draw.Src)
		file, err := os.Create(dir + "/" + shot.name + ".png")
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(file, img); err != nil {
			t.Fatal(err)
		}
		file.Close()
	}
}

func areaTitle(fontIndex int) string {
	for _, area := range preferenceAreas {
		if area.fontIndex == fontIndex {
			return area.title
		}
	}
	return ""
}
