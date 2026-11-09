package main

import (
	"embed"

	"github.com/egoist/mygo/ui"
)

// The app's own icons are drawings of Kubai Icons (icons.kubai.dev, MIT, see
// icons/LICENSE-kubai-icons.txt): each one was taken from the library's
// still pose and reduced to what ui.Icon can draw in the colour of the text —
// the outline the library strokes, the table its white faces cut, and its
// small detail strokes. The library's fills, gradients and shadow are left
// out, so the icons take their colour from the sidebar, the strip or the row
// they sit in, in either appearance, as the app's own drawings did.
//
// They file under the sidebar's 24-DIP box, drawn from x 3 to 27 and y 0.5 to
// 24.5, so a pair of them lines up without the padding the library keeps for
// its shadow.
//
//go:embed icons/*.svg
var iconFiles embed.FS

// libraryIcon parses one of the drawings baked at icons/. The set is part of
// the program, so a name that does not read is a build to fix, not a runtime
// state to report.
func libraryIcon(name string) *ui.SVG {
	data, err := iconFiles.ReadFile("icons/" + name + ".svg")
	if err != nil {
		panic("icons: " + name + ": " + err.Error())
	}
	return ui.MustParseSVG(data)
}
