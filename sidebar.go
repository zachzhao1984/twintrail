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
	// trafficLightsWidth is the room the window's buttons take at its left,
	// which what the user puts beside them keeps clear.
	trafficLightsWidth float32 = 84
)

// sidebarFoldTime is how long the sidebar takes to fold away or to come back,
// and sidebarSectionTime how long a section's items take to fold under their
// title.
const (
	sidebarFoldTime    = 220 * time.Millisecond
	sidebarSectionTime = 180 * time.Millisecond
)

var (
	iconWorkspace    = libraryIcon("layout-dashboard")
	iconNewWorkspace = sidebarIcon(`<path d="M12 5v14M5 12h14"/>`)
	iconHome         = libraryIcon("home")
	iconFolder       = libraryIcon("folder")
	// iconFavorite is a folder with a star in it, which tells the favorites from
	// the recents' plain folders at a glance.
	iconFavorite = libraryIcon("folder-star")
	iconCloud    = libraryIcon("cloud")
	iconDrive    = libraryIcon("hard-drive")
	// iconChevron stays a drawn chevron: it turns a quarter as a section folds,
	// which a library icon with a badge or a body of its own cannot do.
	iconChevron = sidebarIcon(`<path d="m9 6 6 6-6 6"/>`)
	iconSidebar = libraryIcon("sidebar")
	iconRows    = libraryIcon("rows")
	iconColumns = libraryIcon("columns")
)

