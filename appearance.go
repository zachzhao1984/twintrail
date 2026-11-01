package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"

	"github.com/egoist/mygo/ui"
)

const fileRowHeight float32 = 20

// palette is the colors of one appearance: the window and the bands of its
// panes, the sidebar beside them, and the accent that marks what is chosen.
// Both appearances share one accent, so a choice reads the same in either.
type palette struct {
	window       ui.Color // behind the rows of the panes
	sidebar      ui.Color // the sidebar's panel
	strip        ui.Color // the tabs' band, the header's and the footer's
	rowAlt       ui.Color // every other row of a pane
	hover        ui.Color // what the pointer rests on
	pressed      ui.Color // what holds a choice without the focus
	border       ui.Color // the lines between bands and columns
	text         ui.Color
	muted        ui.Color
	accent       ui.Color // chosen rows, the sidebar's icons and outlines
	accentText   ui.Color
	mark         ui.Color // the name of an entry the user marked
	terminal     ui.Color // the terminal's panel
	terminalText ui.Color
	fileIcon     ui.Color // where a file's own icon cannot be read
	dirIcon      ui.Color
	scrim        ui.Color // behind a dialog
}

// darkPalette is the graphite appearance: cool greys under one blue accent.
var darkPalette = palette{
	window:       ui.Hex("#1b1f25"),
	sidebar:      ui.Hex("#15181d"),
	strip:        ui.Hex("#20252c"),
	rowAlt:       ui.Hex("#1f242a"),
	hover:        ui.Hex("#262c34"),
	pressed:      ui.Hex("#2f3742"),
	border:       ui.Hex("#2c323b"),
	text:         ui.Hex("#e3e7ec"),
	muted:        ui.Hex("#98a1ad"),
	accent:       ui.Hex("#3f6fae"),
	accentText:   ui.Hex("#f2f7ff"),
	mark:         ui.Hex("#e6c17c"),
	terminal:     ui.Hex("#12161b"),
	terminalText: ui.Hex("#dbe2ea"),
	fileIcon:     ui.Hex("#cfd8e3"),
	dirIcon:      ui.Hex("#5aa9e6"),
	scrim:        ui.Hex("#00000066"),
}

// lightPalette is the paper appearance: white panes, cool greys, the same
// accent, and a warmer amber for what the user marked.
var lightPalette = palette{
	window:       ui.Hex("#ffffff"),
	sidebar:      ui.Hex("#f1f2f4"),
	strip:        ui.Hex("#f7f8fa"),
	rowAlt:       ui.Hex("#f6f7f9"),
	hover:        ui.Hex("#e8eaee"),
	pressed:      ui.Hex("#dcdfe5"),
	border:       ui.Hex("#dee1e6"),
	text:         ui.Hex("#23272e"),
	muted:        ui.Hex("#6b7280"),
	accent:       ui.Hex("#2f6fd0"),
	accentText:   ui.Hex("#ffffff"),
	mark:         ui.Hex("#a4600a"),
	terminal:     ui.Hex("#1b2026"),
	terminalText: ui.Hex("#e3e7ec"),
	fileIcon:     ui.Hex("#8a94a3"),
	dirIcon:      ui.Hex("#3d8fd6"),
	scrim:        ui.Hex("#00000044"),
}

var folderIcons = loadSystemIcons()

// cssColor returns c as the "#rrggbb" the window takes before its first frame.
func cssColor(c ui.Color) string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// colors picks the colors of the appearance the user chose.
func (a *app) colors() palette {
	if a.light {
		return lightPalette
	}
	return darkPalette
}

func martaTheme(light bool) *ui.Theme {
	p, t := darkPalette, *ui.DarkTheme()
	if light {
		p, t = lightPalette, *ui.LightTheme()
	}
	t.Background = p.window
	t.Surface = p.strip
	t.SurfaceHover = p.hover
	t.SurfacePressed = p.pressed
	t.Border = p.border
	t.Text = p.text
	t.TextMuted = p.muted
	t.Accent = p.accent
	t.AccentText = p.accentText
	t.Focus = ui.Transparent
	t.FontSize = 12
	t.Spacing = 3
	t.Radius = 2
	return &t
}

