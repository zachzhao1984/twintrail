package main

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSidebarShowsSections(t *testing.T) {
	a := fixture(t)
	a.favorites = []savedLocation{{Name: "Marked Folder", Path: a.panes[0].tabs[0].path}}
	a.locations = []savedLocation{{Name: "Macintosh HD", Path: "/"}}
	a.recents = []string{a.panes[1].tabs[0].path}
	tt := ui.NewTester(a.view, 1100, 720)
	for _, text := range []string{"Workspaces", "Favorites", "Recents", "Locations", "New Workspace", "Marked Folder", "Macintosh HD"} {
		if !tt.HasText(text) {
			t.Fatalf("sidebar missing %q: %q", text, tt.Texts())
		}
	}
	// The + sits at the right end of the Workspaces header, in its row.
	header, ok := tt.Find("Workspaces")
	if !ok {
		t.Fatal("no Workspaces header")
	}
	add, ok := tt.Find("New Workspace")
	if !ok {
		t.Fatal("no + button")
	}
	if add.X < header.X+header.W {
		t.Fatalf("+ not right of the header: header=%+v add=%+v", header, add)
	}
	if center := header.Y + header.H/2; add.Y > center || add.Y+add.H < center {
		t.Fatalf("+ not in the header row: header=%+v add=%+v", header, add)
	}
}

