//go:build !darwin

package main

// initPreviewApp gives the app its Space toggle; the Quick Look panel is the
// Mac's, so other platforms preview with the fallback the app started with —
// none for now, and Space does nothing.
func initPreviewApp(a *app) {
	a.previewToggle = func(int) {}
}
