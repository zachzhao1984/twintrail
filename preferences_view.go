package main

import (
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

type preferenceArea struct {
	key, title, description, group string
	fontIndex                      int
}

// Keep the saved font order independent of the order in the navigator.
var preferenceAreas = []preferenceArea{
	{"files", "文件列表", "调整两个窗格中文件和文件夹的显示字体。", "文件浏览", 3},
	{"tableHeader", "列表表头", "调整名称、大小与修改时间等列标题。", "", 5},
	{"breadcrumbs", "路径导航", "调整文件列表上方的文件夹路径。", "", 1},
	{"tabs", "窗格标签页", "调整每个窗格顶部的文件夹标签。", "", 6},
	{"statusBar", "状态栏", "调整文件统计、选中数量与磁盘空间信息。", "", 4},
	{"virtualTabs", "工作区标签", "调整侧边栏中的工作区名称。", "工作区与操作", 2},
	{"actionBar", "操作栏", "调整窗口底部的文件操作按钮。", "", 0},
	{"terminal", "终端", "调整所有标签页中的终端文字，包括已收起的会话。", "其他", 8},
	{"preferences", "设置界面", "调整设置正文的字体；导航和控件保持系统字体。", "", 7},
}

var preferenceMotion = ui.Animation{Duration: 150 * time.Millisecond, Ease: ui.EaseOut}

func (a *app) preferencesView(c *ui.Context) {
	d, colors := a.modal, a.colors()
	area := preferenceAreas[0]
	for _, candidate := range preferenceAreas {
		if candidate.fontIndex == d.index {
			area = candidate
			break
		}
	}
	setting := a.fonts.values()[area.fontIndex]
	theme := *c.Theme()
	theme.FontSize, theme.Spacing, theme.Radius = 13, 4, 6
	theme.Focus = colors.accent
	ui.Themed(c, &theme).Fill().Font("").FontWeight(400).FontSize(13).Children(func() {
		ui.Column(c).Fill().Gap(0).Children(func() {
			ui.Row(c).FillWidth().Height(60).Shrink(0).PaddingX(24).AlignItems(ui.Center).Border(1, colors.border).BorderWidth(0, 0, 1, 0).Children(func() {
				ui.Text(c, "设置").FontSize(20).Bold()
				ui.Box(c).Grow(1)
				ui.Text(c, "⌘ ,").FontSize(12).TextColor(colors.muted)
			})
			ui.Row(c).FillWidth().Grow(1).MinHeight(0).Gap(0).Children(func() {
				ui.Scroll(c).Width(164).Shrink(0).FillHeight().Background(colors.sidebar).Border(1, colors.border).BorderWidth(0, 1, 0, 0).Children(func() {
					ui.Column(c).FillWidth().Padding(12, 10).Gap(2).Children(func() {
						for _, item := range preferenceAreas {
							if item.group != "" {
								ui.Text(c, item.group).FontSize(11).TextColor(colors.muted).Padding(10, 10, 5, 10)
							}
							c.Key(item.key)
							button := ui.ButtonBase(c).FillWidth().Height(32).PaddingX(10).AlignItems(ui.Center).Justify(ui.Start).Radius(6).Label("字体区域：" + item.title)
							bg, fg := colors.sidebar, colors.text
							if item.fontIndex == d.index {
								bg, fg = colors.accent, colors.accentText
							} else if button.Hovered() {
								bg = colors.hover
							}
							button.Background(button.AnimateColor("selection", bg, preferenceMotion)).OnClick(func() {
								d.index = item.fontIndex
								d.preferencesScroll.Y = 0
							}).Children(func() {
								ui.Text(c, item.title).FontSize(13).TextColor(fg).SingleLine()
							})
						}
					})
				})
				c.Key("preferences-content")
				ui.Scroll(c).Grow(1).MinWidth(0).FillHeight().TrackScroll(&d.preferencesScroll).Children(func() {
					ui.Column(c).FillWidth().Padding(24, 28).Gap(0).Children(func() {
						ui.Text(c, area.title).FontSize(20).Bold().ContentTransition(ui.ContentTransition{Animation: &preferenceMotion})
						a.fonts.Preferences.apply(ui.Text(c, area.description)).TextColor(colors.muted).Padding(6, 0, 22, 0)
						a.preferencePreview(c, area, *setting)
						ui.Row(c).FillWidth().Padding(24, 0, 0, 0).Gap(16).Children(func() {
							ui.Column(c).Grow(1).MinWidth(0).Gap(8).Children(func() {
								ui.Text(c, "字体").FontSize(12).TextColor(colors.muted)
								c.Key(area.key + "-family")
								ui.TextInput(c, &setting.Name).FillWidth().Height(34).PaddingX(10).Label(area.key + " font").OnChange(func() {
									if strings.TrimSpace(setting.Name) != "" {
										a.fontsChanged()
									}
								})
							})
							ui.Column(c).Width(136).Shrink(0).Gap(8).Children(func() {
								ui.Text(c, "字号 · pt").FontSize(12).TextColor(colors.muted)
								c.Key(area.key + "-size")
								ui.NumberInput(c, &setting.Size, 8, 48, 1).FillWidth().Height(34).Label(area.key + " size").OnChange(a.fontsChanged)
							})
						})
						hint := "Default 使用系统字体。可输入已安装的字体名称。"
						if area.key == "terminal" {
							hint = "Default 使用系统等宽字体。建议选择等宽字体以保持命令对齐。"
						}
						a.fonts.Preferences.apply(ui.Text(c, hint)).TextColor(colors.muted).Padding(12, 0, 0, 0)
					})
				})
			})
			ui.Row(c).FillWidth().Height(56).Shrink(0).PaddingX(20).Gap(16).AlignItems(ui.Center).Background(colors.strip).Border(1, colors.border).BorderWidth(1, 0, 0, 0).Children(func() {
				ui.Button(c, "恢复默认字体").OnClick(func() { a.fonts = defaultFonts(); a.fontsChanged() })
				ui.Box(c).Grow(1)
				ui.Text(c, "修改即生效 · 自动保存").FontSize(11).TextColor(colors.muted)
				done := ui.PrimaryButton(c, "完成").Width(74).OnClick(func() { a.modal = nil; a.focus = a.active + 1 })
				scale := float32(1)
				if done.Pressed() {
					scale = 0.98
				}
				done.AnimatePose(ui.Pose{Scale: scale}, preferenceMotion)
			})
		})
	})
}

// These are ordinary text and icons, so previews cannot take keyboard focus
// away from the editor or start real terminal sessions / filesystem jobs.
func (a *app) preferencePreview(c *ui.Context, area preferenceArea, font fontSetting) {
	p := a.colors()
	w, _ := c.Size()
	previewWidth := min(float32(760), w-48) - 164 - 56 - 2
	if family, _ := font.familyWeight(); area.key == "terminal" && family == "" {
		font.Name = "monospace"
	}
	ui.Column(c).FillWidth().Height(180).Shrink(0).Background(p.terminal).Border(1, p.border).Radius(8).Clip().Label("字体预览").Children(func() {
		ui.Row(c).FillWidth().Height(32).Shrink(0).PaddingX(12).AlignItems(ui.Center).Background(p.strip).Border(1, p.border).BorderWidth(0, 0, 1, 0).Children(func() {
			ui.Text(c, area.title+"预览").FontSize(11).TextColor(p.muted)
		})
		ui.ScrollBoth(c).FillWidth().Grow(1).MinHeight(0).Children(func() {
			ui.Column(c).MinWidth(previewWidth).Padding(12).Gap(8).Children(func() {
				text := func(value string) {
					font.apply(ui.Text(c, value)).SingleLine().TextColor(p.text)
				}
				switch area.key {
				case "files":
					for i, name := range []string{"Design assets", "项目说明.md", "README.md"} {
						bg := p.terminal
						if i == 0 {
							bg = p.accent.Alpha(0.25)
						}
						ui.Row(c).FillWidth().MinHeight(fontHeight(font, 26)).PaddingX(8).Gap(10).AlignItems(ui.Center).Background(bg).Radius(3).Children(func() {
							a.fileIcon(c, entry{dir: i == 0})
							text(name)
						})
					}
				case "tableHeader":
					ui.Row(c).Gap(32).Padding(8).Children(func() { text("名称"); text("大小"); text("修改时间") })
					ui.Text(c, "Projects     —     Today at 10:30").FontSize(12).TextColor(p.muted).Padding(8)
				case "breadcrumbs":
					ui.Row(c).Gap(12).Padding(8).Children(func() { a.fileIcon(c, entry{dir: true}); text("Projects  /  Twintrail  /  Assets") })
				case "tabs":
					ui.Row(c).Gap(4).Children(func() {
						for _, name := range []string{"Twintrail", "Design assets"} {
							font.apply(ui.Text(c, name)).Padding(8, 12).Background(p.strip).Radius(5).SingleLine()
						}
					})
				case "virtualTabs":
					for i, name := range []string{"Twintrail", "Design assets"} {
						bg := p.strip
						if i == 0 {
							bg = p.accent.Alpha(0.25)
						}
						ui.Row(c).Padding(8, 12).Gap(10).AlignItems(ui.Center).Background(bg).Radius(5).Children(func() {
							ui.Icon(c, iconWorkspace).Size(16, 16).TextColor(p.dirIcon)
							text(name)
						})
					}
				case "statusBar":
					text("24 KB in 3 files, also 2 folders")
					text("2 selected  |  218 GB available")
				case "actionBar":
					ui.Row(c).Gap(10).Children(func() {
						for _, label := range []string{"F5 Copy", "F6 Move", "F7 New Folder"} {
							font.apply(ui.Text(c, label)).Padding(8, 12).Background(p.strip).Radius(4).SingleLine()
						}
					})
				case "terminal":
					font.apply(ui.Text(c, "~/Projects/Twintrail")).TextColor(p.dirIcon).SingleLine()
					text("❯ git status --short")
					text(" M  README.md")
				case "preferences":
					text("让常用的文件与命令，保持清晰。")
					text("字体设置  ·  立即生效")
				}
			})
		})
	})
}