func TestSidebarNavigatesToFavorite(t *testing.T) {
	a := fixture(t)
	target := filepath.Join(a.panes[0].tabs[0].path, "Folder")
	a.favorites = []savedLocation{{Name: "Nested", Path: target}}
	a.locations = nil
	tt := ui.NewTester(a.view, 1100, 720)
	if err := tt.Click("Nested"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if got := a.current().path; got != target {
		t.Fatalf("current path=%q want %q", got, target)
	}
	if len(a.recents) == 0 || a.recents[0] != target {
		t.Fatalf("recent not recorded: %v", a.recents)
	}
}

func TestWorkspaceRemembersLayout(t *testing.T) {
	a := fixture(t)
	a.settingsWriter.save = func(session) {}
	tt := ui.NewTester(a.view, 1100, 720)
	tt.Key(ui.Cmd, ui.KeyT)
	tt.Frame()
	if len(a.panes[0].tabs) != 2 {
		t.Fatalf("new tab missing: %d", len(a.panes[0].tabs))
	}
	if err := tt.Click("New Workspace"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(a.workspaces) != 2 || a.workspaceIndex != 1 {
		t.Fatalf("workspace not created: %d at %d", len(a.workspaces), a.workspaceIndex)
	}
	tt.Key(ui.Cmd, ui.KeyW)
	tt.Frame()
	if len(a.panes[0].tabs) != 1 {
		t.Fatalf("tab not closed: %d", len(a.panes[0].tabs))
	}
	if err := tt.Click("Default"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.workspaceIndex != 0 || len(a.panes[0].tabs) != 2 {
		t.Fatalf("default layout not restored: workspace=%d tabs=%d", a.workspaceIndex, len(a.panes[0].tabs))
	}
	if err := tt.Click("Workspace 2"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.workspaceIndex != 1 || len(a.panes[0].tabs) != 1 {
		t.Fatalf("second layout not restored: workspace=%d tabs=%d", a.workspaceIndex, len(a.panes[0].tabs))
	}
}

func TestWorkspaceRenameAndDelete(t *testing.T) {
	a := fixture(t)
	a.settingsWriter.save = func(session) {}
	tt := ui.NewTester(a.view, 1100, 720)
	if err := tt.Click("New Workspace"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.RightClick("Workspace 2"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Rename Workspace…"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.modal == nil || a.modal.kind != "Rename workspace" {
		t.Fatalf("rename dialog=%+v", a.modal)
	}
	a.modal.value = "  "
	a.confirm()
	tt.Frame()
	if a.modal == nil || a.modal.err == "" {
		t.Fatal("empty name accepted")
	}
	a.modal.value = "Dev"
	a.confirm()
	tt.Frame()
	if a.workspaces[1].Name != "Dev" || !tt.HasText("Dev") {
		t.Fatalf("workspace not renamed: %+v", a.workspaces[1])
	}
	if err := tt.RightClick("Dev"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Delete Workspace"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(a.workspaces) != 1 || a.workspaceIndex != 0 || a.workspaces[0].Name != "Default" {
		t.Fatalf("workspace not deleted: %+v at %d", a.workspaces, a.workspaceIndex)
	}
	if err := tt.RightClick("Default"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Delete Workspace"); err == nil {
		t.Fatal("last workspace can be deleted")
	}
}

func TestWorkspaceDeleteBeforeCurrent(t *testing.T) {
	a := fixture(t)
	a.settingsWriter.save = func(session) {}
	tt := ui.NewTester(a.view, 1100, 720)
	for i := 0; i < 2; i++ {
		if err := tt.Click("New Workspace"); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
	}
	if len(a.workspaces) != 3 || a.workspaceIndex != 2 {
		t.Fatalf("workspaces=%+v at %d", a.workspaces, a.workspaceIndex)
	}
	// Deleting an earlier workspace shrinks the slice while the sidebar is
	// still building from it, which used to panic and close the window.
	if err := tt.RightClick("Default"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Delete Workspace"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(a.workspaces) != 2 || a.workspaceIndex != 1 {
		t.Fatalf("workspaces=%+v at %d", a.workspaces, a.workspaceIndex)
	}
	if a.workspaces[0].Name != "Workspace 2" || a.workspaces[1].Name != "Workspace 3" {
		t.Fatalf("wrong workspace removed: %+v", a.workspaces)
	}
	if tt.HasText("Default") || !tt.HasText("Workspace 3") {
		t.Fatalf("sidebar out of date: %q", tt.Texts())
	}
	if err := tt.Click("Workspace 2"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.workspaceIndex != 0 {
		t.Fatalf("workspace switch after delete failed: %d", a.workspaceIndex)
	}
}

func TestSidebarRemovesFavoritesAndRecents(t *testing.T) {
	a := fixture(t)
	a.settingsWriter.save = func(session) {}
	recent := filepath.Join(t.TempDir(), "Recent Target")
	if err := os.Mkdir(recent, 0755); err != nil {
		t.Fatal(err)
	}
	a.favorites = []savedLocation{{Name: "Favorite Target", Path: a.panes[0].tabs[0].path}}
	a.recents = []string{recent}
	tt := ui.NewTester(a.view, 1100, 720)
	if err := tt.RightClick("Favorite Target"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Remove from Favorites"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(a.favorites) != 0 || tt.HasText("Favorite Target") {
		t.Fatalf("favorite not removed: %+v", a.favorites)
	}
	if err := tt.RightClick("Recent Target"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Remove from Recents"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(a.recents) != 0 || tt.HasText("Recent Target") {
		t.Fatalf("recent not removed: %v", a.recents)
	}
}

func TestSidebarRemovesFirstOfSeveral(t *testing.T) {
	a := fixture(t)
	a.settingsWriter.save = func(session) {}
	first := filepath.Join(t.TempDir(), "First Recent")
	second := filepath.Join(t.TempDir(), "Second Recent")
	for _, path := range []string{first, second} {
		if err := os.Mkdir(path, 0755); err != nil {
			t.Fatal(err)
		}
	}
	a.favorites = []savedLocation{{Name: "First Favorite", Path: a.panes[0].tabs[0].path}, {Name: "Second Favorite", Path: a.panes[1].tabs[0].path}}
	a.recents = []string{first, second}
	tt := ui.NewTester(a.view, 1100, 720)
	if err := tt.RightClick("First Favorite"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Remove from Favorites"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(a.favorites) != 1 || a.favorites[0].Name != "Second Favorite" || !tt.HasText("Second Favorite") {
		t.Fatalf("wrong favorite removed: %+v", a.favorites)
	}
	if err := tt.RightClick("First Recent"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Remove from Recents"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(a.recents) != 1 || a.recents[0] != second || !tt.HasText("Second Recent") {
		t.Fatalf("wrong recent removed: %v", a.recents)
	}
}

func TestEmptyFavoritesStayRemoved(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	a := fixture(t)
	a.favorites = nil
	data, err := json.Marshal(a.snapshotSession())
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, "Library", "Application Support", "Twintrail", "session.json")
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	b := newApp(a.panes[0].tabs[0].path, a.panes[1].tabs[0].path)
	b.restore()
	if len(b.favorites) != 0 {
		t.Fatalf("empty favorites restored defaults: %+v", b.favorites)
	}
}

func TestWorkspaceSessionRoundTrip(t *testing.T) {
	a := fixture(t)
	a.settingsWriter.save = func(session) {}
	a.createWorkspace()
	a.panes[0].tabs[0].hidden = true
	a.panes[0].tabs[0].terminal.visible = true
	tab := &tab{path: a.panes[1].tabs[0].path, hidden: false, marked: map[string]bool{}}
	a.panes[0].tabs = append(a.panes[0].tabs, tab)
	a.panes[0].current = 1
	saved := a.snapshotSession()
	if len(saved.Workspaces) != 2 || saved.Workspace != 1 {
		t.Fatalf("workspaces=%d current=%d", len(saved.Workspaces), saved.Workspace)
	}
	if len(saved.Workspaces[1].Panes[0].Tabs) != 2 || saved.Workspaces[1].Panes[0].Current != 1 {
		t.Fatalf("current layout not captured: %+v", saved.Workspaces[1].Panes[0])
	}
	if !saved.Workspaces[1].Panes[0].Tabs[0].Terminal || !saved.Workspaces[1].Panes[0].Tabs[0].Hidden {
		t.Fatalf("tab state not captured: %+v", saved.Workspaces[1].Panes[0].Tabs[0])
	}
	data, err := json.Marshal(saved)
	if err != nil {
		t.Fatal(err)
	}
	var restored session
	if err = json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if len(restored.Workspaces) != 2 || restored.Workspaces[1].Name != "Workspace 2" {
		t.Fatalf("workspaces lost: %+v", restored.Workspaces)
	}
	if restored.Workspaces[1].Panes[0].Tabs[0].Path != saved.Workspaces[1].Panes[0].Tabs[0].Path {
		t.Fatal("tab path lost")
	}
}

func TestWorkspaceSanitize(t *testing.T) {
	w := workspaceLayout{Panes: [2]savedPane{{Tabs: []savedTab{{Path: "relative"}, {Path: "/tmp"}}, Current: 4}}, Active: 3}
	w.sanitize(0)
	if w.Name != "Default" || w.Active != 0 || w.Split != 550 {
		t.Fatalf("workspace not normalized: %+v", w)
	}
	if len(w.Panes[0].Tabs) != 1 || w.Panes[0].Tabs[0].Path != "/tmp" || w.Panes[0].Current != 0 {
		t.Fatalf("tabs not sanitized: %+v", w.Panes[0])
	}
}

func TestSessionKeepsLegacyPanes(t *testing.T) {
	a := fixture(t)
	saved := a.snapshotSession()
	if saved.Panes[0].Tabs[0].Path != a.panes[0].tabs[0].path {
		t.Fatal("legacy pane snapshot lost")
	}
}

// TestSidebarResizesByDrag drags the line beside the sidebar, which widens it,
// and keeps a width the items read in when it is dragged far left.
func TestSidebarResizesByDrag(t *testing.T) {
	a := fixture(t)
	a.settingsWriter.save = func(session) {}
	tt := ui.NewTester(a.view, 1100, 720)
	start := a.sidebarWidth
	line, y := start+1, float32(400)
	tt.Press(line, y)
	tt.Frame()
	tt.Move(line+90, y)
	tt.Frame()
	tt.Release(line+90, y)
	tt.Frame()
	if want := start + 90; a.sidebarWidth < want-2 || a.sidebarWidth > want+2 {
		t.Fatalf("sidebar width = %v, want about %v", a.sidebarWidth, want)
	}
	// Its items keep room to read: the line stops at the least width.
	line = a.sidebarWidth + 1
	tt.Press(line, y)
	tt.Frame()
	tt.Move(5, y)
	tt.Frame()
	tt.Release(5, y)
	tt.Frame()
	if a.sidebarWidth != minSidebarWidth {
		t.Fatalf("sidebar width = %v, want %v", a.sidebarWidth, minSidebarWidth)
	}
}

// TestSidebarWidthSaved checks the sidebar keeps its width across launches, and
// that a width too narrow to read comes back at the least one.
func TestSidebarWidthSaved(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	a := fixture(t)
	for _, want := range []float32{268, 10} {
		a.sidebarWidth = want
		data, err := json.Marshal(a.snapshotSession())
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(home, "Library", "Application Support", "Twintrail", "session.json")
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		b := newApp(a.panes[0].tabs[0].path, a.panes[1].tabs[0].path)
		b.restore()
		if got := max(want, minSidebarWidth); b.sidebarWidth != got {
			t.Fatalf("sidebar width restored as %v, want %v", b.sidebarWidth, got)
		}
	}
}

// TestSidebarSectionsCollapse checks the button at the right of the Favorites,
// Recents and Locations headers hides and shows their items, as does a click on
// the title, and that a collapsed section comes back collapsed.
func TestSidebarSectionsCollapse(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	a := fixture(t)
	a.settingsWriter.save = func(session) {}
	dir := a.panes[0].tabs[0].path
	if err := os.Mkdir(filepath.Join(dir, "Recent Folder"), 0755); err != nil {
		t.Fatal(err)
	}
	a.recents = []string{filepath.Join(dir, "Recent Folder")}
	a.favorites = []savedLocation{{Name: "Marked", Path: dir}}
	a.locations = []savedLocation{{Name: "Mounted Disk", Path: dir}}
	tt := ui.NewTester(a.view, 1100, 720)
	if err := tt.Click("Collapse Favorites"); err != nil {
		t.Fatalf("%v (labels %v)", err, tt.Texts())
	}
	tt.Frame()
	if a.favoritesOpen || !tt.HasText("Favorites") || tt.HasText("Marked") {
		t.Fatalf("favorites not collapsed: open=%v texts=%v", a.favoritesOpen, tt.Texts())
	}
	if err := tt.Click("Expand Favorites"); err != nil {
		t.Fatalf("%v (labels %v)", err, tt.Texts())
	}
	tt.Frame()
	if !a.favoritesOpen || !tt.HasText("Marked") {
		t.Fatalf("favorites not shown: open=%v texts=%v", a.favoritesOpen, tt.Texts())
	}
	// A click on the header's title does the same, as a framework section's.
	if err := tt.Click("Recents"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.recentsOpen || tt.HasText("Recent Folder") {
		t.Fatalf("recents not collapsed: open=%v texts=%v", a.recentsOpen, tt.Texts())
	}
	if err := tt.Click("Collapse Locations"); err != nil {
		t.Fatalf("%v (labels %v)", err, tt.Texts())
	}
	tt.Frame()
	if a.locationsOpen || !tt.HasText("Locations") || tt.HasText("Mounted Disk") {
		t.Fatalf("locations not collapsed: open=%v texts=%v", a.locationsOpen, tt.Texts())
	}
	// The collapsed sections are remembered.
	data, err := json.Marshal(a.snapshotSession())
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, "Library", "Application Support", "Twintrail", "session.json")
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	b := newApp(dir, dir)
	b.restore()
	if b.recentsOpen || !b.favoritesOpen || b.locationsOpen {
		t.Fatalf("sections restored open=%v/%v/%v, want false/true/false",
			b.favoritesOpen, b.recentsOpen, b.locationsOpen)
	}
}

// TestSidebarHidesAndShows folds the sidebar away from the button at its top
// bar, which gives its room to the panes, brings it back from the button in
// their strip, and keeps it hidden across launches.
func TestSidebarHidesAndShows(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	a := fixture(t)
	a.settingsWriter.save = func(session) {}
	tt := ui.NewTester(a.view, 1100, 720)
	if !tt.HasText("Workspaces") {
		t.Fatalf("sidebar missing: %v", tt.Texts())
	}
	start := a.sidebarShown
	if err := tt.Click("Hide Sidebar"); err != nil {
		t.Fatalf("%v (labels %v)", err, tt.Texts())
	}
	tt.Frame()
	// It folds away over a moment rather than at once.
	tt.Advance(sidebarFoldTime / 2)
	if a.sidebarShown <= 0 || a.sidebarShown >= start {
		t.Fatalf("sidebar width %v of %v halfway through the fold", a.sidebarShown, start)
	}
	tt.Advance(sidebarFoldTime)
	tt.Frame()
	if !a.sidebarHidden || a.sidebarShown != 0 || tt.HasText("Workspaces") || tt.HasText("Twintrail") {
		t.Fatalf("sidebar not hidden: hidden=%v width=%v texts=%v", a.sidebarHidden, a.sidebarShown, tt.Texts())
	}
	if !tt.HasText("Name") || !tt.HasText("hello.txt") {
		t.Fatalf("panes lost: %v", tt.Texts())
	}
	if err := tt.Click("Show Sidebar"); err != nil {
		t.Fatalf("%v (labels %v)", err, tt.Texts())
	}
	tt.Frame()
	tt.Advance(sidebarFoldTime * 2)
	tt.Frame()
	if a.sidebarHidden || !tt.HasText("Workspaces") {
		t.Fatalf("sidebar not shown: hidden=%v texts=%v", a.sidebarHidden, tt.Texts())
	}
	if a.sidebarWidth != start {
		t.Fatalf("sidebar width %v, want the one it had, %v", a.sidebarWidth, start)
	}
	// The window comes back as it was left.
	a.sidebarHidden = true
	data, err := json.Marshal(a.snapshotSession())
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, "Library", "Application Support", "Twintrail", "session.json")
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	b := newApp(a.panes[0].tabs[0].path, a.panes[1].tabs[0].path)
	b.restore()
	if !b.sidebarHidden {
		t.Fatal("the hidden sidebar was not restored")
	}
}

// TestSidebarTextMatchesFileRows checks the sidebar's items read at the size of
// the file rows beside them, and follow the File font setting with them.
func TestSidebarTextMatchesFileRows(t *testing.T) {
	a := fixture(t)
	a.settingsWriter.save = func(session) {}
	dir := a.panes[0].tabs[0].path
	a.favorites = []savedLocation{{Name: "Marked", Path: dir}}
	tt := ui.NewTester(a.view, 1100, 720)
	height := func(name string) float32 {
		r, ok := tt.Find(name)
		if !ok {
			t.Fatalf("no %q: %v", name, tt.Texts())
		}
		return r.H
	}
	item := height("Marked")
	if size := height("5 bytes"); math.Abs(float64(item-size)) > 0.6 {
		t.Fatalf("sidebar item %.2f high, the file rows' text %.2f", item, size)
	}
	// A larger File font takes the sidebar's items with it.
	a.fonts.Files.Size += 6
	tt.Frame()
	grown := height("Marked")
	if grown <= item {
		t.Fatalf("sidebar item %.2f high, was %.2f", grown, item)
	}
	if size := height("5 bytes"); math.Abs(float64(grown-size)) > 0.6 {
		t.Fatalf("sidebar item %.2f high, the file rows' text %.2f", grown, size)
	}
}
