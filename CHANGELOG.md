# Changelog

## Unreleased

- ⌘⌫ moves the selection to the Trash, as Finder's does: it opens the dialog
  Move to Trash opens, which the confirmation empties into the Trash.

- The entries' menus gain Open With…, which opens the picker over the entry it
  was opened on, and Compress, which packs the entries into one zip beside the
  first of them, named for it, whose folders walk their entries; an existing
  zip is left alone. Every view's entries carry both.

- Updated myGO to v0.7.4. The migration to checked elements names nothing this
  app changes: `go tool mygo migrate-ui` rewrites no file, and `go tool mygo
  vet` and the tests pass as they did.

- A tab drags between the panes: its strip takes a tab from either pane, which
  moves across and becomes the current tab there, and the pane it left keeps
  its other tabs. The strips are always shown now, where the single tab of a
  pane had none to show.

- Every view's folders carry a context menu, and every menu gains Pin to
  Favorites, which puts the folder in the sidebar's Favorites; a folder already
  pinned offers no second pin. The icon view's cells and the column view's rows
  act on the entry they were opened on, whose menu the marked entries join.

- Each tab shows its folder as a list, as icons in a grid or as columns, as
  Finder's list, icon and column views do: the three buttons at the right end
  of the breadcrumb row pick the mode, the one the tab is in carrying the
  pressed look. The list keeps its Name, Size and Modified header, which the
  other views drop; the columns open a folder's column beside as it is chosen,
  and the mode is saved with the tab.

- The file lists' header centers its labels, which the Name cell's baseline
  layout sat a little high beside the Size and Modified cells.

- A redesigned pair of appearances, light and dark: one blue accent, graphite
  panes in the dark look and white ones in the light, and the tabs' band, the
  header, the status bar, the sidebar, the alternate rows, the marked names and
  the terminal all take their colors from it. **View → Appearance** picks
  Light or Dark, and the window opens as it was left. The system's own parts
  follow it too: the window's buttons, its menus and the panels it opens are
  drawn in the look chosen, where before they took the system's.
- The labels of the file table's header sit on one baseline: the Name cell
  carries the first column with its sort arrow, and the Size and Modified cells
  line up on it, rather than centering their own boxes.
- The line above the terminal drags to resize it, as the lines beside the panes
  and the sidebar do; the share of the room it takes is remembered.
- The terminal follows the look as well: its panel, text, cursor and selection
  take the palette, and so do the sixteen colors programs paint with, whose
  blue and yellow are the app's accent and mark. A shell that is already
  running changes with the look, as one started after it opens in it.

- The sidebar's right edge drags to resize it, as the line between the panes
  does; it keeps a readable width, and the app remembers it.
- The favorites draw their folders with a star in them, which tells them from
  the recents' plain folders at a glance.
- The sidebar's items read a step under the file rows beside them, and follow
  the File font setting; the icons of the sidebar fill their box without the
  weight of the file lists' icons.
- A button at the top of the sidebar, drawn as a sidebar, folds it away over a
  moment: its items slide out from under the panes while their room grows, and
  the button in their strip brings it back at the width it had. The window
  opens as it was left.
- Favorites, Recents and Locations carry a button at the right of their
  headers that shows and hides their items, turning as it does; a click on
  the header does the same. The items fold away under the title over a moment,
  fading as the sections below slide up, and a collapsed section stays
  collapsed. The app's name no longer heads the sidebar: the sections start
  under the buttons that fold the whole sidebar away.
- A Finder-style sidebar with Workspaces, Favorites, Recents and system
  Locations. Workspaces remember the tabs in both panes, the current tab, the
  active pane and the split position; the **+** beside the Workspaces header
  captures the current layout and switching restores it.
- Updated myGO to v0.7.3, along with purego, x/image, x/net and x/text.
- Tabs can be closed with the × button on each tab or its Close Tab context
  menu, besides ⌘W. Closing a tab also closes its terminal.
- ⌘W closes the current tab instead of the window, as the File menu's Close
  Window item used to claim that shortcut. Shift+⌘W now closes the window;
  the File menu also offers New Tab.
- Workspaces can be renamed or deleted from their sidebar context menu; a
  window keeps at least one workspace. Favorites and recents can be removed
  from their context menus, and removing every favorite does not restore the
  defaults on the next launch. Deleting a workspace other than the last one
  no longer closes the window: the change is applied after the frame that
  drew the sidebar, instead of while it is being built.

## 0.1.2

- No change to the app itself. This release re-runs the release build with
  a fix to the workflow that produced the 0.1.1 disk image, which was
  discarded before it could be attached to the release.

## 0.1.1

- In-app updates. Twintrail now checks for a new version in the background
  and offers to install it from a "Check for Updates…" item in the app
  menu, with the release notes of the new version. The update window is
  drawn in the app's own native UI, so no webview is involved.

## 0.1.0

The first release of Twintrail, a native macOS dual-pane file manager.

- A tab's list, icon and column views show the folder's entries alone, without
  the parent folder row the older views began with; Backspace still goes up, as
  the breadcrumbs do.
, name ascending or descending, hidden-file toggle, and a
  name substring filter.
- Copy, move, rename, create nested directories, and move to the macOS Trash,
  with confirmation dialogs and error display. Copying preserves symlinks,
  refuses to overwrite an existing target, and refuses to copy or move a
  directory into itself. Cross-volume moves report an error rather than
  falling back to a copy and a delete.
- File operations and directory reads run in the background, and a stale
  directory read is discarded rather than shown.
- Open with the system default application, Quick Look preview, reveal in
  Finder, and an external Terminal.
- A native context menu on every row, acting on the row it was opened on
  rather than on whichever row the cursor was on. Its Copy Path item writes
  the absolute paths of the marked entries to the pasteboard, one per line,
  each quoted so that pasting into a terminal is safe.
- Opening the context menu moves the keyboard highlight to that row as the
  menu appears, instead of leaving it on the old row until the menu closes.
- Nine font settings for the action bar, breadcrumbs, virtual tabs, file
  list, status bar, table header, tabs, preferences and terminal, applied as
  they are changed and saved to the window state.
- Open With, listing the applications of the selected files through
  NSWorkspace, with icons rendered in the background and cached.
- An embedded terminal per tab, with its own zsh login shell and scrollback,
  toggled from View, whose working directory follows navigation without
  injecting commands into its input.
- In-app updates, checked in the background and on demand from the app menu.