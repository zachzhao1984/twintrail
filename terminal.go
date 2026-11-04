package main

import (
	"fmt"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/plugins/updater"
	"github.com/egoist/mygo/ui"
)

// defaultTerminalPercent is the share of the room above the terminal that the
// panes take, which the line above it moves.
const defaultTerminalPercent float32 = 60

type terminalSession interface {
	View(*ui.Context) ui.Element
	SetDir(string, uint64) error
	SetFont(fontSetting)
	SetAppearance(light bool)
	Close() error
}
type tabTerminal struct {
	visible, starting bool
	generation        int
	revision          uint64
	session           terminalSession
	err               string
	handle            ui.Handle
}
type nativeTerminal struct {
	term   *terminal.Terminal
	bridge *cwdBridge
	font   fontSetting
}

func (n *nativeTerminal) View(c *ui.Context) ui.Element        { return terminal.View(c, n.term) }
func (n *nativeTerminal) SetDir(path string, rev uint64) error { return n.bridge.setDir(path, rev) }
func (n *nativeTerminal) SetAppearance(light bool) {
	n.term.SetTheme(terminalTheme(light), terminalTheme(false))
}

// terminalTheme draws the shell in the look the app draws: the panel behind it,
// its text, its cursor and what it selects, and the sixteen colors programs
// paint with, whose blue and yellow are the app's accent and mark, so that a
// shell reads as part of the window rather than as another app's.
func terminalTheme(light bool) *terminal.Theme {
	p, t := darkPalette, terminal.DarkTheme()
	if light {
		p, t = lightPalette, terminal.LightTheme()
	}
	t.Background, t.Foreground = p.terminal, p.terminalText
	t.Cursor, t.CursorText = p.accent, p.terminal
	t.Selection = p.accent.Alpha(0.35)
	t.Palette[3], t.Palette[11] = p.mark, p.mark
	t.Palette[4], t.Palette[12] = p.accent, p.accent
	return t
}
func (n *nativeTerminal) Close() error { err := n.term.Close(); n.bridge.close(); return err }
func newTerminal(path string) (terminalSession, error) {
	bridge, err := newCWDBridge(path)
	if err != nil {
		return nil, err
	}
	term, err := terminal.New(terminal.Options{Command: []string{"/bin/zsh", "-l"}, Dir: path, Env: bridge.env(), Theme: terminalTheme(false), DarkTheme: terminalTheme(false), Font: terminal.Font{Size: 12}, Cursor: terminal.CursorBar})
	if err != nil {
		bridge.close()
		return nil, err
	}
	return &nativeTerminal{term: term, bridge: bridge, font: defaultFonts().Terminal}, nil
}
func (a *app) toggleTerminal() {
	if a.modal != nil {
		return
	}
	t := a.current()
	t.terminal.visible = !t.terminal.visible
	if t.terminal.visible {
		a.focus = a.active + 3
		a.ensureTerminal(t)
	} else {
		a.focus = a.active + 1
	}
	a.refreshTerminalMenu()
}
func (a *app) ensureTerminal(t *tab) {
	s := &t.terminal
	if s.session != nil || s.starting {
		return
	}
	s.generation++
	generation := s.generation
	s.starting = true
	s.err = ""
	path := t.path
	factory := a.newTerminal
	if factory == nil {
		factory = newTerminal
	}
	a.terminalJobs.Add(1)
	a.job(func() func() {
		defer a.terminalJobs.Done()
		session, err := factory(path)
		if session != nil {
			a.terminalSessions.Store(session, struct{}{})
		}
		return func() {
			if a.closed || generation != s.generation {
				if session != nil {
					a.releaseTerminal(session)
				}
				return
			}
			s.starting = false
			if err != nil {
				s.err = err.Error()
				return
			}
			s.session = session
			session.SetFont(a.fonts.Terminal)
			session.SetAppearance(a.light)
			a.syncTerminal(t)
		}
	})
}
func (a *app) syncTerminal(t *tab) {
	s := &t.terminal
	if s.session == nil {
		return
	}
	s.revision++
	revision := s.revision
	session := s.session
	path := t.path
	generation := s.generation
	a.job(func() func() {
		err := session.SetDir(path, revision)
		return func() {
			if !a.closed && generation == s.generation && revision == s.revision && err != nil {
				s.err = err.Error()
			}
		}
	})
}
func (a *app) closeTerminal(t *tab) {
	s := &t.terminal
	s.generation++
	session := s.session
	s.session = nil
	s.starting = false
	s.visible = false
	if session != nil {
		a.releaseTerminal(session)
	}
}
func (a *app) releaseTerminal(session terminalSession) {
	a.terminalJobs.Add(1)
	a.job(func() func() {
		defer a.terminalJobs.Done()
		session.Close()
		a.terminalSessions.Delete(session)
		return func() {}
	})
}

