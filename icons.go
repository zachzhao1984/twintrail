package main

import (
	"embed"

	"github.com/egoist/mygo/ui"
)

// The app's own icons are drawn here, on one 16-point grid: a 1.5-point
// stroke, round caps and joins, no fill but a dot or two, and about 12 points
// of ink in the box, so the set reads as one family beside the file list's
// system icons and stays legible at the 12 to 16 points the window keeps for
// it. They take the colour of the row, the strip or the button they sit in,
// in either appearance.
//
// The workspace icon is the window the app opens: two panes side by side,
// which is also what tells it from the sidebar's own button, a window with a
// narrow band for the sidebar.
//
//go:embed icons/*.svg
var iconFiles embed.FS

// icon parses one of the drawings kept at icons/. The set is part of the
// program, so a name that does not read is a build to fix, not a runtime state
// to report.
func icon(name string) *ui.SVG {
	data, err := iconFiles.ReadFile("icons/" + name + ".svg")
	if err != nil {
		panic("icons: " + name + ": " + err.Error())
	}
	return ui.MustParseSVG(data)
}
