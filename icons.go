package main

import (
	"embed"

	"github.com/egoist/mygo/ui"
)

// The app's own icons come from Tabler Icons, MIT licensed, with the licence
// kept beside them at icons/LICENSE-tabler-icons.txt. They are the library's
// outline drawings, unchanged: layout-columns for a workspace, plus for a new
// one, home, star, folder, cloud, server for a mounted volume, folder-down for Downloads, file-text for
// Documents, device-desktop for Desktop, layout-sidebar
// for the fold button, list and columns-2 for the two views, chevron-right for
// a section's fold, and search for the jump box. MyGo draws an icon as the
// alpha mask of its artwork, tinted by the row or the button it sits in, so
// the library's currentColor strokes take the colour of whatever carries
// them in either appearance.
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
