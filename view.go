package main

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"path/filepath"
	"strings"
)

func (a *app) view(c *ui.Context) {
	a.runPending()
	a.clampSidebarWidth()
	width, _ := c.Size()
	contentWidth := width - a.sidebarShown
	if a.sidebarShown <= 0 {
		contentWidth = width
	}
	if contentWidth > 0 && a.lastWidth > 0 && contentWidth != a.lastWidth {
		a.split *= contentWidth / a.lastWidth
	}
	if contentWidth > 0 {
		a.lastWidth = contentWidth
	}
	a.refreshTerminalMenu()
	a.refreshFileMenu()
	theme := martaTheme(a.light)
	c.SetTheme(theme)
	root := ui.Column(c).Fill().Gap(0).Background(theme.Background)
	a.foldSidebar(root)
	root.Children(func() {
		content := func() {
			ui.Column(c).Fill().MinWidth(0).Children(func() {
				// Keep the drag strip the traffic lights sit beside, which holds
				// the button that brings the sidebar back while it is hidden.
				ui.Row(c).Height(titleStripHeight(c)).AlignItems(ui.Center).Background(theme.Background).Children(func() {
					if a.sidebarShown <= 0 {
						// Clear the traffic lights, which sit at the window's left.
						ui.Box(c).Width(trafficLightsWidth).Height(titleStripHeight(c)).DragWindow()
						a.sidebarToggle(c, "Show Sidebar")
					}
					ui.Box(c).Height(titleStripHeight(c)).Grow(1).DragWindow()
				})
				ui.Split(c, &a.split, func() { a.paneView(c, 0) }, func() { a.paneView(c, 1) }).Grow(1)
				if a.status != "Ready" && a.status != "" {
					ui.Text(c, a.status).Height(20).PaddingX(6).TextColor(theme.TextMuted)
				}
				a.fonts.ActionBar.apply(ui.Row(c).Height(fontHeight(a.fonts.ActionBar, 20) + 6).Padding(3).Gap(3).Background(theme.Background)).Children(func() {
					for _, b := range []struct {
						name string
						fn   func()
					}{
						{"⌘↵ Open With", a.openWith}, {"F4 Edit", a.edit}, {"F5 Copy", func() { a.ask("Copy") }},
						{"F6 Move", func() { a.ask("Move") }}, {"⇧F6 Rename", func() { a.ask("Rename") }},
						{"F7 New Folder", func() { a.ask("New folder") }}, {"⇧F8 Delete", func() { a.ask("Delete") }},
					} {
						ui.ButtonBase(c).Grow(1).Height(fontHeight(a.fonts.ActionBar, 20)).Background(theme.Surface).Radius(2).PaddingX(5).AlignItems(ui.Center).Justify(ui.Start).Children(func() { ui.Text(c, b.name).TextColor(theme.TextMuted) }).OnClick(b.fn).Disabled(a.busy || a.modal != nil)
					}
				})
			})
		}
		panes := func() {
			// The line above the terminal resizes it, as the lines beside it
			// resize the panes and the sidebar.
			if a.current().terminal.visible {
				ui.SplitVerticalPercent(c, &a.terminalPercent, content, func() { a.terminalView(c) }).Fill()
				return
			}
			content()
		}
		// The line between the sidebar and the panes resizes the sidebar, as
		// the line between the panes resizes them.
		if a.sidebarShown <= 0 {
			panes()
		} else {
			ui.Split(c, &a.sidebarShown, func() { a.sidebarView(c) }, panes).Fill()
		}
		if a.modal != nil {
			a.dialogView(c)
		}
	})
	c.OnShortcut(ui.Cmd, ui.KeyComma, a.openPreferences)
	c.OnShortcut(ui.Cmd, ui.KeyBackquote, a.toggleTerminal)
	c.OnShortcut(ui.Cmd|ui.Shift, ui.KeyP, func() { a.ask("Actions") })
	if a.modal != nil || !a.current().terminal.handle.FocusWithin(c) {
		c.OnShortcut(0, ui.KeyEscape, func() {
			a.modal = nil
			a.current().filter = ""
			a.current().searchOpen = false
			a.current().marked = map[string]bool{}
			a.focus = a.active + 1
		})
	}
	if a.modal == nil && !a.current().terminal.handle.FocusWithin(c) {
		services := c.Services()
		c.OnShortcut(ui.Cmd, ui.KeyEnter, a.openWith)
		c.OnShortcut(0, ui.KeyF4, a.edit)
		c.OnShortcut(ui.Cmd, ui.KeyF, func() { a.current().searchOpen = true })
		c.OnShortcut(0, ui.KeySlash, func() { a.current().searchOpen = true })
		c.OnShortcut(0, ui.KeyF12, func() { services.WriteClipboard(a.current().path) })
		c.OnShortcut(ui.Cmd, ui.KeyF12, func() { services.WriteClipboard(strings.Join(a.targets(), "\n")) })
		c.OnShortcut(0, ui.KeyF11, func() { a.launch("open", "-a", "Terminal", a.current().path) })
		c.OnShortcut(0, ui.KeyTab, func() { a.active = 1 - a.active; a.focus = a.active + 1 })
		c.OnShortcut(ui.Cmd, ui.KeyT, a.newTab)
		c.OnShortcut(ui.Cmd, ui.KeyW, func() {
			p := &a.panes[a.active]
			a.closeTab(p, p.current)
		})
		c.OnShortcut(ui.Cmd, ui.KeyBackquote, a.toggleTerminal)
		c.OnShortcut(ui.Cmd|ui.Shift, ui.KeyPeriod, func() { t := a.current(); t.hidden = !t.hidden; a.load(t, t.path) })
		c.OnShortcut(ui.Cmd, ui.KeyR, func() { t := a.current(); a.load(t, t.path) })
		for _, s := range []struct {
			key  ui.Key
			mods ui.Modifiers
			kind string
		}{{ui.KeyF5, 0, "Copy"}, {ui.KeyF6, 0, "Move"}, {ui.KeyF6, ui.Shift, "Rename"}, {ui.KeyF7, 0, "New folder"}, {ui.KeyF8, 0, "Move to Trash"}, {ui.KeyF8, ui.Shift, "Delete"}, {ui.KeyL, ui.Cmd, "Go to path"}} {
			c.OnShortcut(s.mods, s.key, func() { a.ask(s.kind) })
		}
	}
}
func (a *app) paneView(c *ui.Context, index int) {
	p := &a.panes[index]
	t := p.tabs[p.current]
	theme, colors := c.Theme(), a.colors()
	active := a.active == index
	c.Key(fmt.Sprintf("pane-%d", index))
	ui.Column(c).Fill().Gap(0).Disabled(a.modal != nil).Children(func() {
		if len(p.tabs) > 1 {
			tabHeight := max(fontHeight(a.fonts.Tabs, 24), fontHeight(a.fonts.VirtualTabs, 24))
			ui.Row(c).Height(tabHeight).Gap(1).Background(theme.Border).Border(1, theme.Border).BorderWidth(0, 0, 1, 0).Children(func() {
				for i, tab := range p.tabs {
					bg, foreground := colors.strip, colors.muted
					if i == p.current && active {
						bg, foreground = colors.window, colors.text
					}
					tabFont := a.fonts.Tabs
					if i == p.current {
						tabFont = a.fonts.VirtualTabs
					}
					name := filepath.Base(tab.path)
					if tab.path == "/" {
						name = "Macintosh HD"
					}
					c.Key(fmt.Sprintf("tab-%p", tab))
					ui.Row(c).Grow(1).Height(tabHeight).Background(bg).Children(func() {
						c.Key(fmt.Sprintf("select-tab-%p", tab))
						tabFont.apply(ui.ButtonBase(c).Grow(1).Height(tabHeight).PaddingX(10).AlignItems(ui.Center)).Children(func() {
							ui.Text(c, name).SingleLine().TextColor(foreground)
						}).OnClick(func() { p.current = i; a.active = index; a.focus = index + 1 })
						c.Key(fmt.Sprintf("close-tab-%p", tab))
						ui.ButtonBase(c).Width(20).Height(tabHeight).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
							ui.Text(c, "×").FontSize(13).TextColor(foreground)
						}).Label(fmt.Sprintf("Close tab %d in pane %d", i+1, index+1)).OnClick(func() { a.active = index; a.focus = index + 1; a.closeTab(p, i) })
					}).ContextMenu(func(m *ui.Menu) {
						if m.Item("Close Tab").Disabled(len(p.tabs) <= 1).Chosen() {
							a.active = index
							a.focus = index + 1
							a.closeTab(p, i)
						}
					})
				}
			})
		}
		a.fonts.Breadcrumbs.apply(ui.Row(c).Height(fontHeight(a.fonts.Breadcrumbs, 24)).PaddingX(6).Gap(4).AlignItems(ui.Center)).Children(func() {
			ui.ButtonBase(c).Label(fmt.Sprintf("Root %d", index)).Children(func() { ui.Text(c, "/").TextColor(theme.TextMuted) }).OnClick(func() { a.active = index; a.load(t, "/") })
			parts := strings.Split(strings.Trim(t.path, "/"), "/")
			start := 0
			if len(parts) > 3 {
				start = len(parts) - 3
				ui.Button(c, "…").OnClick(func() { a.active = index; a.ask("Go to path") })
			}
			for i, part := range parts {
				if i < start {
					continue
				}
				if part == "" {
					continue
				}
				path := "/" + filepath.Join(parts[:i+1]...)
				ui.Text(c, "›").TextColor(theme.TextMuted)
				ui.ButtonBase(c).PaddingX(3).Children(func() { ui.Text(c, part).TextColor(theme.TextMuted) }).OnClick(func() { a.active = index; a.load(t, path) })
			}
			ui.Box(c).Grow(1)
			a.viewSwitch(c, index)
		})
		header := colors.strip
		sizeWidth := max(float32(80), float32(a.fonts.Files.Size)*6.7)
		modifiedWidth := max(float32(136), float32(a.fonts.Files.Size)*11.5)
		c.Key(fmt.Sprintf("header-%d-%d-%d", index, p.current, t.view))
		if t.view == asList {
			a.filesHeader(c, t, header, sizeWidth, modifiedWidth)
		}
		es := t.visible()
		c.Key(fmt.Sprintf("files-%d-%d-%d", index, p.current, t.view))
		switch t.view {
		case asList:
			a.filesList(c, index, t, es, active)
		case asIcons:
			a.filesIcons(c, index, t, es, active)
		case asColumns:
			a.filesColumns(c, index, t, es, active)
		}
		if t.searchOpen {
			c.Key(fmt.Sprintf("search-%d", index))
			ui.TextInput(c, &t.filter).Height(24).Placeholder("Quick search").Label(fmt.Sprintf("Quick search %d", index)).Focus().OnChange(func() { a.active = index; t.cursor = 0 })
		}
		footer := colors.strip
		a.fonts.StatusBar.apply(ui.Row(c).Height(fontHeight(a.fonts.StatusBar, 22)).PaddingX(4).Gap(4).AlignItems(ui.Center).Background(footer).Border(1, theme.Border).BorderWidth(1, 0, 0, 0)).Children(func() {
			ui.Text(c, tabSummary(t)).Grow(1).MinWidth(0).SingleLine().TextColor(theme.TextMuted)
			if t.hidden {
				ui.Text(c, "Hidden").TextColor(theme.TextMuted)
			}
		})
	})
}

