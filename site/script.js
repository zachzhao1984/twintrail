const images = { dark: "assets/app-dark.webp", light: "assets/app-light.webp" };
const copy = {
  "zh-CN": {
    skip: "跳到主要内容",
    navWorkflow: "操作方式",
    navSource: "源代码",
    navDownload: "下载 Twintrail",
    platform: "为 macOS 而生的原生双栏文件管理器",
    heroTitle: "操作流畅，<br>思路不断。",
    heroLead:
      "项目在左，交付在右。熟悉的快捷键、独立的标签页，让每一步操作都顺着你的思路。",
    download: "下载 macOS 版",
    otherPlatforms: "Windows 与 Linux 下载",
    terminalTitle: "文件与命令行，<br>在同一个工作现场。",
    terminalLead:
      "浏览文件，运行命令，再回到文件。集成 Terminal 就在双栏下方，少一次窗口切换，多一点连贯。",
    terminalCaption: "macOS 开发预览，终端内容为示例；点击查看大图",
    terminalShortcut: "显示 / 隐藏 Terminal",
    terminalFollowTitle: "跟随当前目录",
    terminalFollowBody:
      "在文件面板中进入文件夹，终端工作目录随之同步，接着执行下一条命令。",
    terminalTabsTitle: "每个标签页，各自继续",
    terminalTabsBody:
      "标签页保留独立终端会话。收起面板后会话仍在，回到标签页就能继续。",
    terminalSizeTitle: "留出刚好的空间",
    terminalSizeBody:
      "拖动终端上边缘调整高度。明暗配色随应用外观切换，让命令行融入工作区。",
    terminalImage:
      "Twintrail 双栏文件列表下方展开集成 Terminal，显示示例项目的 git status 与 ls 命令",
    terminalFullImage: "查看集成 Terminal 完整截图",
    downloadRelease: "公开版本 v1.2.0 · 按系统与芯片选择下载",
    macChip: "Apple Silicon 与 Intel 通用",
    windowsPackage: "安装程序 .exe",
    portablePackage: "便携版 .zip",
    debianPackage: "Debian / Ubuntu .deb",
    archivePackage: "压缩包 .tar.gz",
    chipHelp: "x64 适用于 Intel / AMD 64 位芯片；ARM64 适用于 ARM 芯片。",
    checksums: "SHA256 校验文件",
    release: "公开版本 v1.2.0",
    releaseNotes: "发行说明",
    keysCaption: "从熟悉的按键开始。",
    preview: "开发预览，使用示例工作区",
    promisePanes: "两个位置，同一个视野。",
    promiseKeys: "为键盘操作设计。",
    promiseNative: "原生界面，日常工具。",
    workflowTitle: "你的双手，<br>已经知道怎么做。",
    workflowLead:
      "预览、切换、打开终端。常用动作就在指尖，不必为了下一步操作打断眼前的工作。",
    commandPreview: "即刻预览",
    subPreview: "用 Quick Look 看清内容",
    commandTabs: "保留上下文",
    subTabs: "在当前目录打开新标签页",
    commandTerminal: "继续用命令行",
    subTerminal: "终端跟随当前文件夹",
    illustration: "快捷键交互示意",
    nativeTitle: "工作有两面，<br>工具不必分心。",
    nativeLead:
      "用工作区保存两侧的标签页与目录布局。回到项目时，让熟悉的位置一起回来。",
    nativeBody:
      "浅色、深色，选择适合你的工作环境。标签页保留各自的终端，让文件与命令行始终在手边。",
    details: "了解工作区与标签页",
    dark: "深色",
    light: "浅色",
    closingTitle: "把工作留在顺手的位置。",
    closingLead: "为你的下一个项目，留出清晰的双栏。",
    allReleases: "全部平台与版本",
    backTop: "回到顶部",
    heroImage: "Twintrail 深色双栏界面，左侧项目文件夹与右侧交付目录同时显示",
    appearanceImage: "Twintrail 开发预览",
    demo: "快捷键交互示意，可选择空格、Command T 或终端快捷键",
    appearance: "明暗外观预览",
  },
  en: {
    skip: "Skip to content",
    navWorkflow: "Workflow",
    navSource: "Source code",
    navDownload: "Get Twintrail",
    platform: "A native dual-pane file manager for macOS",
    heroTitle: "Stay in<br>your flow.",
    heroLead:
      "Projects on the left. Deliverables on the right. Familiar shortcuts and independent tabs keep every next step close at hand.",
    download: "Download for macOS",
    otherPlatforms: "Windows & Linux downloads",
    terminalTitle: "Your files and shell.<br>One working space.",
    terminalLead:
      "Browse a folder, run a command, return to your files. The embedded terminal sits beneath both panes, keeping the next step in the same window.",
    terminalCaption:
      "macOS development preview with example output; click to enlarge",
    terminalShortcut: "Show / hide Terminal",
    terminalFollowTitle: "Follow the folder",
    terminalFollowBody:
      "Navigate to a folder in the file pane and the terminal’s working directory follows, ready for your next command.",
    terminalTabsTitle: "A session for every tab",
    terminalTabsBody:
      "Each tab keeps its own terminal session. Hiding the panel keeps the session running, ready when you return.",
    terminalSizeTitle: "Make room as you go",
    terminalSizeBody:
      "Drag the top edge to resize the terminal. Its light and dark colors follow the app’s appearance, keeping your workspace together.",
    terminalImage:
      "Twintrail with its embedded terminal below the file panes, showing example git status and ls commands",
    terminalFullImage: "View the full embedded Terminal screenshot",
    downloadRelease: "Public release v1.2.0 · Choose your system and chip",
    macChip: "Universal for Apple Silicon & Intel",
    windowsPackage: "Installer .exe",
    portablePackage: "Portable .zip",
    debianPackage: "Debian / Ubuntu .deb",
    archivePackage: "Archive .tar.gz",
    chipHelp: "x64 is for 64-bit Intel / AMD chips; ARM64 is for ARM chips.",
    checksums: "SHA256 checksums",
    release: "Public release v1.2.0",
    releaseNotes: "Release notes",
    keysCaption: "Start with the keys you know.",
    preview: "Development preview with example files",
    promisePanes: "Two places. One clear view.",
    promiseKeys: "Made for your keyboard.",
    promiseNative: "Native to your everyday.",
    workflowTitle: "Your hands already<br>know what to do.",
    workflowLead:
      "Preview a file, open a tab, drop into a terminal. Keep familiar actions close, and your attention on the work in front of you.",
    commandPreview: "A closer look",
    subPreview: "Preview with native Quick Look",
    commandTabs: "Keep your context",
    subTabs: "Open a new tab in the current folder",
    commandTerminal: "Take it to the terminal",
    subTerminal: "Your shell follows the current folder",
    illustration: "Interactive shortcut illustration",
    nativeTitle: "Two sides of your work.<br>One place to focus.",
    nativeLead:
      "Save both panes, their tabs and folder layouts in a workspace. Come back to a project with everything where you left it.",
    nativeBody:
      "Light or dark, choose a view that suits your desk. Each tab keeps its own terminal, with your files and shell close at hand.",
    details: "Explore workspaces and tabs",
    dark: "Dark",
    light: "Light",
    closingTitle: "Keep your work within reach.",
    closingLead: "A clearer pair of panes for your next project.",
    allReleases: "All platforms and releases",
    backTop: "Back to top",
    heroImage:
      "Twintrail dark dual-pane interface showing project folders and deliverables side by side",
    appearanceImage: "Twintrail development preview",
    demo: "Interactive shortcut illustration: select Space, Command T or the terminal shortcut",
    appearance: "Light and dark appearance preview",
  },
};
const states = {
  preview: {
    "zh-CN": [
      "先看清，再继续。",
      "选中文件，按下空格键。Quick Look 原生预览面板会在需要的时候出现。",
      "Brief.pdf",
    ],
    en: [
      "Look first. Keep going.",
      "Select a file and press Space. The native Quick Look panel appears when you need a closer look.",
      "Brief.pdf",
    ],
  },
  tabs: {
    "zh-CN": [
      "换个位置，保留思路。",
      "在当前目录打开新标签页。项目、素材、交付，各自留在熟悉的位置。",
      "⌘ T",
    ],
    en: [
      "Another tab. Same train of thought.",
      "Open a new tab at the current folder. Keep projects, assets and deliverables in their own familiar places.",
      "⌘ T",
    ],
  },
  terminal: {
    "zh-CN": [
      "目录就在眼前，命令就在手边。",
      "显示当前标签页的独立终端。目录导航成功后，终端会跟随到同一个位置。",
      "⌘ `",
    ],
    en: [
      "See the folder. Reach for the shell.",
      "Show the current tab’s own terminal. After successful navigation, its working directory follows the file pane.",
      "⌘ `",
    ],
  },
};
let language = "zh-CN",
  activeCommand = "preview";
