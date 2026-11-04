package main

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// fuzzyFolder matches characters in order, ignoring case, against the whole path.
func fuzzyFolder(path, query string) bool {
	return subsequence(strings.ToLower(path), strings.ToLower(strings.TrimSpace(query)))
}

func (a *app) closeModal() {
	if a.modal != nil && a.modal.jumpCancel != nil {
		a.modal.jumpCancel()
	}
	a.modal = nil
}

func (a *app) openJump() {
	if a.modal != nil {
		return
	}
	d := &dialog{kind: "Jump", open: true, jumpTab: a.current(), jumpPane: a.active}
	seen := map[string]bool{}
	add := func(path string) {
		if path != "" && !seen[path] {
			seen[path] = true
			d.jumpResults = append(d.jumpResults, path)
		}
	}
	add(a.current().path)
	for _, p := range a.recents {
		add(p)
	}
	for _, p := range a.favorites {
		add(p.Path)
	}
	a.modal = d
	if a.update != nil {
		if a.folderIndex == nil {
			a.folderIndex = newFolderIndex(folderCachePath())
		}
		roots := []string{filepath.Dir(d.jumpTab.path), d.jumpTab.path}
		roots = append(roots, a.recents...)
		a.folderIndex.refresh(folderSearchRoots(roots), false)
	}
}

func (a *app) searchJump(d *dialog) {
	if d.jumpCancel != nil {
		d.jumpCancel()
		d.jumpCancel = nil
	}
	d.err = ""
	query := strings.TrimSpace(d.value)
	if query == "" {
		d.jumpResults = nil
		d.index = 0
		d.jumpQuery = ""
		d.jumpList.ScrollTo(0, ui.Start)
		// Restore suggestions without replacing the dialog or its focused input.
		seen := map[string]bool{}
		add := func(p string) {
			if p != "" && !seen[p] {
				seen[p] = true
				d.jumpResults = append(d.jumpResults, p)
			}
		}
		add(d.jumpTab.path)
		for _, p := range a.recents {
			add(p)
		}
		for _, p := range a.favorites {
			add(p.Path)
		}
		d.loading = false
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	d.jumpCancel = cancel
	d.loading = true
	publish := a.update
	if publish == nil {
		d.loading = false
		return
	}
	idx := a.folderIndex
	if idx == nil {
		d.loading = false
		return
	}
	go func() {
		timer := time.NewTimer(75 * time.Millisecond)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		var revision uint64
		first := true
		for {
			idx.mu.RLock()
			changed := first || idx.revision != revision
			building := idx.building
			idx.mu.RUnlock()
			if changed {
				paths, version, scanning := idx.search(ctx, query)
				revision = version
				first = false
				publish(func() {
					if ctx.Err() != nil || a.modal != d || a.closed {
						return
					}
					if !slices.Equal(d.jumpResults, paths) || d.jumpQuery != query {
						selected := ""
						if d.jumpQuery == query && d.index >= 0 && d.index < len(d.jumpResults) {
							selected = d.jumpResults[d.index]
						}
						d.jumpResults = paths
						d.jumpQuery = query
						d.index = 0
						if at := slices.Index(paths, selected); at >= 0 {
							d.index = at
						}
						d.jumpList.ScrollIntoView(d.index)
					}
					d.loading = scanning
				})
			}
			if !building {
				return
			}
			timer.Reset(250 * time.Millisecond)
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
			}
		}
	}()
}

func (a *app) jumpOpen(d *dialog, path string) {
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		d.err = "Folder is no longer available"
		return
	}
	a.closeModal()
	a.active = d.jumpPane
	a.load(d.jumpTab, path)
	a.focus = a.active + 1
}

var jumpSearchIcon = sidebarIcon(`<circle cx="10.5" cy="10.5" r="6.5"/><path d="m16 16 5 5"/>`)

// The palette matches the reference's quiet, neutral command window.
func (a *app) jumpColors() palette {
	p := a.colors()
	if !a.light {
		p.window = ui.Hex("#303030")
		p.strip = ui.Hex("#34363a")
		p.hover = ui.Hex("#3d3f43")
		p.pressed = ui.Hex("#484848")
		p.border = ui.Hex("#454545")
		p.text = ui.Hex("#eeeeee")
		p.muted = ui.Hex("#a4a4a4")
	}
	return p
}