// fileMenu fills a file row's context menu. The row that owns the menu is the
// one that was clicked, so the cursor moves to it and every item acts on that
// entry rather than on the row the cursor happened to be on. Marked entries
// still win, as targets returns them in place of the cursor's.
func (a *app) fileMenu(c *ui.Context, m *ui.Menu, index, row int) {
	a.active, a.focus = index, index+1
	t := a.panes[index].tabs[a.panes[index].current]
	moved := false
	if row >= 0 && row < len(t.visible()) && row != t.cursor {
		t.cursor = row
		moved = true
	}
	// The frame the menu opens in is the one the row moved in, but the
	// runtime shows the native menu once the frame is already presented,
	// and it stays up until an item is chosen, so nothing the row moved to
	// reaches the screen while the menu is there. Asking for a frame now
	// has it drawn again before the menu shows, as a menu of Finder's does
	// when it chooses the row it was opened on.
	if moved {
		c.Invalidate()
	}
	none := len(a.targets()) == 0 // the parent row, which has nothing to act on
	for _, kind := range []string{"Copy", "Move", "Rename", "New folder", "Move to Trash", "Delete"} {
		if m.Item(kind).Disabled(a.busy || none).Chosen() {
			a.ask(kind)
		}
	}
	m.Separator()
	if m.Item("Reveal in Finder").Disabled(none).Chosen() {
		if paths := a.targets(); len(paths) > 0 {
			a.launch("open", "-R", paths[0])
		}
	}
	if m.Item("Copy Path").Disabled(a.busy || none).Chosen() {
		if paths := a.targets(); len(paths) > 0 {
			c.Services().WriteClipboard(joinedPaths(paths))
		}
	}
	if m.Item("Open Terminal Here").Chosen() {
		a.launch("open", "-a", "Terminal", t.path)
	}
}
func (a *app) markMove(t *tab, es []entry, delta int) {
	if t.cursor >= 0 && t.cursor < len(es) && es[t.cursor].name != ".." {
		e := es[t.cursor]
		t.marked[e.name] = !t.marked[e.name]
	}
	t.cursor += delta
	if t.cursor < 0 {
		t.cursor = 0
	}
	if t.cursor >= len(es) {
		t.cursor = len(es) - 1
	}
	t.list.ScrollIntoView(t.cursor)
}
func (a *app) dialogView(c *ui.Context) {
	d := a.modal
	if !d.open {
		a.modal = nil
		a.focus = a.active + 1
		return
	}
	c.Key(fmt.Sprintf("dialog-%p", d))
	ui.DialogBase(c, &d.open, func(backdrop, panel ui.Element) {
		backdrop.Background(a.colors().scrim)
		panelWidth := float32(440)
		if d.kind == "Preference" {
			panelWidth = 660
		}
		if d.kind == "Open With" {
			panelWidth = 360
		}
		panel.Width(panelWidth).Background(c.Theme().Background).Border(1, c.Theme().Border).Radius(7).Padding(16)
		ui.Column(c).Gap(10).Children(func() {
			if d.kind == "Open With" {
				a.openWithView(c, d)
				return
			}
			if d.kind == "Preference" {
				a.preferencesView(c)
				return
			}
			title, action, label := d.kind, d.kind, d.kind+" destination"
			if d.kind == "Rename workspace" {
				title, action, label = "Rename Workspace", "Rename", "Workspace name"
			}
			ui.Text(c, title).Bold().FontSize(15)
			if d.kind == "Actions" {
				ui.TextInput(c, &d.value).Placeholder("Search actions…").Bind(&d.input).Focus()
				for _, kind := range []string{"Go to path", "Copy", "Move", "Rename", "New folder", "Move to Trash", "Delete"} {
					if strings.Contains(strings.ToLower(kind), strings.ToLower(d.value)) {
						ui.Button(c, kind).OnClick(func() { a.ask(kind) })
					}
				}
			} else {
				if len(d.sources) > 0 {
					ui.Text(c, fmt.Sprintf("%d item(s) · %s", len(d.sources), filepath.Base(d.sources[0]))).TextColor(c.Theme().TextMuted)
				}
				if d.kind != "Move to Trash" && d.kind != "Delete" {
					ui.TextInput(c, &d.value).Label(label).Bind(&d.input).Focus().OnSubmit(a.confirm)
				} else {
					message := "Move the selected items to the macOS Trash?"
					if d.kind == "Delete" {
						message = "Permanently delete these items? This cannot be undone."
					}
					ui.Text(c, message)
				}
				if d.err != "" {
					ui.Text(c, d.err).TextColor(c.Theme().Danger)
				}
				ui.Row(c).Gap(8).Children(func() {
					ui.PrimaryButton(c, action).Disabled(a.busy).OnClick(a.confirm)
					ui.Button(c, "Cancel").OnClick(func() { a.modal = nil; a.focus = a.active + 1 })
				})
			}
		})
	})
}

