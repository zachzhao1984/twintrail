package main

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// The system preview is the Quick Look panel, which Finder's Space opens and
// closes, whose arrows move the preview along the folder. The panel is
// QLPreviewPanel, which asks a controller for the entries to preview: the app
// registers one, holding the pane's folder and the cursor the preview opened
// at, and the panel reads the entries through its data source.
var previewNative struct {
	once sync.Once
	err  error

	class objc.Class
}

var (
	selAcceptsControl = objc.RegisterName("acceptsPreviewPanelControl:")
	selBeginControl   = objc.RegisterName("beginPreviewPanelControl:")
	selEndControl     = objc.RegisterName("endPreviewPanelControl:")
	selSetCurrent     = objc.RegisterName("setCurrentPreviewItemIndex:")
)

// previewState is the controller's Go side: the folder the panel previews,
// the entries it holds, and the entry the cursor is on.
var previewController struct {
	mu      sync.Mutex
	panel   objc.ID
	entries []string
	at      int
	folder  string
}

// previewRect is an NSRect as the runtime passes it.
type previewRect struct{ X, Y, W, H float64 }

func previewSend(object objc.ID, selector objc.SEL, args ...any) objc.ID {
	return object.Send(selector, args...)
}

func previewString(s string) objc.ID {
	return previewSend(objc.ID(objc.GetClass("NSString")), objc.RegisterName("stringWithUTF8String:"), s)
}

// Keep the panel on the OS thread where it was created, in a pool, as the
// window server asks of the AppKit calls.
func previewPool() func() {
	runtime.LockOSThread()
	pool := previewSend(objc.ID(objc.GetClass("NSAutoreleasePool")), objc.RegisterName("new"))
	return func() { previewSend(pool, objc.RegisterName("drain")); runtime.UnlockOSThread() }
}

func initPreviewNative() error {
	previewNative.once.Do(func() {
		if _, previewNative.err = purego.Dlopen("/System/Library/Frameworks/Quartz.framework/Quartz", purego.RTLD_NOW|purego.RTLD_GLOBAL); previewNative.err != nil {
			return
		}
		// The controller: the panel's data source and delegate, whose Go
		// functions read the previewController above.
		previewNative.class, previewNative.err = objc.RegisterClass(
			"TwintrailPreviewController",
			objc.GetClass("NSObject"),
			[]*objc.Protocol{
				objc.GetProtocol("QLPreviewPanelDataSource"),
				objc.GetProtocol("QLPreviewPanelDelegate"),
			},
			nil,
			[]objc.MethodDef{
				{Cmd: objc.RegisterName("acceptsPreviewPanelControl:"), Fn: func(self objc.ID, cmd objc.SEL) bool { return true }},
				{Cmd: objc.RegisterName("beginPreviewPanelControl:"), Fn: func(self objc.ID, cmd objc.SEL) {}},
				{Cmd: objc.RegisterName("endPreviewPanelControl:"), Fn: func(self objc.ID, cmd objc.SEL) {}},
				{Cmd: objc.RegisterName("numberOfPreviewItemsInPreviewPanel:"), Fn: func(self objc.ID, cmd objc.SEL) int {
					previewController.mu.Lock()
					defer previewController.mu.Unlock()
					return len(previewController.entries)
				}},
				{Cmd: objc.RegisterName("previewPanel:previewItemAtIndex:"), Fn: func(self objc.ID, cmd objc.SEL, panel objc.ID, index int) objc.ID {
					previewController.mu.Lock()
					defer previewController.mu.Unlock()
					if index < 0 || index >= len(previewController.entries) {
						return 0
					}
					return previewItemOf(previewController.entries[index])
				}},
			},
		)
		if previewNative.err != nil {
			return
		}
	})
	return previewNative.err
}

// previewItem wraps one entry as a QLPreviewItem: the URL of the entry,
// whose title the panel takes from the file name.
func previewItemOf(path string) objc.ID {
	return previewSend(objc.ID(objc.GetClass("NSURL")), objc.RegisterName("fileURLWithPath:"), previewString(path))
}

// previewPanel opens the Quick Look panel over the pane's folder at the
// entry, or reloads it with what the pane holds now.
func (a *app) previewPanel(index int) error {
	if err := initPreviewNative(); err != nil {
		return err
	}
	drain := previewPool()
	defer drain()

	p := &a.panes[index]
	t := p.tabs[p.current]
	paths := a.targets()
	if len(paths) == 0 {
		return nil
	}

	panel := previewSend(objc.ID(objc.GetClass("QLPreviewPanel")), objc.RegisterName("sharedPreviewPanel"))
	if panel == 0 {
		return fmt.Errorf("preview: Quick Look is unavailable")
	}

	controller := previewSend(objc.ID(previewNative.class), objc.RegisterName("alloc"))
	controller = previewSend(controller, objc.RegisterName("init"))
	previewSend(panel, objc.RegisterName("setDataSource:"), controller)
	previewSend(panel, objc.RegisterName("setDelegate:"), controller)

	previewController.mu.Lock()
	previewController.entries = paths
	at := 0
	for i, e := range t.visible() {
		if e.name != ".." && filepath.Join(t.path, e.name) == paths[0] {
			at = i
		}
	}
	previewController.at = at
	previewController.folder = t.path
	previewController.mu.Unlock()

	previewSend(panel, objc.RegisterName("reloadData"))
	previewSend(panel, objc.RegisterName("makeKeyAndOrderFront:"), 0)
	return nil
}

// previewPanelClose closes the panel, as another press of Space does.
func previewPanelClose() {
	drain := previewPool()
	defer drain()
	previewSend(previewSend(objc.ID(objc.GetClass("QLPreviewPanel")), objc.RegisterName("sharedPreviewPanel")), objc.RegisterName("close"))
}

// previewPanelVisible tells whether the panel shows.
func previewPanelVisible() bool {
	drain := previewPool()
	defer drain()
	return previewSend(previewSend(objc.ID(objc.GetClass("QLPreviewPanel")), objc.RegisterName("sharedPreviewPanel")), objc.RegisterName("isVisible")) != 0
}

// initPreviewApp makes the app's Space toggle the Quick Look panel.
func initPreviewApp(a *app) {
	a.previewToggle = func(index int) {
		if previewPanelVisible() {
			previewPanelClose()
			return
		}
		if err := a.previewPanel(index); err != nil {
			a.status = err.Error()
		}
	}
}