// Called after the event loop ends, including sessions whose creation
// finished after the window closed and could not publish a UI callback.
func (a *app) finishTerminals() {
	a.terminalJobs.Wait()
	a.terminalSessions.Range(func(key, value any) bool { key.(terminalSession).Close(); a.terminalSessions.Delete(key); return true })
}
func (a *app) closeTerminals() {
	for i := range a.panes {
		for _, t := range a.panes[i].tabs {
			a.closeTerminal(t)
		}
	}
}
func (a *app) terminalView(c *ui.Context) {
	t := a.current()
	s := &t.terminal
	c.Key(fmt.Sprintf("terminal-%p", t))
	colors := a.colors()
	ui.Column(c).Fill().Background(colors.terminal).Children(func() {
		ui.Row(c).Height(24).PaddingX(6).Gap(8).AlignItems(ui.Center).Background(colors.strip).Children(func() {
			ui.Text(c, "Terminal").Bold()
			ui.Text(c, t.path).Grow(1).MinWidth(0).SingleLine().TextColor(c.Theme().TextMuted)
			ui.ButtonBase(c).Label("Close Terminal").Children(func() { ui.Text(c, "×").FontSize(16) }).OnClick(a.toggleTerminal)
		})
		if s.starting {
			ui.Text(c, "Starting terminal…").Padding(8)
		} else if s.err != "" {
			ui.Text(c, s.err).Padding(8).TextColor(c.Theme().Danger)
			ui.Button(c, "Retry Terminal").OnClick(func() { a.closeTerminal(t); t.terminal.visible = true; a.ensureTerminal(t) })
		} else if s.session != nil {
			c.Key("shell")
			s.session.View(c).Grow(1).FillWidth().Bind(&s.handle).FocusBind(&a.focus, a.active+3).OnShortcut(ui.Cmd, ui.KeyBackquote, a.toggleTerminal)
		}
	})
}

// updateMenuItem builds the updater's "Check for Updates…" item. Building it
// reads the user's language, which asks the app's event loop, so it is only
// safe once the loop runs: the app builds its menu from WhenReady. A test
// has no loop, and overrides this with an item that does not read one.
var updateMenuItem = updater.MenuItem

func (a *app) applicationMenu() *mygo.Menu {
	a.terminalMenu = &mygo.MenuItem{ID: "toggle-terminal", Label: "Toggle Terminal", Type: mygo.MenuItemCheckbox, Accelerator: "Cmd+`", Click: func(item *mygo.MenuItem, w *mygo.Window) {
		a.toggleTerminal()
		if w != nil {
			w.Invalidate()
		}
	}}
	return mygo.NewMenu([]*mygo.MenuItem{{Role: mygo.RoleAppMenu, Submenu: []*mygo.MenuItem{{Role: mygo.RoleAbout}, updateMenuItem(), mygo.Separator(), a.preferenceMenu(), mygo.Separator(), {Role: mygo.RoleServices}, mygo.Separator(), {Role: mygo.RoleHide}, {Role: mygo.RoleHideOthers}, {Role: mygo.RoleUnhide}, mygo.Separator(), {Role: mygo.RoleQuit}}}, a.applicationFileMenu(), {Role: mygo.RoleEditMenu}, {Label: "View", Submenu: []*mygo.MenuItem{a.appearanceMenu(), mygo.Separator(), a.actionBarMenu(), a.terminalMenu, mygo.Separator(), {Role: mygo.RoleToggleFullScreen}}}, {Role: mygo.RoleWindowMenu}})
}

func (a *app) actionBarMenu() *mygo.MenuItem {
	return &mygo.MenuItem{ID: "toggle-actionbar", Label: "Toggle ActionBar", Type: mygo.MenuItemCheckbox, Checked: !a.actionBarHidden, Click: func(item *mygo.MenuItem, w *mygo.Window) {
		a.actionBarHidden = !a.actionBarHidden
		item.SetChecked(!a.actionBarHidden)
		a.persistSettings(true)
		if w != nil {
			w.Invalidate()
		}
	}}
}