// viewSwitch draws the three buttons at the right end of the breadcrumb row,
// which show a pane's folder as a list, as icons or as columns, the way Finder
// does; the mode the pane is in has the accent's look.
func (a *app) viewSwitch(c *ui.Context, index int) {
	theme := c.Theme()
	modes := []struct {
		mode  viewMode
		icon  *ui.SVG
		label string
	}{{asList, iconRows, "As List"}, {asIcons, iconIcons, "As Icons"}, {asColumns, iconColumns, "As Columns"}}
	for _, m := range modes {
		m := m
		chosen := a.panes[index].tabs[a.panes[index].current].view == m.mode
		button := ui.ButtonBase(c).Label(m.label).Tooltip(m.label).Padding(theme.Space(1.5)).Radius(theme.Radius * 2)
		if chosen {
			button.Background(theme.SurfacePressed)
		} else if button.Hovered() {
			button.Background(theme.SurfaceHover)
		}
		button.Children(func() {
			ui.Icon(c, m.icon).Size(15, 15).TextColor(theme.TextMuted)
		}).OnClick(func() {
			a.active, a.focus = index, index+1
			a.setViewMode(index, m.mode)
		})
	}
}

// setViewMode shows a pane's folder the way mode says, keeping the entry the
// cursor was on.
func (a *app) setViewMode(index int, mode viewMode) {
	t := a.panes[index].tabs[a.panes[index].current]
	if t.view == mode {
		return
	}
	t.view = mode
	a.persistSettings(true)
}

