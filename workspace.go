package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func workspaceItemID(index int) string {
	return fmt.Sprintf("workspace:%d", index)
}

// captureLayout turns the live panes into the layout a workspace remembers.
func (a *app) captureLayout() ([2]savedPane, int, float32) {
	var panes [2]savedPane
	for i, p := range a.panes {
		panes[i].Current = p.current
		for _, t := range p.tabs {
			panes[i].Tabs = append(panes[i].Tabs, savedTab{Path: t.path, Hidden: t.hidden, Terminal: t.terminal.visible, View: uint8(t.view)})
		}
	}
	return panes, a.active, a.split
}

// stashWorkspace records the live layout in the workspace being left.
func (a *app) stashWorkspace() {
	if a.workspaceIndex < 0 || a.workspaceIndex >= len(a.workspaces) {
		return
	}
	panes, active, split := a.captureLayout()
	w := &a.workspaces[a.workspaceIndex]
	w.Panes, w.Active, w.Split = panes, active, split
}

// createWorkspace adds a workspace holding the current layout and switches to it.
func (a *app) createWorkspace() {
	a.stashWorkspace()
	panes, active, split := a.captureLayout()
	name := fmt.Sprintf("Workspace %d", len(a.workspaces)+1)
	a.workspaces = append(a.workspaces, workspaceLayout{Name: name, Panes: panes, Active: active, Split: split})
	a.workspaceIndex = len(a.workspaces) - 1
	a.sidebarSelection = workspaceItemID(a.workspaceIndex)
	a.persistSettings(true)
}

func (a *app) switchWorkspace(index int) {
	if index < 0 || index >= len(a.workspaces) || index == a.workspaceIndex {
		return
	}
	a.stashWorkspace()
	a.closePaneTerminals()
	a.workspaceIndex = index
	a.applyWorkspace(a.workspaces[index], true)
	a.sidebarSelection = workspaceItemID(index)
	a.persistSettings(true)
}

func (a *app) renameWorkspace(index int) {
	if index < 0 || index >= len(a.workspaces) {
		return
	}
	a.modal = &dialog{kind: "Rename workspace", open: true, value: a.workspaces[index].Name, index: index}
}

func (a *app) confirmRenameWorkspace(d *dialog) {
	name := strings.TrimSpace(d.value)
	if name == "" {
		d.err = "Enter a workspace name"
		return
	}
	if d.index >= 0 && d.index < len(a.workspaces) {
		a.workspaces[d.index].Name = name
		a.persistSettings(true)
	}
	a.modal = nil
	a.focus = a.active + 1
}

// deleteWorkspace removes a saved layout. Deleting the current workspace
// shows the one before it, so a window always has a workspace.
func (a *app) deleteWorkspace(index int) {
	if index < 0 || index >= len(a.workspaces) || len(a.workspaces) <= 1 {
		return
	}
	if index == a.workspaceIndex {
		a.workspaces = append(a.workspaces[:index], a.workspaces[index+1:]...)
		next := index - 1
		if next < 0 {
			next = 0
		}
		if next >= len(a.workspaces) {
			next = len(a.workspaces) - 1
		}
		a.workspaceIndex = next
		a.closePaneTerminals()
		a.applyWorkspace(a.workspaces[next], true)
	} else {
		a.workspaces = append(a.workspaces[:index], a.workspaces[index+1:]...)
		if index < a.workspaceIndex {
			a.workspaceIndex--
		}
	}
	if strings.HasPrefix(a.sidebarSelection, "workspace:") {
		a.sidebarSelection = workspaceItemID(a.workspaceIndex)
	}
	a.persistSettings(true)
}

// applyWorkspace rebuilds both panes from a saved layout. Loads only when the
// app is running; the initial restore lets main start the tabs it restored.
func (a *app) applyWorkspace(w workspaceLayout, load bool) {
	home, _ := os.UserHomeDir()
	if home == "" {
		home = "/"
	}
	for i := range a.panes {
		tabs := make([]*tab, 0, len(w.Panes[i].Tabs))
		for _, st := range w.Panes[i].Tabs {
			if st.Path == "" || !filepath.IsAbs(st.Path) {
				continue
			}
			view := viewMode(0)
			if st.View <= uint8(asColumns) {
				view = viewMode(st.View)
			}
			t := &tab{path: st.Path, hidden: st.Hidden, marked: map[string]bool{}, view: view}
			t.terminal.visible = st.Terminal
			tabs = append(tabs, t)
		}
		if len(tabs) == 0 {
			tabs = append(tabs, &tab{path: home, marked: map[string]bool{}})
		}
		current := w.Panes[i].Current
		if current < 0 || current >= len(tabs) {
			current = 0
		}
		a.panes[i] = pane{tabs: tabs, current: current}
	}
	a.active = 0
	if w.Active == 1 {
		a.active = 1
	}
	a.focus = a.active + 1
	if w.Split >= 180 {
		a.split = w.Split
	}
	if !load {
		return
	}
	for i := range a.panes {
		for _, t := range a.panes[i].tabs {
			a.load(t, t.path)
			if t.terminal.visible {
				a.ensureTerminal(t)
			}
		}
	}
}

func (a *app) closePaneTerminals() {
	for i := range a.panes {
		for _, t := range a.panes[i].tabs {
			a.closeTerminal(t)
		}
	}
}

// sanitize drops paths that could not come from this machine and fills the
// values older sessions lack.
func (w *workspaceLayout) sanitize(index int) {
	if strings.TrimSpace(w.Name) == "" {
		if index == 0 {
			w.Name = "Default"
		} else {
			w.Name = fmt.Sprintf("Workspace %d", index+1)
		}
	}
	for i := range w.Panes {
		tabs := make([]savedTab, 0, len(w.Panes[i].Tabs))
		for _, t := range w.Panes[i].Tabs {
			if filepath.IsAbs(t.Path) {
				tabs = append(tabs, t)
			}
		}
		w.Panes[i].Tabs = tabs
		if w.Panes[i].Current < 0 || w.Panes[i].Current >= len(tabs) {
			w.Panes[i].Current = 0
		}
	}
	if w.Active < 0 || w.Active > 1 {
		w.Active = 0
	}
	if w.Split < 180 {
		w.Split = 550
	}
}
