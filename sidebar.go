package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

const (
	// defaultSidebarWidth is the width the sidebar starts with; the line
	// beside it resizes it from there.
	defaultSidebarWidth float32 = 210
	// minSidebarWidth keeps the sidebar's items readable when it is dragged
	// narrow.
	minSidebarWidth float32 = 120
)

var (
	iconWorkspace    = sidebarIcon(`<rect x="3" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5"/><rect x="3" y="14" width="7" height="7" rx="1.5"/><rect x="14" y="14" width="7" height="7" rx="1.5"/>`)
	iconNewWorkspace = sidebarIcon(`<path d="M12 5v14M5 12h14"/>`)
	iconHome         = sidebarIcon(`<path d="M3 10.5 12 3l9 7.5"/><path d="M5 9.5V21h14V9.5"/>`)
	iconFolder       = sidebarIcon(`<path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>`)
	iconCloud        = sidebarIcon(`<path d="M17.5 19a4.5 4.5 0 0 0 .5-9 6 6 0 0 0-11.6 1.6A3.5 3.5 0 0 0 6.5 19z"/>`)
	iconDrive        = sidebarIcon(`<rect x="3" y="5" width="18" height="14" rx="2"/><path d="M3 15h18"/><path d="M7 18h.01M11 18h.01"/>`)
	iconChevron      = sidebarIcon(`<path d="m9 6 6 6-6 6"/>`)
)

func sidebarIcon(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">` + shapes + `</svg>`))
}

// clampSidebarWidth keeps the sidebar readable: the split beside it stops the
// pointer later, at a width that shows little of an item.
func (a *app) clampSidebarWidth() {
	a.sidebarWidth = max(a.sidebarWidth, minSidebarWidth)
}

func (a *app) sidebarView(c *ui.Context) {
	theme := c.Theme()
	background := theme.Surface
	if !a.light {
		background = ui.Hex("#1b2229")
	}
	bar := c.TitleBar()
	ui.Column(c).Fill().Background(background).Children(func() {
		// The traffic lights sit over this strip, as Tilde's sidebar holds them.
		ui.Box(c).Height(max(bar.Height, 28)).DragWindow()
		ui.Text(c, "Twintrail").FontSize(15).Bold().Padding(10, 12).TextColor(theme.Text)
		sidebar := ui.Sidebar(c, &a.sidebarSelection, func() {
			// The Workspaces header carries a + of its own, so it is a row
			// rather than a SidebarSection, whose header has no room for
			// one.
			ui.Row(c).AlignItems(ui.Center).Padding(theme.Space(1), theme.Space(2)).Children(func() {
				ui.Text(c, "Workspaces").FontSize(theme.FontSize - 1).FontWeight(600).TextColor(theme.TextMuted).Grow(1)
				ui.ButtonBase(c).Label("New Workspace").Tooltip("New Workspace").Padding(theme.Space(0.5)).Radius(theme.Radius).Children(func() {
					ui.Icon(c, iconNewWorkspace).Size(12, 12).TextColor(theme.TextMuted)
				}).OnClick(a.createWorkspace)
			})
			for i := range a.workspaces {
				id := workspaceItemID(i)
				ui.SidebarItem(c, id, iconWorkspace, a.workspaces[i].Name).OnClick(func() { a.chooseSidebarItem(id) }).ContextMenu(func(m *ui.Menu) {
					if m.Item("Rename Workspace…").Chosen() {
						a.deferAction(c, func() { a.renameWorkspace(i) })
					}
					if m.Item("Delete Workspace").Disabled(len(a.workspaces) <= 1).Chosen() {
						a.deferAction(c, func() { a.deleteWorkspace(i) })
					}
				})
			}
			a.sidebarSection(c, "Favorites", &a.favoritesOpen, func() {
				for _, item := range a.favorites {
					item := item
					ui.SidebarItem(c, "favorite:"+item.Path, favoriteIcon(item), item.Name).OnClick(func() { a.navigateTo(item.Path) }).ContextMenu(func(m *ui.Menu) {
						if m.Item("Remove from Favorites").Chosen() {
							a.deferAction(c, func() { a.removeFavorite(item.Path) })
						}
					})
				}
			})
			if len(a.recents) > 0 {
				a.sidebarSection(c, "Recents", &a.recentsOpen, func() {
					for _, path := range a.recents {
						path := path
						ui.SidebarItem(c, "recent:"+path, iconFolder, displayName(path)).OnClick(func() { a.navigateTo(path) }).ContextMenu(func(m *ui.Menu) {
							if m.Item("Remove from Recents").Chosen() {
								a.deferAction(c, func() { a.removeRecent(path) })
							}
						})
					}
				})
			}
			if len(a.locations) > 0 {
				a.sidebarSection(c, "Locations", &a.locationsOpen, func() {
					for _, item := range a.locations {
						item := item
						ui.SidebarItem(c, "location:"+item.Path, locationIcon(item), item.Name).OnClick(func() { a.navigateTo(item.Path) })
					}
				})
			}
		})
		sidebar.Grow(1).Background(background)
	})
}

