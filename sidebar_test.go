package main

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

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
// Recents and Locations headers hides and shows their items over a moment, as
// does a click on the title, and that a collapsed section comes back collapsed.
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
	tt.Frame()
	// The items fold away over a moment rather than at once.
	settle := func() {
		tt.Advance(sidebarSectionTime * 2)
		tt.Frame()
	}
	if err := tt.Click("Collapse Favorites"); err != nil {
		t.Fatalf("%v (labels %v)", err, tt.Texts())
	}
	tt.Frame()
	tt.Advance(sidebarSectionTime / 2)
	if fold := a.favoritesFold.folded; fold <= 0 || fold >= 1 {
		t.Fatalf("favorites folded %v halfway, want a fold on its way", fold)
	}
	settle()
	if a.favoritesFold.open || !tt.HasText("Favorites") || tt.HasText("Marked") {
		t.Fatalf("favorites not collapsed: open=%v texts=%v", a.favoritesFold.open, tt.Texts())
	}
	if err := tt.Click("Expand Favorites"); err != nil {
		t.Fatalf("%v (labels %v)", err, tt.Texts())
	}
	tt.Frame()
	settle()
	if !a.favoritesFold.open || a.favoritesFold.folded != 0 || !tt.HasText("Marked") {
		t.Fatalf("favorites not shown: fold=%+v texts=%v", a.favoritesFold, tt.Texts())
	}
	// A click on the header's title does the same, as a framework section's.
	if err := tt.Click("Recents"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	settle()
	if a.recentsFold.open || tt.HasText("Recent Folder") {
		t.Fatalf("recents not collapsed: open=%v texts=%v", a.recentsFold.open, tt.Texts())
	}
	if err := tt.Click("Collapse Locations"); err != nil {
		t.Fatalf("%v (labels %v)", err, tt.Texts())
	}
	tt.Frame()
	settle()
	if a.locationsFold.open || !tt.HasText("Locations") || tt.HasText("Mounted Disk") {
		t.Fatalf("locations not collapsed: open=%v texts=%v", a.locationsFold.open, tt.Texts())
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
	if b.recentsFold.open || !b.favoritesFold.open || b.locationsFold.open {
		t.Fatalf("sections restored open=%v/%v/%v, want false/true/false",
			b.favoritesFold.open, b.recentsFold.open, b.locationsFold.open)
	}
	if b.recentsFold.folded != 1 || b.locationsFold.folded != 1 || b.favoritesFold.folded != 0 {
		t.Fatalf("sections restored folded=%v/%v/%v",
			b.favoritesFold.folded, b.recentsFold.folded, b.locationsFold.folded)
	}
	// One that came back folded folds open over a moment too.
	bt := ui.NewTester(b.view, 1100, 720)
	bt.Frame()
	if err := bt.Click("Expand Recents"); err != nil {
		t.Fatalf("%v (labels %v)", err, bt.Texts())
	}
	bt.Frame()
	bt.Advance(sidebarSectionTime / 2)
	if fold := b.recentsFold.folded; fold <= 0 || fold >= 1 {
		t.Fatalf("recents folded %v halfway open, want a fold on its way", fold)
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

// TestSidebarTextFollowsFileRows checks the sidebar's items read a step under
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
	item, size := height("Marked"), height("5 bytes")
	if item >= size || size-item > 3 {
		t.Fatalf("sidebar item %.2f high, the file rows' text %.2f, want a step under it", item, size)
	}
	// A larger File font takes the sidebar's items with it.
	a.fonts.Files.Size += 6
	tt.Frame()
	grown, grownSize := height("Marked"), height("5 bytes")
	if grown <= item {
		t.Fatalf("sidebar item %.2f high, was %.2f", grown, item)
	}
	if grown >= grownSize || grownSize-grown > 3 {
		t.Fatalf("sidebar item %.2f high, the file rows' text %.2f, want a step under it", grown, grownSize)
	}
}

// TestFavoritesUseTheirOwnIcon checks the favorites draw their folders with a
// star in them, which tells them from the recents' plain folders; the home
// folder keeps the roof it has everywhere.
func TestFavoritesUseTheirOwnIcon(t *testing.T) {
	a := fixture(t)
	dir := a.panes[0].tabs[0].path
	if favoriteIcon(savedLocation{Name: "Code", Path: dir}) != iconFavorite {
		t.Fatal("a favorite's folder does not take the starred icon")
	}
	if favoriteIcon(savedLocation{Name: "Home", Path: dir}) != iconHome {
		t.Fatal("the home favorite lost its roof")
	}
	if iconFavorite == iconFolder {
		t.Fatal("the favorites and the recents share one icon")
	}
	if locationIcon(savedLocation{Name: "Mounted Disk", Path: dir}) != iconFolder {
		t.Fatal("a location's folder does not take the plain icon")
	}
}

// TestViewModesSwitch checks the buttons at the right end of the breadcrumb row
// show a pane's folder as a list, as icons and as columns, that the mode a tab
// is in has the pressed look, and that the choice is kept.
func TestViewModesSwitch(t *testing.T) {
	a := fixture(t)
	a.settingsWriter.save = func(session) {}
	os.Mkdir(filepath.Join(a.panes[0].tabs[0].path, "Folder", "Inner"), 0755)
	tt := ui.NewTester(a.view, 1100, 720)
	tt.Frame()
	if !tt.HasText("Name") || !tt.HasText("Modified") {
		t.Fatalf("the list view has no header: %v", tt.Texts())
	}
	if a.panes[0].tabs[0].view != asList {
		t.Fatalf("a tab starts as %v", a.panes[0].tabs[0].view)
	}
	if err := tt.Click("As Icons"); err != nil {
		t.Fatalf("%v (labels %v)", err, tt.Texts())
	}
	tt.Frame()
	// The other pane keeps its header, whose labels Find returns first: the
	// first pane's is gone when its label sits in the first pane's half.
	if a.panes[0].tabs[0].view != asIcons {
		t.Fatalf("icons mode: view=%v", a.panes[0].tabs[0].view)
	}
	if r, ok := tt.Find("Modified"); ok && r.X < 550 {
		t.Fatalf("the first pane still shows its header at %v", r)
	}
	if !tt.HasText("hello.txt") || !tt.HasText("Folder") {
		t.Fatalf("icons mode lost the entries: %v", tt.Texts())
	}
	if err := tt.Click("As Columns"); err != nil {
		t.Fatalf("%v (labels %v)", err, tt.Texts())
	}
	tt.Frame()
	if a.panes[0].tabs[0].view != asColumns {
		t.Fatalf("columns mode: view=%v", a.panes[0].tabs[0].view)
	}
	if r, ok := tt.Find("Modified"); ok && r.X < 550 {
		t.Fatalf("the first pane still shows its header at %v", r)
	}
	if !tt.HasText("hello.txt") {
		t.Fatalf("columns mode lost the entries: %v", tt.Texts())
	}
	// Choosing the folder in the pane's column opens its column beside.
	if err := tt.Click("Folder"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("Folder") {
		t.Fatal("the column of the folder chosen is gone")
	}
	// The tabs keep their modes apart, and the choice is saved.
	if err := tt.Click("As List"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("Modified") {
		t.Fatal("the list view's header is gone")
	}
	if got := a.snapshotSession().Workspaces[0].Panes[0].Tabs[0].View; got != uint8(asList) {
		t.Fatalf("session keeps view %v", got)
	}
}

// TestNoParentRow checks the views of a folder show its entries alone, without
// the row of the parent folder the older views began with.
func TestNoParentRow(t *testing.T) {
	a := fixture(t)
	a.settingsWriter.save = func(session) {}
	tt := ui.NewTester(a.view, 1100, 720)
	tt.Frame()
	if tt.HasText("..") {
		t.Fatalf("the list shows a parent row: %v", tt.Texts())
	}
	if len(a.current().entries) != 2 {
		t.Fatalf("entries = %v", a.current().entries)
	}
	a.setViewMode(0, asIcons)
	tt.Frame()
	if tt.HasText("..") {
		t.Fatal("the icon view shows a parent row")
	}
	a.setViewMode(0, asColumns)
	tt.Frame()
	if tt.HasText("..") {
		t.Fatal("the column view shows a parent row")
	}
	// Going up still works, without a row of its own to click.
	before := a.current().path
	a.load(a.current(), filepath.Dir(before))
	tt.Frame()
	if got, want := a.current().path, filepath.Dir(before); got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}

// TestColumnsFollowClicks checks that choosing a folder in a column of the
// column view opens its column beside, and that choosing another folder in that
// column takes the columns after it with it.
func TestColumnsFollowClicks(t *testing.T) {
	a := fixture(t)
	a.settingsWriter.save = func(session) {}
	os.MkdirAll(filepath.Join(a.panes[0].tabs[0].path, "Folder", "Inner", "Deepest"), 0755)
	os.WriteFile(filepath.Join(a.panes[0].tabs[0].path, "Folder", "Inner", "note.txt"), []byte("hi"), 0644)
	a.load(a.current(), a.current().path)
	a.setViewMode(0, asColumns)
	tt := ui.NewTester(a.view, 1100, 720)
	tt.Frame()
	if err := tt.Click("Folder"); err != nil {
		t.Fatalf("no Folder in the first column: %v", tt.Texts())
	}
	tt.Frame()
	if len(a.panes[0].tabs[0].columns) < 2 || a.panes[0].tabs[0].columns[1].path == "" {
		t.Fatalf("choosing Folder opened no column: %v", a.panes[0].tabs[0].columns)
	}
	if !tt.HasText("Inner") {
		t.Fatalf("the second column does not show Inner: %v", tt.Texts())
	}
	if err := tt.Click("Inner"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("Deepest") {
		t.Fatalf("the third column does not show Deepest: %v", tt.Texts())
	}
	// Choosing a file opens nothing beside: Deepest stays a row of Inner's
	// column, and no fourth column takes it.
	if err := tt.Click("note.txt"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.panes[0].tabs[0].columns[3].path != "" {
		t.Fatalf("choosing a file opened a column: %q", a.panes[0].tabs[0].columns[3].path)
	}
	// And choosing the folder again opens it.
	if err := tt.Click("Deepest"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.panes[0].tabs[0].columns[3].path == "" {
		t.Fatal("choosing Deepest opened no column")
	}
}

// TestPinToFavorites checks the menus' Pin to Favorites puts a folder in the
// sidebar's Favorites, that a pinned folder's menu offers no second pin, and
// that the item acts on the entry it was opened on in every view.
func TestPinToFavorites(t *testing.T) {
	a := fixture(t)
	a.settingsWriter.save = func(session) {}
	os.MkdirAll(filepath.Join(a.panes[0].tabs[0].path, "Folder", "Inner"), 0755)
	a.favorites = nil
	tt := ui.NewTester(a.view, 1100, 720)
	tt.Frame()
	for _, mode := range []viewMode{asList, asIcons, asColumns} {
		a.setViewMode(0, mode)
		tt.Frame()
		if err := tt.RightClick("Folder"); err != nil {
			t.Fatalf("mode %v: %v", mode, err)
		}
		menu := tt.Menu()
		if len(menu) == 0 {
			tt.Frame()
			menu = tt.Menu()
		}
		found := false
		for _, item := range menu {
			if item == "Pin to Favorites" {
				found = true
			}
		}
		if !found {
			t.Fatalf("mode %v has no Pin to Favorites: %v", mode, menu)
		}
		if err := tt.ChooseMenuItem("Pin to Favorites"); err != nil {
			t.Fatalf("mode %v: %v", mode, err)
		}
		tt.Frame()
		if len(a.favorites) != 1 || a.favorites[0].Name != "Folder" {
			t.Fatalf("mode %v pinned nothing: %+v", mode, a.favorites)
		}
		// The pinned folder's menu offers no second pin.
		if err := tt.RightClick("Folder"); err != nil {
			t.Fatal(err)
		}
		if err := tt.ChooseMenuItem("Pin to Favorites"); err == nil {
			t.Fatalf("mode %v pins a pinned folder again", mode)
		}
		tt.CloseMenu()
		a.removeFavorite(filepath.Join(a.panes[0].tabs[0].path, "Folder"))
		tt.Frame()
	}
	// Pinning Inner from a column of the column view pins Inner, not the pane's
	// own folder.
	a.setViewMode(0, asColumns)
	tt.Frame()
	tt.Click("Folder")
	tt.Frame()
	if err := tt.RightClick("Inner"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Pin to Favorites"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	inner := filepath.Join(a.panes[0].tabs[0].path, "Folder", "Inner")
	if len(a.favorites) != 1 || a.favorites[0].Path != inner {
		t.Fatalf("pinned %+v, want Inner at %q", a.favorites, inner)
	}
}

// TestTabDragsBetweenPanes drags a tab of the left pane onto the right pane's
// strip, which moves the tab across and makes it current there; the pane it
// left keeps its other tab.
func TestTabDragsBetweenPanes(t *testing.T) {
	a := fixture(t)
	a.settingsWriter.save = func(session) {}
	a.newTerminal = func(string) (terminalSession, error) { return &fakeTerminal{}, nil }
	left := &a.panes[0]
	os.MkdirAll(filepath.Join(a.panes[0].tabs[0].path, "Second Folder"), 0755)
	tt := ui.NewTester(a.view, 1100, 720)
	tt.Frame()
	// A second tab of the left pane, at another folder.
	a.active, a.focus = 0, 1
	a.newTab()
	second := left.tabs[1]
	a.load(second, filepath.Join(a.panes[0].tabs[0].path, "Second Folder"))
	tt.Frame()
	if len(left.tabs) != 2 || len(a.panes[1].tabs) != 1 {
		t.Fatalf("panes: %d and %d tabs", len(left.tabs), len(a.panes[1].tabs))
	}
	// The strips of both panes sit on one row: the second tab's close button
	// marks the tab, and the first tab of the right pane marks where its strip
	// is. Press the tab beside its close button, which is no button itself.
	from, ok := tt.Find("Close tab 2 in pane 1")
	if !ok {
		t.Fatalf("no second tab: %v", tt.Texts())
	}
	to, ok := tt.Find("Close tab 1 in pane 2")
	if !ok {
		t.Fatalf("no first tab right: %v", tt.Texts())
	}
	sx, sy := from.X-60, from.Y+from.H/2
	tx, ty := to.X-60, sy
	tt.Press(sx, sy)
	tt.Frame()
	tt.Move((sx+tx)/2, ty)
	tt.Frame()
	tt.Move(tx, ty)
	tt.Frame()
	tt.Release(tx, ty)
	tt.Frame()
	if len(a.panes[1].tabs) != 2 || a.panes[1].tabs[1] != second {
		t.Fatalf("the tab did not move across: %d tabs right", len(a.panes[1].tabs))
	}
	if len(left.tabs) != 1 || left.tabs[0] == second {
		t.Fatalf("the pane it left: %d tabs", len(left.tabs))
	}
	if a.active != 1 || a.current() != second {
		t.Fatalf("active=%d, want the pane the tab joined at its moved tab", a.active)
	}
	if a.current().path != filepath.Join(a.panes[0].tabs[0].path, "Second Folder") {
		t.Fatalf("the moved tab holds %q", a.current().path)
	}
	tt.Frame()
	if !tt.HasText("Second Folder") {
		t.Fatalf("the moved tab's folder is not shown: %v", tt.Texts())
	}
}

// TestCompressPacksSelection checks the menus' Compress writes one zip beside
// the first of the entries it was opened on, whose entries walk the folders,
// and that the menu acts on the entry it was opened on in every view.
func TestCompressPacksSelection(t *testing.T) {
	a := fixture(t)
	a.settingsWriter.save = func(session) {}
	dir := a.panes[0].tabs[0].path
	os.MkdirAll(filepath.Join(dir, "Folder", "Inner"), 0755)
	os.WriteFile(filepath.Join(dir, "Folder", "Inner", "note.txt"), []byte("note"), 0644)
	os.WriteFile(filepath.Join(dir, "Folder", "top.txt"), []byte("top"), 0644)
	a.load(a.current(), dir)
	a.favorites = nil
	tt := ui.NewTester(a.view, 1100, 720)
	tt.Frame()
	for _, mode := range []viewMode{asList, asIcons, asColumns} {
		a.setViewMode(0, mode)
		tt.Frame()
		if err := tt.RightClick("Folder"); err != nil {
			t.Fatalf("mode %v: %v", mode, err)
		}
		menu := tt.Menu()
		if len(menu) == 0 {
			tt.Frame()
			menu = tt.Menu()
		}
		if !slices.Contains(menu, "Compress") || !slices.Contains(menu, "Open With…") {
			t.Fatalf("mode %v lacks Compress or Open With…: %v", mode, menu)
		}
		if err := tt.ChooseMenuItem("Compress"); err != nil {
			t.Fatalf("mode %v: %v", mode, err)
		}
		tt.Frame()
		for a.busy {
			tt.Advance(time.Second)
			tt.Frame()
		}
		zipPath := filepath.Join(dir, "Folder.zip")
		if _, err := os.Lstat(zipPath); err != nil {
			t.Fatalf("mode %v wrote no zip: %v", mode, err)
		}
		// The zip holds Folder's entries under its name.
		r, err := zip.OpenReader(zipPath)
		if err != nil {
			t.Fatal(err)
		}
		names := []string{}
		for _, f := range r.File {
			names = append(names, f.Name)
		}
		r.Close()
		if !slices.Contains(names, "Folder/top.txt") || !slices.Contains(names, "Folder/Inner/note.txt") {
			t.Fatalf("mode %v zip holds %v", mode, names)
		}
		if err := tt.RightClick("Folder"); err != nil {
			t.Fatalf("mode %v: %v", mode, err)
		}
		if err := tt.ChooseMenuItem("Open With…"); err != nil {
			t.Fatalf("mode %v: %v", mode, err)
		}
		tt.Frame()
		if a.modal == nil || a.modal.kind != "Open With" {
			t.Fatalf("mode %v opened no Open With dialog: %+v", mode, a.modal)
		}
		a.modal = nil
		tt.Frame()
		os.Remove(zipPath)
	}
	// Compressing a file zips the file alone.
	a.setViewMode(0, asList)
	tt.Frame()
	if err := tt.RightClick("hello.txt"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Compress"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	for a.busy {
		tt.Advance(time.Second)
		tt.Frame()
	}
	r, err := zip.OpenReader(filepath.Join(dir, "hello.txt.zip"))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if len(r.File) != 1 || r.File[0].Name != "hello.txt" {
		t.Fatalf("the zip of a file holds %v", r.File)
	}
	// An existing zip is not overwritten.
	before, _ := os.Stat(filepath.Join(dir, "hello.txt.zip"))
	if err := tt.RightClick("hello.txt"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Compress"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	for a.busy {
		tt.Advance(time.Second)
		tt.Frame()
	}
	if a.status == "Compress completed" {
		t.Fatal("an existing zip was overwritten")
	}
	after, _ := os.Stat(filepath.Join(dir, "hello.txt.zip"))
	if after.ModTime() != before.ModTime() {
		t.Fatal("an existing zip was rewritten")
	}
}

// TestEntryDragsBetweenPanes drags a file of the left pane onto the right
// pane, which moves it into the right pane's folder; the pane it left reloads
// without it.
func TestEntryDragsBetweenPanes(t *testing.T) {
	a := fixture(t)
	a.settingsWriter.save = func(session) {}
	dir := a.panes[0].tabs[0].path
	os.WriteFile(filepath.Join(dir, "moved.txt"), []byte("x"), 0644)
	a.load(a.current(), dir)
	tt := ui.NewTester(a.view, 1100, 720)
	tt.Frame()
	src, ok := tt.Find("moved.txt")
	if !ok {
		t.Fatalf("no moved.txt: %v", tt.Texts())
	}
	// Drop on the right pane: its tab marks where its strip is.
	right := a.panes[1].tabs[0].path
	dst, ok := tt.Find("Close tab 1 in pane 2")
	if !ok {
		t.Fatalf("no pane 2 tab: %v", tt.Texts())
	}
	dx, dy := dst.X-60, dst.Y+200
	tt.Press(src.X+src.W/2, src.Y+src.H/2)
	tt.Frame()
	tt.Move((src.X+dx)/2, dy)
	tt.Frame()
	tt.Move(dx, dy)
	tt.Frame()
	tt.Release(dx, dy)
	tt.Frame()
	tt.Frame()
	for a.busy {
		tt.Advance(time.Second)
		tt.Frame()
	}
	tt.Frame()
	tt.Frame()
	if _, err := os.Stat(filepath.Join(dir, "moved.txt")); !os.IsNotExist(err) {
		t.Fatalf("moved.txt stayed in the pane it left: %v", err)
	}
	if _, err := os.Stat(filepath.Join(right, "moved.txt")); err != nil {
		t.Fatalf("moved.txt is not in the pane it joined: %v", err)
	}
}