// Read the small Retina PNG representations from macOS' own icon files.
// Icons are decoded once, outside view construction, and aren't bundled.
func loadSystemIcons() map[string]*ui.Bitmap {
	out := map[string]*ui.Bitmap{}
	for name, file := range map[string]string{"folder": "GenericFolderIcon", "file": "GenericDocumentIcon", "Applications": "ApplicationsFolderIcon", "Desktop": "DesktopFolderIcon", "Developer": "DeveloperFolderIcon", "Documents": "DocumentsFolderIcon", "Downloads": "DownloadsFolder", "Movies": "MovieFolderIcon", "Music": "MusicFolderIcon", "Pictures": "PicturesFolderIcon", "Public": "PublicFolderIcon"} {
		data, err := os.ReadFile("/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/" + file + ".icns")
		if err != nil || len(data) < 8 || string(data[:4]) != "icns" {
			continue
		}
		var smallest *ui.Bitmap
		for at := 8; at+8 <= len(data); {
			size := int(binary.BigEndian.Uint32(data[at+4 : at+8]))
			if size < 8 || at+size > len(data) {
				break
			}
			chunk := data[at+8 : at+size]
			if bytes.HasPrefix(chunk, []byte("\x89PNG\r\n\x1a\n")) {
				bitmap, e := ui.DecodeBitmap(chunk)
				if e == nil {
					w, _ := bitmap.Size()
					old := 100000
					if smallest != nil {
						old, _ = smallest.Size()
					}
					if w >= 32 && w < old {
						smallest = bitmap
					}
				}
			}
			at += size
		}
		if smallest != nil {
			out[name] = smallest
		}
	}
	return out
}
func (a *app) fileIcon(c *ui.Context, e entry) {
	key := "file"
	if e.dir {
		key = "folder"
		if _, ok := folderIcons[e.name]; ok {
			key = e.name
		}
	}
	if icon := folderIcons[key]; icon != nil {
		ui.Image(c, icon).Width(16).Height(16).Fit(ui.Contain).Shrink(0)
	} else {
		colors := a.colors()
		ui.Box(c).Width(16).Height(16).Draw(func(p *ui.Painter, r ui.Rect) {
			color := colors.fileIcon
			if e.dir {
				color = colors.dirIcon
			}
			p.Fill(ui.Rect{X: r.X + 1, Y: r.Y + 3, W: 14, H: 12}, color, 1)
		})
	}
}
func modifiedText(t time.Time) string {
	now := time.Now()
	if t.Year() == now.Year() && t.YearDay() == now.YearDay() {
		return "Today at " + t.Format("15:04")
	}
	yesterday := now.AddDate(0, 0, -1)
	if t.Year() == yesterday.Year() && t.YearDay() == yesterday.YearDay() {
		return "Yesterday at " + t.Format("15:04")
	}
	return t.Format("2 Jan 2006 at 15:04")
}
func martaSize(n int64) string {
	if n < 1024 {
		return fmt.Sprintf("%d bytes", n)
	}
	for _, unit := range []struct {
		n    int64
		name string
	}{{1 << 30, "GB"}, {1 << 20, "MB"}, {1 << 10, "KB"}} {
		if n >= unit.n {
			return fmt.Sprintf("%.0f %s", float64(n)/float64(unit.n), unit.name)
		}
	}
	return ""
}
func tabSummary(t *tab) string {
	files, folders := 0, 0
	var total int64
	for _, e := range t.entries {
		if e.name == ".." {
			continue
		}
		if e.dir {
			folders++
		} else {
			files++
			total += e.size
		}
	}
	text := fmt.Sprintf("%d folders", folders)
	if files > 0 {
		text = fmt.Sprintf("%s in %d files, also %d folders", martaSize(total), files, folders)
	}
	selected := 0
	for _, v := range t.marked {
		if v {
			selected++
		}
	}
	if selected > 0 {
		text += fmt.Sprintf(" · %d selected", selected)
	}
	if t.disk != "" {
		text += "  |  " + t.disk
	}
	if t.loading {
		return "Loading…"
	}
	return text
}
func (a *app) edit() {
	paths := a.targets()
	if len(paths) > 0 {
		a.launch("open", append([]string{"-a", "TextEdit"}, paths...)...)
	}
}