function showCommand(command) {
  activeCommand = command;
  document
    .querySelectorAll("[data-command]")
    .forEach((b) =>
      b.setAttribute("aria-pressed", String(b.dataset.command === command)),
    );
  const state = states[command][language];
  document.getElementById("demo-title").textContent = state[0];
  document.getElementById("demo-description").textContent = state[1];
  document.getElementById("demo-hint").textContent = state[2];
  const art = document.getElementById("demo-art");
  art.className = command === "terminal" ? "terminal-example" : "file-glyph";
  art.innerHTML =
    command === "terminal"
      ? "<span>~/Projects</span><br>$ git status"
      : command === "tabs"
        ? "⌘ T"
        : "Aa";
}
function setLanguage(lang) {
  language = lang;
  document.title =
    lang === "en"
      ? "Twintrail — Stay in your flow."
      : "Twintrail — 操作流畅，思路不断。";
  document.querySelector('meta[name="description"]').content =
    lang === "en"
      ? "A native dual-pane file manager for macOS. Familiar shortcuts, tabs, workspaces and an embedded terminal."
      : "Twintrail：为 macOS 而生的原生双栏文件管理器，提供熟悉的快捷键、标签页、工作区和内置终端。";
  try {
    localStorage.setItem("twintrail-language", lang);
  } catch (_) {}
  document.documentElement.lang = lang;
  document
    .querySelectorAll("[data-i18n]")
    .forEach((e) => (e.textContent = copy[lang][e.dataset.i18n]));
  document
    .querySelectorAll("[data-i18n-html]")
    .forEach((e) => (e.innerHTML = copy[lang][e.dataset.i18nHtml]));
  document
    .querySelectorAll("[data-alt]")
    .forEach((e) => (e.alt = copy[lang][e.dataset.alt]));
  document
    .querySelectorAll("[data-aria]")
    .forEach((e) => e.setAttribute("aria-label", copy[lang][e.dataset.aria]));
  document
    .querySelectorAll("[data-language]")
    .forEach((e) =>
      e.setAttribute("aria-pressed", String(e.dataset.language === lang)),
    );
  document
    .querySelector(".wordmark")
    .setAttribute(
      "aria-label",
      lang === "en" ? "Twintrail home" : "Twintrail 首页",
    );
  document
    .querySelector(".navlinks")
    .setAttribute("aria-label", lang === "en" ? "Main navigation" : "主导航");
  document
    .querySelector(".lang")
    .setAttribute("aria-label", lang === "en" ? "Language" : "语言");
  showCommand(activeCommand);
}
document
  .querySelectorAll("[data-language]")
  .forEach((b) =>
    b.addEventListener("click", () => setLanguage(b.dataset.language)),
  );