// sidebarIcon draws an icon of the sidebar: the artwork is drawn in a 24 DIP
// box, which the view box trims a little, so its shapes fill their box without
// the weight of the file lists' icons beside them.
func sidebarIcon(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="1 1 22 22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">` + shapes + `</svg>`))
}

// sidebarItemFont is the font of the sidebar's items: the file lists' font, a
// step smaller, so that the sidebar reads as a band of its own beside the panes
// and follows the File font setting with them.
func (a *app) sidebarItemFont() fontSetting {
	f := a.fonts.Files
	f.Size--
	return f
}

// sidebarTitleFont is the font of a section's title, a step under the items'.
func (a *app) sidebarTitleFont() fontSetting {
	f := a.fonts.Files
	f.Size -= 2
	return f
}

// titleStripHeight is the height of the strip beside the traffic lights, which
// drags the window.
func titleStripHeight(c *ui.Context) float32 {
	return max(c.TitleBar().Height, 28)
}

// sidebarToggle draws the button that folds the sidebar away, or brings it
// back, as a sidebar with its items beside it.
func (a *app) sidebarToggle(c *ui.Context, tooltip string) {
	theme := c.Theme()
	button := ui.ButtonBase(c).Label(tooltip).Tooltip(tooltip).Padding(theme.Space(1.5)).Radius(theme.Radius * 2)
	if button.Hovered() {
		button.Background(theme.SurfaceHover)
	}
	button.Children(func() {
		ui.Icon(c, iconSidebar).Size(16, 16).TextColor(theme.TextMuted)
	}).OnClick(a.toggleSidebar)
}

// toggleSidebar hides the sidebar, giving its room to the panes over a moment,
// or shows it again at the width it had.
func (a *app) toggleSidebar() {
	a.sidebarHidden, a.sidebarFolding = !a.sidebarHidden, true
	a.persistSettings(true)
}

// foldSidebar moves the width the sidebar is laid out at to its own, or to
// nothing while it is hidden, easing out: e is the frame's root, whose state
// keeps the value. While no fold is on its way the divider owns that width, and
// the width the sidebar rests at follows it.
func (a *app) foldSidebar(e ui.Element) {
	if !a.sidebarFolding {
		if a.sidebarHidden {
			a.sidebarShown = 0
		} else {
			a.sidebarShown = max(a.sidebarShown, minSidebarWidth)
			a.sidebarWidth = a.sidebarShown
		}
		// Where a fold would start from.
		e.Jump("sidebar", a.sidebarShown)
		return
	}
	target := a.sidebarWidth
	if a.sidebarHidden {
		target = 0
	}
	a.sidebarShown = e.AnimateTo("sidebar", target, ui.Animation{Duration: sidebarFoldTime, Ease: ui.EaseInOut})
	if d := a.sidebarShown - target; d > -0.5 && d < 0.5 {
		a.sidebarShown, a.sidebarFolding = target, false
	}
}

// clampSidebarWidth keeps the sidebar readable: the split beside it stops the
// pointer later, at a width that shows little of an item.
func (a *app) clampSidebarWidth() {
	a.sidebarWidth = max(a.sidebarWidth, minSidebarWidth)
}

func (a *app) sidebarView(c *ui.Context) {
	theme := c.Theme()
	background := a.colors().sidebar
	column := ui.Column(c).Fill().Width(a.sidebarWidth).Background(background)
	column.Translate(a.sidebarShown-a.sidebarWidth, 0)
	column.Children(func() {
		// The traffic lights sit over this strip, as Tilde's sidebar holds them,
		// and the button at its right folds the sidebar away.
		ui.Row(c).Height(titleStripHeight(c)).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Height(titleStripHeight(c)).Grow(1).DragWindow()
			a.sidebarToggle(c, "Hide Sidebar")
		})
		sidebar := ui.Sidebar(c, &a.sidebarSelection, func() {
			ui.Column(c).FillWidth().Padding(6, 0, 12, 0).Children(func() {
				colors := a.jumpColors()
				ui.ButtonBase(c).Label("Jump to folder").Tooltip("Jump to folder (⌘K)").FillWidth().Height(32).PaddingX(10).Gap(8).AlignItems(ui.Center).Radius(8).Border(1, colors.border).Background(colors.strip).Children(func() {
					ui.Icon(c, jumpSearchIcon).Size(14, 14).TextColor(colors.muted).Shrink(0)
					ui.Text(c, "Jump to…").FontSize(14).TextColor(colors.muted).SingleLine().Grow(1).MinWidth(0)
					ui.Text(c, "⌘K").FontSize(11).FontWeight(600).Padding(3, 5).Radius(4).Background(colors.hover).TextColor(colors.muted).Shrink(0)
				}).OnClick(a.openJump)
			})
			// The Workspaces header carries a + of its own, so it is a row
			// rather than a SidebarSection, whose header has no room for
			// one.
			ui.Row(c).AlignItems(ui.Center).Padding(theme.Space(1), theme.Space(2)).Children(func() {
				a.sidebarTitleFont().apply(ui.Text(c, "Workspaces")).FontWeight(600).TextColor(theme.TextMuted).Grow(1)
				ui.ButtonBase(c).Label("New Workspace").Tooltip("New Workspace").Padding(theme.Space(0.5)).Radius(theme.Radius).Children(func() {
					ui.Icon(c, iconNewWorkspace).Size(12, 12).TextColor(theme.TextMuted)
				}).OnClick(a.createWorkspace)
			})
			for i := range a.workspaces {
				id := workspaceItemID(i)
				item := a.sidebarItemFont().apply(ui.SidebarItem(c, id, iconWorkspace, a.workspaces[i].Name))
				if h := item.Bounds().H; h > 0 {
					a.sidebarItemHeight = h
				}
				item.OnClick(func() { a.chooseSidebarItem(id) }).ContextMenu(func(m *ui.Menu) {
					if m.Item("Rename Workspace…").Chosen() {
						a.deferAction(c, func() { a.renameWorkspace(i) })
					}
					if m.Item("Delete Workspace").Disabled(len(a.workspaces) <= 1).Chosen() {
						a.deferAction(c, func() { a.deleteWorkspace(i) })
					}
				})
			}
			a.sidebarSection(c, "Favorites", &a.favoritesFold, len(a.favorites), func() {
				for _, item := range a.favorites {
					item := item
					a.sidebarItemFont().apply(ui.SidebarItem(c, "favorite:"+item.Path, favoriteIcon(item), item.Name)).OnClick(func() { a.navigateTo(item.Path) }).ContextMenu(func(m *ui.Menu) {
						if m.Item("Remove from Favorites").Chosen() {
							a.deferAction(c, func() { a.removeFavorite(item.Path) })
						}
					})
				}
			})
			if len(a.recents) > 0 {
				a.sidebarSection(c, "Recents", &a.recentsFold, len(a.recents), func() {
					for _, path := range a.recents {
						path := path
						a.sidebarItemFont().apply(ui.SidebarItem(c, "recent:"+path, iconFolder, displayName(path))).OnClick(func() { a.navigateTo(path) }).ContextMenu(func(m *ui.Menu) {
							if m.Item("Remove from Recents").Chosen() {
								a.deferAction(c, func() { a.removeRecent(path) })
							}
						})
					}
				})
			}
			if len(a.locations) > 0 {
				a.sidebarSection(c, "Locations", &a.locationsFold, len(a.locations), func() {
					for _, item := range a.locations {
						item := item
						a.sidebarItemFont().apply(ui.SidebarItem(c, "location:"+item.Path, locationIcon(item), item.Name)).OnClick(func() { a.navigateTo(item.Path) })
					}
				})
			}
		})
		sidebar.Grow(1).Background(background)
	})
}

// sectionFold is how far a section of the sidebar has folded its items away:
// whether they show, how far they have gone over a moment, and the height they
// take when they show, which the last frame that laid them out measured.
type sectionFold struct {
	open   bool
	folded float32
	height float32
}

// animate moves the fold to where the section's open state asks, over a
// moment: e is an element of the frame, whose state keeps the value.
func (f *sectionFold) animate(e ui.Element) {
	target := float32(0)
	if !f.open {
		target = 1
	}
	if f.folded == target {
		e.Jump("fold", target)
		return
	}
	f.folded = e.AnimateTo("fold", target, ui.Animation{Duration: sidebarSectionTime, Ease: ui.EaseInOut})
	if d := f.folded - target; d > -0.01 && d < 0.01 {
		f.folded = target
	}
}

// sidebarSection builds a section of the sidebar whose header always shows a
// button at its right, which shows and hides the section's items and points
// down while they show. A framework SidebarSection shows its arrow only while
// the pointer rests on the title, so a section the user collapsed would look
// like an empty one. A click on the title shows and hides the items too.
//
// The items fold away under the title over a moment, and unfold from it: the
// box that holds them takes the height they still take, and they fade as they
// go, while the sections under them slide up to where the title leaves them.
func (a *app) sidebarSection(c *ui.Context, title string, fold *sectionFold, count int, items func()) {
	theme := c.Theme()
	toggle := func() { fold.open = !fold.open }
	ui.Column(c).Gap(theme.Space(0.5)).Shrink(0).Margin(theme.Space(3), 0, 0, 0).Children(func() {
		head := ui.Row(c).AlignItems(ui.Center).Padding(theme.Space(1), theme.Space(2)).Radius(theme.Radius)
		fold.animate(head)
		head.OnClick(toggle).Children(func() {
			a.sidebarTitleFont().apply(ui.Text(c, title)).FontWeight(600).TextColor(theme.TextMuted).Grow(1)
			action := "Collapse " + title
			turn := float32(90)
			if !fold.open {
				action, turn = "Expand "+title, 0
			}
			button := ui.ButtonBase(c).Label(action).Tooltip(action).Padding(theme.Space(1)).Radius(theme.Radius)
			turn = button.Animate("open", turn, 150*time.Millisecond)
			button.Children(func() {
				ui.Icon(c, iconChevron).Size(12, 12).TextColor(theme.TextMuted).Rotate(turn)
			}).OnClick(toggle)
		})
		// Their box shows while the section is open, and while it folds away.
		// A section that starts folded has no height laid out yet: the items it
		// would hold take the height of one item, which the workspaces show.
		if fold.height <= 0 {
			fold.height = float32(count)*tallestItemHeight(a) + float32(max(count-1, 0))*theme.Space(0.5)
		}
		if fold.open || fold.folded < 0.99 {
			box := ui.Column(c).Clip().Shrink(0)
			// They take the height they still take, once theirs is known: while
			// they show whole, the layout gives them theirs to measure.
			if fold.height > 0 && fold.folded > 0 {
				box.Height(fold.height * (1 - fold.folded))
			}
			box.Children(func() {
				inner := ui.Column(c).Gap(theme.Space(0.5)).Shrink(0).Opacity(1 - fold.folded)
				inner.Children(items)
				// Their own height, which the box does not take from them.
				if fold.folded == 0 {
					if h := inner.Bounds().H; h > 0 {
						fold.height = h
					}
				}
			})
		}
	})
}

// tallestItemHeight is the height an item of the sidebar takes, as the last
// frame laid one out, or a guess before it laid any.
func tallestItemHeight(a *app) float32 {
	if a.sidebarItemHeight > 0 {
		return a.sidebarItemHeight
	}
	return float32(a.fonts.Files.Size)*1.4 + 2*3
}

func favoriteIcon(item savedLocation) *ui.SVG {
	if item.Name == "Home" {
		return iconHome
	}
	return iconFavorite
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

// addFavorite pins a folder to the end of Favorites once, as its menu's Pin to
// Favorites does; a folder already pinned stays where it is.
func (a *app) addFavorite(path string) {
	path = filepath.Clean(path)
	if path == "" || !filepath.IsAbs(path) || !isDirectory(path) || a.isFavorite(path) {
		return
	}
	a.favorites = append(a.favorites, savedLocation{Name: displayName(path), Path: path})
	a.favoritesFold.open = true
	a.persistSettings(true)
}

func (a *app) isFavorite(path string) bool {
	path = filepath.Clean(path)
	for _, item := range a.favorites {
		if item.Path == path {
			return true
		}
	}
	return false
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
