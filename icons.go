package main

import (
	"embed"

	"github.com/egoist/mygo/ui"
)

// The app's own icons are drawn here, on one 16-point grid. Each is a solid
// shape with its detail cut out of it — the panes of a workspace, the door of
// the house, the inner star of a favourite, the slot of a disk — so the detail
// survives the way MyGo draws an icon: as the alpha mask of the artwork, tinted
// in one colour by the row or the button it sits in. A line drawing inside the
// shape would fill in; a cut-out stays open in either appearance.
//
// The shapes come from a round of work with Kubai Icons, where each one was
// drawn as a layered, animated object and checked at 24 and 28 points, in
// every palette, for size, centring, seams and an end to its motion that
// settles. That layered colour cannot cross into an icon mask, so what carries
// over here is the silhouette and the cut-out: bolder than the old hairlines,
// still one family beside the file list's system icons. The moving versions
// stay in the design notes at design/kubai-sidebar/.
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
