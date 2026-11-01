package main

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func fixture(t *testing.T) *app {
	t.Helper()
	left := t.TempDir()
	right := t.TempDir()
	if err := os.WriteFile(filepath.Join(left, "hello.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	os.Mkdir(filepath.Join(left, "Folder"), 0755)
	os.WriteFile(filepath.Join(left, ".hidden"), nil, 0644)
	a := newApp(left, right)
	for i := range a.panes {
		tab := a.panes[i].tabs[0]
		a.load(tab, tab.path)
	}
	return a
}
func TestNavigationAndView(t *testing.T) {
	a := fixture(t)
	tt := ui.NewTester(a.view, 1100, 720)
	if !tt.HasText("hello.txt") {
		t.Fatal(tt.Texts())
	}
	if tt.HasText(".hidden") {
		t.Fatal("hidden file visible")
	}
	t1 := a.current()
	t1.hidden = true
	a.load(t1, t1.path)
	tt.Frame()
	if !tt.HasText(".hidden") {
		t.Fatal("hidden file missing")
	}
	if err := tt.Click("F7 New Folder"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	a.modal.value = "Created"
	a.confirm()
	tt.Frame()
	if !tt.HasText("Created") {
		t.Fatal("new folder missing")
	}
	for i := 0; i < 8; i++ {
		a.ask("Actions")
		tt.Frame()
		a.modal = nil
		tt.Frame()
	}
}
func TestTransfer(t *testing.T) {
	a := fixture(t)
	src := filepath.Join(a.current().path, "hello.txt")
	right := a.panes[1].tabs[0].path
	if err := perform(&dialog{kind: "Copy", sources: []string{src}, value: right}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(right, "hello.txt"))
	if err != nil || string(data) != "hello" {
		t.Fatalf("copy %s %v", data, err)
	}
	if err := perform(&dialog{kind: "Copy", sources: []string{src}, value: right}); err == nil {
		t.Fatal("overwrite allowed")
	}
	if err := validateDestination(a.current().path, filepath.Join(a.current().path, "Folder", "nested")); err == nil {
		t.Fatal("recursive destination allowed")
	}
	if err := perform(&dialog{kind: "Rename", sources: []string{src}, target: a.current().path, value: "renamed.txt"}); err != nil {
		t.Fatal(err)
	}
	if err := perform(&dialog{kind: "Move", sources: []string{filepath.Join(a.current().path, "renamed.txt")}, value: right}); err != nil {
		t.Fatal(err)
	}
}
func TestRender(t *testing.T) {
	a := fixture(t)
	second := &tab{path: a.panes[0].tabs[0].path, marked: map[string]bool{}}
	a.panes[0].tabs = append(a.panes[0].tabs, second)
	a.panes[0].current = 1
	a.load(second, second.path)
	img := ui.Render(a.view, 1100, 720, 1)
	f, err := os.Create("/tmp/twintrail-preview.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err = png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestFileMenuClosesTab(t *testing.T) {
	a := fixture(t)
	m := a.applicationMenu()
	closeTab := m.ItemByID("close-tab")
	if closeTab == nil || closeTab.Accelerator != "Cmd+W" {
		t.Fatalf("close tab item=%+v", closeTab)
	}
	var closeWindow *mygo.MenuItem
	for _, top := range m.Items() {
		if top.Label != "File" {
			continue
		}
		for _, item := range top.Submenu {
			if item.Role == mygo.RoleClose {
				closeWindow = item
			}
		}
	}
	if closeWindow == nil || closeWindow.Accelerator != "Shift+Cmd+W" {
		t.Fatalf("close window item=%+v", closeWindow)
	}
	tt := ui.NewTester(a.view, 1100, 720)
	tt.Frame()
	if closeTab.IsEnabled() {
		t.Fatal("Close Tab enabled with one tab")
	}
	a.newTab()
	tt.Frame()
	if !closeTab.IsEnabled() {
		t.Fatal("Close Tab disabled with two tabs")
	}
	closeTab.Click(closeTab, nil)
	tt.Frame()
	if len(a.panes[0].tabs) != 1 {
		t.Fatalf("menu did not close the tab: %d", len(a.panes[0].tabs))
	}
}

func TestTabCloseButton(t *testing.T) {
	a := fixture(t)
	tt := ui.NewTester(a.view, 1100, 720)
	tt.Key(ui.Cmd, ui.KeyT)
	tt.Frame()
	if len(a.panes[0].tabs) != 2 {
		t.Fatalf("tabs=%d", len(a.panes[0].tabs))
	}
	terminal := &fakeTerminal{}
	a.panes[0].tabs[0].terminal.session = terminal
	if err := tt.Click("Close tab 1 in pane 1"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(a.panes[0].tabs) != 1 {
		t.Fatalf("tab not closed: %d", len(a.panes[0].tabs))
	}
	if !terminal.closed {
		t.Fatal("closing a tab did not close its terminal")
	}
	if a.panes[0].current != 0 {
		t.Fatalf("current tab=%d", a.panes[0].current)
	}
	a.closeTab(&a.panes[0], 0)
	if len(a.panes[0].tabs) != 1 {
		t.Fatal("last tab was closed")
	}
}

func TestKeyboard(t *testing.T) {
	a := fixture(t)
	tt := ui.NewTester(a.view, 1100, 720)
	tt.Key(0, ui.KeyTab)
	if a.active != 1 {
		t.Fatalf("Tab active=%d", a.active)
	}
	tt.Key(0, ui.KeyTab)
	if a.active != 0 {
		t.Fatalf("Tab active=%d", a.active)
	}
	tt.Key(0, ui.KeyDown)
	if a.current().cursor != 1 {
		t.Fatalf("Down cursor=%d", a.current().cursor)
	}
	tt.Key(ui.Shift, ui.KeyDown)
	if !a.current().marked["Folder"] {
		t.Fatal("Shift Down did not mark Folder")
	}
	tt.Key(ui.Cmd, ui.KeyA)
	if !a.current().marked["hello.txt"] {
		t.Fatal("select all failed")
	}
	tt.Key(0, ui.KeyEscape)
	if len(a.current().marked) != 0 {
		t.Fatal("Escape did not clear selection")
	}
	tt.Key(ui.Cmd|ui.Shift, ui.KeyP)
	if a.modal == nil || a.modal.kind != "Actions" {
		t.Fatal("action panel missing")
	}
	tt.Type("rename")
	if a.modal.value != "rename" {
		t.Fatalf("action filter=%q", a.modal.value)
	}
	tt.Key(0, ui.KeyEscape)
	tt.Key(0, ui.KeyF7)
	if a.modal == nil || a.modal.kind != "New folder" {
		t.Fatal("new folder dialog missing")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Key(ui.Cmd, ui.KeyT)
	if len(a.panes[0].tabs) != 2 {
		t.Fatal("new tab missing")
	}
	tt.Key(ui.Cmd, ui.KeyW)
	if len(a.panes[0].tabs) != 1 {
		t.Fatal("tab close failed")
	}
}

// TestAppearanceSwitchesLook checks the View menu's appearance choice, which
// draws the window in the other look at once and keeps it in the session.
func TestAppearanceSwitchesLook(t *testing.T) {
	a := fixture(t)
	a.settingsWriter.save = func(session) {}
	a.light = false
	m := a.applicationMenu()
	light, dark := m.ItemByID("appearance-light"), m.ItemByID("appearance-dark")
	if light == nil || dark == nil {
		t.Fatal("the View menu has no appearance choice")
	}
	if light.Type != mygo.MenuItemRadio || dark.Type != mygo.MenuItemRadio {
		t.Fatalf("appearance items type %q, %q, want radio", light.Type, dark.Type)
	}
	if light.Checked || !dark.Checked {
		t.Fatalf("checked light=%v dark=%v, want the dark look", light.Checked, dark.Checked)
	}
	tt := ui.NewTester(a.view, 600, 400)
	tt.Frame()
	brightness := func(x, y int) float64 {
		r, g, b, _ := tt.Image().At(x, y).RGBA()
		return float64(r+g+b) / 3 / 256
	}
	pane, sidebar := brightness(500, 300), brightness(60, 300)
	light.Click(light, nil)
	tt.Frame()
	if !a.light {
		t.Fatal("the light look was not chosen")
	}
	if got := brightness(500, 300); got <= pane {
		t.Fatalf("the panes' background is %v, was %v before the light look", got, pane)
	}
	if got := brightness(60, 300); got <= sidebar {
		t.Fatalf("the sidebar's background is %v, was %v before the light look", got, sidebar)
	}
	if !a.snapshotSession().Light {
		t.Fatal("the session does not keep the light look")
	}
	m = a.applicationMenu()
	if !m.ItemByID("appearance-light").Checked || m.ItemByID("appearance-dark").Checked {
		t.Fatal("the menu does not check the look the window is drawn in")
	}
}