// sidebarSection builds a section of the sidebar whose header always shows a
// button at its right, which shows and hides the section's items and points
// down while they show. A framework SidebarSection shows its arrow only while
// the pointer rests on the title, so a section the user collapsed would look
// like an empty one. A click on the title shows and hides the items too.
func (a *app) sidebarSection(c *ui.Context, title string, open *bool, items func()) {
	theme := c.Theme()
	toggle := func() { *open = !*open }
	ui.Column(c).Gap(theme.Space(0.5)).Shrink(0).Margin(theme.Space(3), 0, 0, 0).Children(func() {
		head := ui.Row(c).AlignItems(ui.Center).Padding(theme.Space(1), theme.Space(2)).Radius(theme.Radius)
		head.OnClick(toggle).Children(func() {
			ui.Text(c, title).FontSize(theme.FontSize - 1).FontWeight(600).TextColor(theme.TextMuted).Grow(1)
			action := "Collapse " + title
			turn := float32(90)
			if !*open {
				action, turn = "Expand "+title, 0
			}
			button := ui.ButtonBase(c).Label(action).Tooltip(action).Padding(theme.Space(1)).Radius(theme.Radius)
			turn = button.Animate("open", turn, 150*time.Millisecond)
			button.Children(func() {
				ui.Icon(c, iconChevron).Size(12, 12).TextColor(theme.TextMuted).Rotate(turn)
			}).OnClick(toggle)
		})
		if *open {
			items()
		}
	})
}

func favoriteIcon(item savedLocation) *ui.SVG {
	if item.Name == "Home" {
		return iconHome
	}
	return iconFolder
}

func locationIcon(item savedLocation) *ui.SVG {
	home, _ := os.UserHomeDir()
	switch {
	case item.Name == "iCloud Drive":
		return iconCloud
	case item.Path == "/" || strings.HasPrefix(item.Path, "/Volumes/"):
		return iconDrive
	case item.Path == home:
		return iconHome
	default:
		return iconFolder
	}
}

func (a *app) chooseSidebarItem(id string) {
	if strings.HasPrefix(id, "workspace:") {
		index := 0
		if _, err := fmt.Sscanf(id, "workspace:%d", &index); err == nil {
			a.switchWorkspace(index)
		}
		return
	}
	for _, prefix := range []string{"favorite:", "recent:", "location:"} {
		if strings.HasPrefix(id, prefix) {
			a.navigateTo(strings.TrimPrefix(id, prefix))
			return
		}
	}
}

func (a *app) navigateTo(path string) {
	if path == "" {
		return
	}
	a.load(a.current(), path)
	a.focus = a.active + 1
}

// recordRecent keeps the last eight visited folders, newest first.
func (a *app) recordRecent(path string) {
	path = filepath.Clean(path)
	if path == "" || path == "." {
		return
	}
	if len(a.recents) > 0 && a.recents[0] == path {
		return
	}
	recents := []string{path}
	for _, p := range a.recents {
		if p != path && len(recents) < 8 {
			recents = append(recents, p)
		}
	}
	a.recents = recents
}

func (a *app) removeFavorite(path string) {
	favorites := a.favorites[:0]
	for _, item := range a.favorites {
		if item.Path != path {
			favorites = append(favorites, item)
		}
	}
	a.favorites = favorites
	if a.sidebarSelection == "favorite:"+path {
		a.sidebarSelection = ""
	}
	a.persistSettings(true)
}

func (a *app) removeRecent(path string) {
	recents := a.recents[:0]
	for _, recent := range a.recents {
		if recent != path {
			recents = append(recents, recent)
		}
	}
	a.recents = recents
	if a.sidebarSelection == "recent:"+path {
		a.sidebarSelection = ""
	}
	a.persistSettings(true)
}

func displayName(path string) string {
	if path == "/" {
		return "Macintosh HD"
	}
	home, _ := os.UserHomeDir()
	if home != "" && path == home {
		return filepath.Base(home)
	}
	return filepath.Base(path)
}

func defaultFavorites() []savedLocation {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return nil
	}
	out := []savedLocation{{Name: "Home", Path: home}}
	for _, name := range []string{"Desktop", "Documents", "Downloads"} {
		path := filepath.Join(home, name)
		if isDirectory(path) {
			out = append(out, savedLocation{Name: name, Path: path})
		}
	}
	return out
}

// systemLocations lists iCloud Drive, the home folder and mounted volumes.
func systemLocations() []savedLocation {
	home, _ := os.UserHomeDir()
	out := []savedLocation{}
	if home != "" {
		icloud := filepath.Join(home, "Library", "Mobile Documents", "com~apple~CloudDocs")
		if isDirectory(icloud) {
			out = append(out, savedLocation{Name: "iCloud Drive", Path: icloud})
		}
		out = append(out, savedLocation{Name: displayName(home), Path: home})
	}
	boot := false
	if entries, err := os.ReadDir("/Volumes"); err == nil {
		for _, entry := range entries {
			name := entry.Name()
			if strings.HasPrefix(name, ".") || name == "Recovery" {
				continue
			}
			path := filepath.Join("/Volumes", name)
			if !isDirectory(path) {
				continue
			}
			if resolved, err := filepath.EvalSymlinks(path); err == nil && resolved == "/" {
				boot = true
			}
			out = append(out, savedLocation{Name: name, Path: path})
		}
	}
	if !boot {
		out = append(out, savedLocation{Name: "Macintosh HD", Path: "/"})
	}
	return out
}

func existingLocations(items []savedLocation) []savedLocation {
	out := []savedLocation{}
	seen := map[string]bool{}
	for _, item := range items {
		if seen[item.Path] || !filepath.IsAbs(item.Path) || !isDirectory(item.Path) {
			continue
		}
		seen[item.Path] = true
		name := item.Name
		if name == "" {
			name = displayName(item.Path)
		}
		out = append(out, savedLocation{Name: name, Path: item.Path})
	}
	return out
}

func existingPaths(paths []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, path := range paths {
		if seen[path] || !filepath.IsAbs(path) || !isDirectory(path) {
			continue
		}
		seen[path] = true
		out = append(out, path)
	}
	return out
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
