package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type savedTab struct {
	Path     string
	Hidden   bool
	Terminal bool
}
type savedPane struct {
	Tabs    []savedTab
	Current int
}
type workspaceLayout struct {
	Name   string
	Panes  [2]savedPane
	Active int
	Split  float32
}
type savedLocation struct {
	Name string
	Path string
}
type session struct {
	Panes        [2]savedPane
	Active       int
	Split        float32
	Width        float32
	SidebarWidth float32 `json:"sidebarWidth,omitempty"`
	// SidebarHidden tells the window was left with the sidebar folded away.
	SidebarHidden bool `json:"sidebarHidden,omitempty"`
	Light         bool
	Fonts         fontSettings      `json:"fonts"`
	Workspaces    []workspaceLayout `json:"workspaces,omitempty"`
	Workspace     int               `json:"workspace,omitempty"`
	Recents       []string          `json:"recents,omitempty"`
	Favorites     []savedLocation   `json:"favorites,omitempty"`
	FavoritesSet  bool              `json:"favoritesSet,omitempty"`
	// The sidebar sections the user collapsed, open by default.
	FavoritesCollapsed bool `json:"favoritesCollapsed,omitempty"`
	RecentsCollapsed   bool `json:"recentsCollapsed,omitempty"`
	LocationsCollapsed bool `json:"locationsCollapsed,omitempty"`
}

func sessionPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "Twintrail", "session.json")
}

// The old directory is read only when the new session does not exist.
// Saving always uses Twintrail, so existing preferences migrate on exit.
const legacyApplicationName = "mygo-commander"

func readSessionData(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if !errors.Is(err, os.ErrNotExist) || path == "" {
		return data, err
	}
	legacy := filepath.Join(filepath.Dir(filepath.Dir(path)), legacyApplicationName, "session.json")
	return os.ReadFile(legacy)
}

func (a *app) restore() {
	data, err := readSessionData(sessionPath())
	if err != nil {
		return
	}
	var s session
	if json.Unmarshal(data, &s) != nil {
		return
	}
	a.light = s.Light
	a.fonts = s.Fonts.normalized()
	if s.Width > 0 {
		a.lastWidth = s.Width
	}
	if s.SidebarWidth > 0 {
		a.sidebarWidth = s.SidebarWidth
		a.clampSidebarWidth()
	}
	a.sidebarHidden, a.sidebarShown = s.SidebarHidden, a.sidebarWidth
	if a.sidebarHidden {
		a.sidebarShown = 0
	}
	if s.FavoritesSet || len(s.Favorites) > 0 {
		a.favorites = existingLocations(s.Favorites)
	}
	a.favoritesOpen, a.recentsOpen = !s.FavoritesCollapsed, !s.RecentsCollapsed
	a.locationsOpen = !s.LocationsCollapsed
	a.recents = existingPaths(s.Recents)
	if len(s.Workspaces) > 0 {
		a.workspaces = append([]workspaceLayout(nil), s.Workspaces...)
		for i := range a.workspaces {
			a.workspaces[i].sanitize(i)
		}
	} else {
		a.workspaces = []workspaceLayout{{Name: "Default", Panes: s.Panes, Active: s.Active, Split: s.Split}}
		a.workspaces[0].sanitize(0)
	}
	if s.Workspace >= 0 && s.Workspace < len(a.workspaces) {
		a.workspaceIndex = s.Workspace
	}
	a.applyWorkspace(a.workspaces[a.workspaceIndex], false)
	a.sidebarSelection = workspaceItemID(a.workspaceIndex)
}
func (a *app) snapshotSession() session {
	panes, active, split := a.captureLayout()
	workspaces := append([]workspaceLayout(nil), a.workspaces...)
	if a.workspaceIndex >= 0 && a.workspaceIndex < len(workspaces) {
		workspaces[a.workspaceIndex] = workspaceLayout{Name: workspaces[a.workspaceIndex].Name, Panes: panes, Active: active, Split: split}
	}
	return session{
		Panes: panes, Active: active, Split: split, Width: a.lastWidth, SidebarWidth: a.sidebarWidth,
		SidebarHidden: a.sidebarHidden, Light: a.light, Fonts: a.fonts,
		Workspaces: workspaces, Workspace: a.workspaceIndex, Recents: a.recents, Favorites: a.favorites,
		FavoritesSet: true, FavoritesCollapsed: !a.favoritesOpen, RecentsCollapsed: !a.recentsOpen,
		LocationsCollapsed: !a.locationsOpen,
	}
}
func saveSession(s session) {
	path := sessionPath()
	if path == "" {
		return
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return
	}
	if os.MkdirAll(filepath.Dir(path), 0700) != nil {
		return
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "session-*")
	if err != nil {
		return
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
		return
	}
	if tmp.Close() != nil {
		return
	}
	os.Rename(name, path)
}