// filesHeader draws the header of the list view, the Name, Size and Modified
// labels of the columns under it, which sort by name on a click.
func (a *app) filesHeader(c *ui.Context, t *tab, header ui.Color, sizeWidth, modifiedWidth float32) {
	theme := c.Theme()
	a.fonts.TableHeader.apply(ui.Row(c).Height(fontHeight(a.fonts.TableHeader, 20)).Background(header).Border(1, theme.Border).BorderWidth(1, 0, 1, 0).AlignItems(ui.Center)).Children(func() {
		ui.ButtonBase(c).Grow(1).PaddingX(4).Justify(ui.Start).Children(func() {
			ui.Row(c).Fill().AlignItems(ui.Center).Children(func() {
				ui.Text(c, "Name").Grow(1).TextColor(theme.TextMuted)
				arrow := "⌃"
				if t.descending {
					arrow = "⌄"
				}
				ui.Text(c, arrow).TextColor(theme.TextMuted)
			})
		}).OnClick(func() { t.descending = !t.descending; sortEntries(t.entries, t.descending) })
		ui.Text(c, "Size").Width(sizeWidth).PaddingX(4).Border(1, theme.Border).BorderWidth(0, 0, 0, 1).TextColor(theme.TextMuted)
		ui.Text(c, "Modified").Width(modifiedWidth).PaddingX(4).Border(1, theme.Border).BorderWidth(0, 0, 0, 1).TextColor(theme.TextMuted)
	})
}