// appearanceMenu checks the selected preference, independently of its resolved color.
func (a *app) appearanceMenu() *mygo.MenuItem {
	choose := func(source mygo.ThemeSource) func(*mygo.MenuItem, *mygo.Window) {
		return func(item *mygo.MenuItem, w *mygo.Window) {
			a.setAppearance(source)
			if w != nil {
				w.Invalidate()
			}
		}
	}
	return &mygo.MenuItem{Label: "Appearance", Submenu: []*mygo.MenuItem{
		{ID: "appearance-system", Label: "System", Type: mygo.MenuItemRadio, Checked: a.appearance == mygo.ThemeSystem, Click: choose(mygo.ThemeSystem)},
		{ID: "appearance-light", Label: "Light", Type: mygo.MenuItemRadio, Checked: a.appearance == mygo.ThemeLight, Click: choose(mygo.ThemeLight)},
		{ID: "appearance-dark", Label: "Dark", Type: mygo.MenuItemRadio, Checked: a.appearance == mygo.ThemeDark, Click: choose(mygo.ThemeDark)},
	}}
}

// applyTerminalAppearance draws the shells that are running in the look the app
// now draws, as the look changes.
func (a *app) applyTerminalAppearance() {
	for i := range a.panes {
		for _, t := range a.panes[i].tabs {
			if t.terminal.session != nil {
				t.terminal.session.SetAppearance(a.light)
			}
		}
	}
}

func (a *app) setAppearance(source mygo.ThemeSource) {
	if a.appearance == source {
		return
	}
	a.appearance = source
	a.applyAppearance()
	a.persistSettings(true)
}

// applyAppearance resolves System only after the native app is ready.
func (a *app) applyAppearance() {
	if a.appearance == mygo.ThemeSystem {
		followSystemAppearance()
		a.updateAppearance(!systemAppearanceIsDark())
	} else {
		light := a.appearance == mygo.ThemeLight
		setSystemAppearance(light)
		a.updateAppearance(light)
	}
}

func (a *app) updateAppearance(light bool) {
	if a.light == light {
		return
	}
	a.light = light
	a.applyTerminalAppearance()
}

// fileMenu replaces RoleFileMenu, whose Close Window item owns Cmd+W on
// macOS and would close the window before the view's tab shortcut ran.
func (a *app) applicationFileMenu() *mygo.MenuItem {
	a.closeTabMenu = &mygo.MenuItem{ID: "close-tab", Label: "Close Tab", Accelerator: "Cmd+W", Disabled: len(a.panes[a.active].tabs) <= 1, Click: func(item *mygo.MenuItem, w *mygo.Window) {
		p := &a.panes[a.active]
		a.closeTab(p, p.current)
		if w != nil {
			w.Invalidate()
		}
	}}
	return &mygo.MenuItem{Label: "File", Submenu: []*mygo.MenuItem{
		{ID: "new-tab", Label: "New Tab", Accelerator: "Cmd+T", Click: func(item *mygo.MenuItem, w *mygo.Window) {
			a.newTab()
			if w != nil {
				w.Invalidate()
			}
		}},
		a.closeTabMenu,
		mygo.Separator(),
		{Role: mygo.RoleClose, Accelerator: "Shift+Cmd+W"},
	}}
}

func (a *app) refreshFileMenu() {
	if a.closeTabMenu == nil {
		return
	}
	enabled := len(a.panes[a.active].tabs) > 1
	if a.closeTabMenu.IsEnabled() != enabled {
		a.closeTabMenu.SetEnabled(enabled)
	}
}
func (a *app) refreshTerminalMenu() {
	if a.terminalMenu != nil && a.terminalMenu.IsChecked() != a.current().terminal.visible {
		a.terminalMenu.SetChecked(a.current().terminal.visible)
	}
}

func (n *nativeTerminal) SetFont(f fontSetting) {
	if n.font == f {
		return
	}
	family, weight := f.familyWeight()
	n.term.SetFont(terminal.Font{Family: family, Weight: weight, Size: float32(f.Size)})
	n.font = f
}
