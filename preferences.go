package main

import (
	"math"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

type fontSetting struct {
	Name string
	Size float64
}
type fontSettings struct {
	ActionBar   fontSetting `json:"actionBar"`
	Breadcrumbs fontSetting `json:"breadcrumbs"`
	VirtualTabs fontSetting `json:"virtualTabs"`
	Files       fontSetting `json:"files"`
	StatusBar   fontSetting `json:"statusBar"`
	TableHeader fontSetting `json:"tableHeader"`
	Tabs        fontSetting `json:"tabs"`
	Preferences fontSetting `json:"preferences"`
	Terminal    fontSetting `json:"terminal"`
}

func defaultFonts() fontSettings {
	return fontSettings{fontSetting{"Default", 12}, fontSetting{"Default", 14}, fontSetting{"Default Bold", 12}, fontSetting{"Default", 14}, fontSetting{"Default Light", 11}, fontSetting{"Default", 11}, fontSetting{"Default", 12}, fontSetting{"Monaco", 14}, fontSetting{"Default", 12}}
}
func (f fontSetting) familyWeight() (string, int) {
	family := strings.TrimSpace(f.Name)
	weight := 400
	switch family {
	case "Default", "":
		family = ""
	case "Default Bold":
		family = ""
		weight = 700
	case "Default Light":
		family = ""
		weight = 300
	}
	return family, weight
}
func (f fontSetting) apply(e ui.Element) ui.Element {
	family, weight := f.familyWeight()
	return e.Font(family).FontWeight(weight).FontSize(float32(f.Size))
}
func (f *fontSettings) values() []*fontSetting {
	return []*fontSetting{&f.ActionBar, &f.Breadcrumbs, &f.VirtualTabs, &f.Files, &f.StatusBar, &f.TableHeader, &f.Tabs, &f.Preferences, &f.Terminal}
}
func (f fontSettings) normalized() fontSettings {
	defaults := defaultFonts()
	d := defaults.values()
	for i, v := range f.values() {
		if strings.TrimSpace(v.Name) == "" {
			v.Name = d[i].Name
		}
		if math.IsNaN(v.Size) || math.IsInf(v.Size, 0) || v.Size < 8 || v.Size > 48 {
			v.Size = d[i].Size
		}
	}
	return f
}
func fontHeight(f fontSetting, minimum float32) float32 {
	return max(minimum, float32(math.Ceil(f.Size*1.4+4)))
}
func (a *app) openPreferences() {
	a.focus = 0
	a.modal = &dialog{kind: "Preference", open: true, index: 3}
}
func (a *app) preferenceMenu() *mygo.MenuItem {
	return &mygo.MenuItem{ID: "preferences", Label: "Settings…", Accelerator: "Cmd+,", Click: func(item *mygo.MenuItem, w *mygo.Window) {
		a.openPreferences()
		if w != nil {
			w.Invalidate()
		}
	}}
}

// Serialize writes and reject snapshots superseded by a more recent edit.
type sessionWriter struct {
	mu      sync.Mutex
	version atomic.Uint64
	save    func(session)
}

func (a *app) persistSettings(background bool) {
	snapshot := a.snapshotSession()
	revision := a.settingsWriter.version.Add(1)
	work := func() {
		a.settingsWriter.mu.Lock()
		defer a.settingsWriter.mu.Unlock()
		if revision == a.settingsWriter.version.Load() {
			save := a.settingsWriter.save
			if save == nil {
				save = saveSession
			}
			save(snapshot)
		}
	}
	if background {
		go work()
	} else {
		work()
	}
}

// Apply to existing sessions, including terminals hidden in other tabs.
func (a *app) fontsChanged() {
	for _, pane := range a.panes {
		for _, tab := range pane.tabs {
			if tab.terminal.session != nil {
				tab.terminal.session.SetFont(a.fonts.Terminal)
			}
		}
	}
	a.persistSettings(true)
}