// filesList draws a pane's folder as rows, the Name, Size and Modified
// columns of the original view.
func (a *app) filesList(c *ui.Context, index int, t *tab, es []entry, active bool) {
	theme := c.Theme()
	colors := a.colors()
	t.list.Selected = &t.cursor
	t.list.Label = func(i int) string { return es[i].name }
	sizeWidth := max(float32(80), float32(a.fonts.Files.Size)*6.7)
	modifiedWidth := max(float32(136), float32(a.fonts.Files.Size)*11.5)
	rowHeight := fontHeight(a.fonts.Files, fileRowHeight)
	ui.List(c, &t.list, len(es)).Grow(1).FocusRing(false).Draw(func(p *ui.Painter, r ui.Rect) {
		for y := float32(0); y < r.H; y += rowHeight * 2 {
			p.Fill(ui.Rect{X: r.X, Y: r.Y + y + rowHeight, W: r.W, H: min(rowHeight, max(0, r.H-y-rowHeight))}, colors.rowAlt, 0)
		}
	}).FocusBind(&a.focus, index+1).ItemKey(func(i int) any { return es[i].name }).Rows(func(r ui.ListRow) {
		e := es[r.Index]
		bg := colors.window
		if r.Index%2 == 1 {
			bg = colors.rowAlt
		}
		if r.Selected() && active {
			bg = theme.Accent
		}
		// A chosen row takes the accent's colors, its columns included.
		color, muted := colors.text, theme.TextMuted
		if r.Selected() && active {
			muted = colors.accentText.Alpha(0.8)
		}
		// A marked name keeps the amber, unless the row is chosen: the
		// accent says which row the keys act on, as it does in Finder.
		if t.marked[e.name] && !(r.Selected() && active) {
			color = colors.mark
		}
		a.fonts.Files.apply(ui.Row(c).Height(rowHeight).Background(bg).AlignItems(ui.Center)).Children(func() {
			ui.Row(c).Grow(1).PaddingX(4).Gap(4).AlignItems(ui.Center).Children(func() { a.fileIcon(c, e); ui.Text(c, e.name).TextColor(color).Grow(1).MinWidth(0).SingleLine() })
			size := martaSize(e.size)
			if e.dir {
				size = "--"
			}
			if e.name == ".." {
				size = ""
			}
			ui.Text(c, size).Width(sizeWidth).PaddingX(4).TextAlign(ui.End).TextColor(muted)
			date := modifiedText(e.modified)
			if e.name == ".." {
				date = ""
			}
			ui.Text(c, date).Width(modifiedWidth).PaddingX(4).TextColor(muted).SingleLine()
		}).ContextMenu(func(m *ui.Menu) { a.fileMenu(c, m, index, r.Index) })
	}).OnClick(func() { a.active = index; a.focus = index + 1 }).OnChange(func() { a.active = index }).OnSubmit(func() { a.active = index; a.open() }).OnShortcut(0, ui.KeyBackspace, func() { a.active = index; a.load(t, filepath.Dir(t.path)) }).OnShortcut(0, ui.KeySpace, func() {
		a.active = index
		targets := a.targets()
		if len(targets) > 0 {
			a.launch("qlmanage", append([]string{"-p"}, targets...)...)
		}
	}).OnShortcut(0, ui.KeyInsert, func() {
		if t.cursor >= 0 && t.cursor < len(es) && es[t.cursor].name != ".." {
			e := es[t.cursor]
			t.marked[e.name] = !t.marked[e.name]
		}
	}).OnShortcut(ui.Cmd, ui.KeyA, func() {
		for _, e := range t.entries {
			if e.name != ".." {
				t.marked[e.name] = true
			}
		}
	}).OnShortcut(ui.Shift, ui.KeyDown, func() { a.markMove(t, es, 1) }).OnShortcut(ui.Shift, ui.KeyUp, func() { a.markMove(t, es, -1) })
}