document
  .querySelectorAll("[data-command]")
  .forEach((b) =>
    b.addEventListener("click", () => showCommand(b.dataset.command)),
  );
document.querySelector(".walkthrough").addEventListener("keydown", (e) => {
  if (e.key === " " && e.target === e.currentTarget) {
    e.preventDefault();
    showCommand("preview");
  }
  if (e.key === "ArrowDown" || e.key === "ArrowUp") {
    e.preventDefault();
    const order = ["preview", "tabs", "terminal"];
    showCommand(
      order[
        (order.indexOf(activeCommand) + (e.key === "ArrowDown" ? 1 : 2)) % 3
      ],
    );
  }
});
document.querySelectorAll("[data-appearance]").forEach((b) =>
  b.addEventListener("click", () => {
    document.getElementById("appearance-image").src =
      images[b.dataset.appearance];
    document.querySelectorAll("[data-appearance]").forEach((x) => {
      const on = x === b;
      x.classList.toggle("active", on);
      x.setAttribute("aria-pressed", String(on));
    });
  }),
);
let preferredLanguage = "zh-CN";
try {
  preferredLanguage =
    localStorage.getItem("twintrail-language") || preferredLanguage;
} catch (_) {}
const requestedLanguage = new URLSearchParams(location.search).get("lang");
if (requestedLanguage)
  preferredLanguage = requestedLanguage === "en" ? "en" : "zh-CN";
setLanguage(preferredLanguage === "en" ? "en" : "zh-CN");