func (a *app) jumpView(c *ui.Context, d *dialog) {
	colors := a.jumpColors()
	ui.Column(c).Fill().Gap(0).Children(func() {
		ui.Row(c).FillWidth().Height(62).Shrink(0).PaddingX(20).Gap(12).AlignItems(ui.Center).Border(1, colors.border).BorderWidth(0, 0, 1, 0).Children(func() {
			ui.Icon(c, jumpSearchIcon).Size(22, 22).TextColor(colors.muted).Shrink(0)
			c.Key("jump-search")
			ui.TextInputBase(c, &d.value).Grow(1).MinWidth(0).Height(40).FontSize(18).TextColor(colors.text).FocusRing(false).Label("Search folders").Placeholder("Search folders…").Bind(&d.input).Focus().OnChange(func() { a.searchJump(d) }).OnSubmit(func() {
				if d.jumpQuery == strings.TrimSpace(d.value) && d.index >= 0 && d.index < len(d.jumpResults) {
					a.jumpOpen(d, d.jumpResults[d.index])
				}
			})
		})
		ui.Column(c).Grow(1).MinHeight(0).FillWidth().Padding(10).Children(func() {
			c.Key("jump-results")
			ui.List(c, &d.jumpList, len(d.jumpResults)).Fill().FocusRing(false).ItemKey(func(i int) any { return d.jumpResults[i] }).Rows(func(r ui.ListRow) {
				path := d.jumpResults[r.Index]
				bg := colors.window
				if r.Index == d.index {
					bg = colors.pressed
				}
				ui.ButtonBase(c).FillWidth().Height(44).PaddingX(10).Gap(10).AlignItems(ui.Center).Radius(9).Background(bg).Label(path).Children(func() {
					a.fileIconAt(c, entry{dir: true}, 25, 25)
					ui.Text(c, displayName(path)).FontSize(15).FontWeight(500).TextColor(colors.text).SingleLine().MaxWidth(260).Shrink(0)
					ui.Text(c, filepath.Dir(path)).FontSize(14).TextColor(colors.muted).SingleLine().MinWidth(0).Grow(1)
				}).OnClick(func() {
					if d.jumpQuery == strings.TrimSpace(d.value) {
						a.jumpOpen(d, path)
					}
				})
			})
		})
		ui.Row(c).FillWidth().Height(44).Shrink(0).PaddingX(16).Gap(7).AlignItems(ui.Center).Border(1, colors.border).BorderWidth(1, 0, 0, 0).Children(func() {
			jumpKey(c, "↵", colors)
			ui.Text(c, "Open").FontSize(12).TextColor(colors.muted)
			jumpKey(c, "↑ ↓", colors)
			ui.Text(c, "Select").FontSize(12).TextColor(colors.muted)
			status := ""
			if d.loading {
				status = "Indexing folders…"
			} else if len(d.jumpResults) == 0 {
				status = "No folders found"
			}
			if d.err != "" {
				status = d.err
			}
			ui.Text(c, status).FontSize(12).TextColor(colors.muted).SingleLine().Grow(1)
			jumpKey(c, "esc", colors)
			ui.Text(c, "Close").FontSize(12).TextColor(colors.muted)
		})
	})
	c.OnShortcut(ui.Cmd, ui.KeyR, func() {
		if a.folderIndex != nil {
			a.folderIndex.refresh(folderSearchRoots([]string{filepath.Dir(d.jumpTab.path), d.jumpTab.path}), true)
			a.searchJump(d)
		}
	})
	c.OnShortcut(0, ui.KeyDown, func() {
		if d.index+1 < len(d.jumpResults) {
			d.index++
			d.jumpList.ScrollIntoView(d.index)
		}
	})
	c.OnShortcut(0, ui.KeyUp, func() {
		if d.index > 0 {
			d.index--
			d.jumpList.ScrollIntoView(d.index)
		}
	})
}

func jumpKey(c *ui.Context, key string, colors palette) {
	ui.Text(c, key).FontSize(11).TextColor(colors.muted).Padding(3, 5).Radius(4).Border(1, colors.muted).Shrink(0)
}