// filesIcons draws a pane's folder as icons in a grid, as Finder's icon view:
// the entry the cursor was on stays chosen, and the shortcuts of the list work
// here too.
func (a *app) filesIcons(c *ui.Context, index int, t *tab, es []entry, active bool) {
	colors := a.colors()
	t.grid.Selected = &t.cursor
	t.grid.Label = func(i int) string { return es[i].name }
	ui.GridView(c, &t.grid, len(es), 128, 116, func(i int) {
		e := es[i]
		color := colors.text
		if t.marked[e.name] {
			color = colors.mark
		}
		a.fonts.Files.apply(ui.Column(c).Fill().Padding(10, 6).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Grow(1).AlignItems(ui.Center).Children(func() { a.fileIconAt(c, e, 48, 48) })
			ui.Text(c, e.name).TextColor(color).MaxLines(2).TextAlign(ui.Center).SingleLine()
		}))
	}).Grow(1).FocusRing(false).FocusBind(&a.focus, index+1).ItemKey(func(i int) any { return es[i].name }).OnChange(func() { a.active = index }).OnSubmit(func() { a.active = index; a.open() }).OnShortcut(0, ui.KeyBackspace, func() { a.active = index; a.load(t, filepath.Dir(t.path)) }).OnShortcut(0, ui.KeySpace, func() {
		a.active = index
		targets := a.targets()
		if len(targets) > 0 {
			a.launch("qlmanage", append([]string{"-p"}, targets...)...)
		}
	}).OnShortcut(0, ui.KeyInsert, func() {
		if t.cursor >= 0 && t.cursor < len(es) && es[t.cursor].name != ".." {
			e := es[t.cursor]
			t.marked[e.name] = !t.marked[e.name]
		}
	}).OnShortcut(ui.Cmd, ui.KeyA, func() {
		for _, e := range t.entries {
			if e.name != ".." {
				t.marked[e.name] = true
			}
		}
	})
}

