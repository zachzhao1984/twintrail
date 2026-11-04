package main

import (
	"encoding/json"
	"errors"
	"github.com/egoist/mygo"
	"os"
	"path/filepath"
)

type savedTab struct {
	Path     string
	Hidden   bool
	Terminal bool
	// View is how the tab showed its folder, list when zero.
	View uint8 `json:"view,omitempty"`
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
	// TerminalPercent is the share of the room above the terminal that the
	// panes took, as the line above it was left.
	TerminalPercent float32 `json:"terminalPercent,omitempty"`
	// SidebarHidden tells the window was left with the sidebar folded away.
	SidebarHidden   bool             `json:"sidebarHidden,omitempty"`
	ActionBarHidden bool             `json:"actionBarHidden,omitempty"`
	Appearance      mygo.ThemeSource `json:"appearance,omitempty"`
	Light           bool
	Fonts           fontSettings      `json:"fonts"`
	Workspaces      []workspaceLayout `json:"workspaces,omitempty"`
	Workspace       int               `json:"workspace,omitempty"`
	Recents         []string          `json:"recents,omitempty"`
	Favorites       []savedLocation   `json:"favorites,omitempty"`
	FavoritesSet    bool              `json:"favoritesSet,omitempty"`
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
	a.appearance = s.Appearance
	if a.appearance != mygo.ThemeSystem && a.appearance != mygo.ThemeLight && a.appearance != mygo.ThemeDark {
		a.appearance = mygo.ThemeDark
		if s.Light {
			a.appearance = mygo.ThemeLight
		}
	}
	a.actionBarHidden = s.ActionBarHidden
	a.fonts = s.Fonts.normalized()
	if s.Width > 0 {
		a.lastWidth = s.Width
	}
	if s.SidebarWidth > 0 {
		a.sidebarWidth = s.SidebarWidth
		a.clampSidebarWidth()
	}
	a.sidebarHidden, a.sidebarShown = s.SidebarHidden, a.sidebarWidth
	if s.TerminalPercent > 0 {
		a.terminalPercent = s.TerminalPercent
	}
	if a.sidebarHidden {
		a.sidebarShown = 0
	}
	if s.FavoritesSet || len(s.Favorites) > 0 {
		a.favorites = existingLocations(s.Favorites)
	}
	a.favoritesFold.open, a.recentsFold.open = !s.FavoritesCollapsed, !s.RecentsCollapsed
	a.locationsFold.open = !s.LocationsCollapsed
	for _, fold := range []*sectionFold{&a.favoritesFold, &a.recentsFold, &a.locationsFold} {
		if !fold.open {
			fold.folded = 1
		}
	}
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
		SidebarHidden: a.sidebarHidden, ActionBarHidden: a.actionBarHidden, TerminalPercent: a.terminalPercent,
		Light: a.light, Appearance: a.appearance, Fonts: a.fonts,
		Workspaces: workspaces, Workspace: a.workspaceIndex, Recents: a.recents, Favorites: a.favorites,
		FavoritesSet: true, FavoritesCollapsed: !a.favoritesFold.open, RecentsCollapsed: !a.recentsFold.open,
		LocationsCollapsed: !a.locationsFold.open,
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
