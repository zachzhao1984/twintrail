package main

import (
	"fmt"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestFuzzyFolder(t *testing.T) {
	for _, tc := range []struct {
		path, query string
		want        bool
	}{
		{"/Users/Zach/Workspace/Twinrail", "twnrl", true},
		{"/Users/Zach/Workspace/Twinrail", "WORK/twin", true},
		{"/Users/Zach/Workspace/Twinrail", "twinzz", false},
	} {
		if got := fuzzyFolder(tc.path, tc.query); got != tc.want {
			t.Errorf("fuzzyFolder(%q, %q) = %v", tc.path, tc.query, got)
		}
	}
}

func TestJumpOpensCurrentTab(t *testing.T) {
	a := fixture(t)
	target := filepath.Join(a.current().path, "Folder")
	a.recents = []string{target}
	tt := ui.NewTester(a.view, 1100, 720)
	tt.Key(ui.Cmd, ui.KeyK)
	tt.Frame()
	if a.modal == nil || a.modal.kind != "Jump" {
		t.Fatal("shortcut did not open jump")
	}
	if err := tt.Click(target); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.modal != nil || a.current().path != target {
		t.Fatalf("jump failed: modal=%v path=%s", a.modal, a.current().path)
	}
	if err := tt.Click("Jump to folder"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.modal == nil || a.modal.kind != "Jump" {
		t.Fatal("sidebar button did not open jump")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if a.modal != nil {
		t.Fatal("escape did not close jump")
	}
}

func TestJumpLayoutDoesNotMoveWithResults(t *testing.T) {
	a := fixture(t)
	a.openJump()
	tt := ui.NewTester(a.view, 1100, 720)
	before, ok := tt.Find("Search folders")
	if !ok {
		t.Fatal("search input missing")
	}
	a.modal.jumpResults = nil
	a.modal.loading = true
	tt.Frame()
	empty, _ := tt.Find("Search folders")
	a.modal.jumpResults = []string{"/a", "/b", "/c", "/d", "/e", "/f", "/g", "/h"}
	tt.Frame()
	full, _ := tt.Find("Search folders")
	if before != empty || empty != full {
		t.Fatalf("input moved: initial=%+v loading=%+v results=%+v", before, empty, full)
	}
}

func TestJumpKeepsResultsWhileTypingAndScrolls(t *testing.T) {
	a := fixture(t)
	a.openJump()
	d := a.modal
	old := append([]string(nil), d.jumpResults...)
	a.update = func(func()) { t.Error("canceled search published results") }
	d.value = "new query"
	a.searchJump(d)
	d.jumpCancel()
	if len(d.jumpResults) != len(old) || d.jumpResults[0] != old[0] {
		t.Fatal("typing cleared visible results")
	}
	a.update = nil
	d.value = ""
	a.searchJump(d)
	for i := 0; i < 30; i++ {
		d.jumpResults = append(d.jumpResults, filepath.Join("/folders", fmt.Sprint(i)))
	}
	tt := ui.NewTester(a.view, 1100, 720)
	before, _ := tt.Find("Search folders")
	for i := 0; i < 20; i++ {
		tt.Key(0, ui.KeyDown)
		tt.Frame()
	}
	if d.index != 20 {
		t.Fatalf("selected row = %d", d.index)
	}
	after, _ := tt.Find("Search folders")
	if before != after {
		t.Fatal("scroll moved search field")
	}
	if _, ok := tt.Find(d.jumpResults[20]); !ok {
		t.Fatal("selected row not scrolled into view")
	}
}

func TestJumpSearchUsesIndexAndRejectsOldQueries(t *testing.T) {
	a := fixture(t)
	idx := newFolderIndex("")
	target := filepath.Join(a.current().path, "Folder")
	idx.add([]string{target, "/old/bucketeer"})
	idx.finished = time.Now()
	a.folderIndex = idx
	updates := make(chan func(), 16)
	a.update = func(fn func()) { updates <- fn }
	a.openJump()
	d := a.modal
	d.value = "bucketeer"
	a.searchJump(d)
	var obsolete func()
	select {
	case obsolete = <-updates:
	case <-time.After(time.Second):
		t.Fatal("first query timed out")
	}
	d.value = "Folder"
	a.searchJump(d)
	obsolete()
	if slices.Contains(d.jumpResults, "/old/bucketeer") {
		t.Fatal("obsolete result published")
	}
	select {
	case fn := <-updates:
		fn()
	case <-time.After(time.Second):
		t.Fatal("second query timed out")
	}
	if len(d.jumpResults) == 0 || d.jumpResults[0] != target {
		t.Fatal(d.jumpResults)
	}
	a.update = nil
	tt := ui.NewTester(a.view, 1100, 720)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if a.modal != nil || a.current().path != target {
		t.Fatalf("indexed result did not open in current tab: %s", a.current().path)
	}
}