// filesColumns draws a pane's folder as columns, as Finder's column view: one
// column of the folder along the path the pane's cursor holds, and one of every
// folder chosen in the column before it. Choosing a folder opens its column
// beside; opening, the shortcuts of the list, act on the pane's entry, which the
// column of the pane's folder chooses.
func (a *app) filesColumns(c *ui.Context, index int, t *tab, es []entry, active bool) {
	a.fonts.Files.apply(ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
		path := t.path
		for i := range t.columns {
			fold := &t.columns[i]
			sub := a.dirEntries(path)
			if sub == nil {
				break
			}
			fold.entries, fold.path = sub, path
			a.filesColumn(c, index, fold, active)
			if fold.cursor < 0 || fold.cursor >= len(sub) || !sub[fold.cursor].dir {
				break
			}
			path = filepath.Join(path, sub[fold.cursor].name)
		}
	}))
}

// column is one column of the column view: the entries of the folder it shows,
// the path of that folder, and the entry chosen there. The list is the state of
// the column's rows, which each tab keeps for itself.
type column struct {
	entries []entry
	path    string
	cursor  int
	list    ui.ListState
}

// dirEntries reads a folder's entries, the parent row first, or nil where the
// path holds none.
func (a *app) dirEntries(path string) []entry {
	es, err := readDirectory(path, false)
	if err != nil {
		return nil
	}
	return es
}

// filesColumn draws one column of the column view: the rows of the list, which
// choose in the column's cursor.
func (a *app) filesColumn(c *ui.Context, index int, col *column, active bool) {
	colors := a.colors()
	theme := c.Theme()
	col.list.Selected = &col.cursor
	col.list.Label = func(i int) string { return col.entries[i].name }
	ui.List(c, &col.list, len(col.entries)).Width(220).Shrink(0).FocusRing(false).ItemKey(func(i int) any { return col.entries[i].name }).Rows(func(r ui.ListRow) {
		e := col.entries[r.Index]
		bg := colors.window
		if r.Selected() && active {
			bg = theme.Accent
		}
		color := colors.text
		a.fonts.Files.apply(ui.Row(c).Height(fontHeight(a.fonts.Files, fileRowHeight)).Background(bg).AlignItems(ui.Center)).Children(func() {
			ui.Row(c).Grow(1).PaddingX(4).Gap(4).AlignItems(ui.Center).Children(func() { a.fileIcon(c, e); ui.Text(c, e.name).TextColor(color).Grow(1).MinWidth(0).SingleLine() })
		}).OnClick(func() { a.active, a.focus = index, index+1 })
	}).OnChange(func() { a.active = index })
}
