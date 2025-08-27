package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	iauth "github.com/arvid-berndtsson/lazyms/internal/auth"
	"github.com/arvid-berndtsson/lazyms/internal/config"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type keymap struct {
	NextPane, PrevPane, FocusLeft, FocusRight, AuthMenu, Shortcuts, Quit key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.NextPane, k.PrevPane, k.FocusLeft, k.FocusRight, k.AuthMenu, k.Shortcuts, k.Quit}
}
func (k keymap) FullHelp() [][]key.Binding { return [][]key.Binding{k.ShortHelp()} }

var keys = keymap{
	NextPane:   key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next pane")),
	PrevPane:   key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("S-tab", "prev pane")),
	FocusLeft:  key.NewBinding(key.WithKeys("ctrl+h", "left"), key.WithHelp("←/C-h", "focus left")),
	FocusRight: key.NewBinding(key.WithKeys("ctrl+l", "right"), key.WithHelp("→/C-l", "focus right")),
	AuthMenu:   key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "auth menu")),
	Shortcuts:  key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "shortcuts")),
	Quit:       key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

type paneKind int

const (
	paneTable paneKind = iota
	paneViewport
)

type pane struct {
	title                               string
	kind                                paneKind
	table                               table.Model
	vp                                  viewport.Model
	focused                             bool
	posX, posY, widthCells, heightCells int
}

type styles struct {
	focus, blur, title, status lipgloss.Style
}

func newStyles() styles {
	return styles{
		focus:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("12")),
		blur:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")),
		title:  lipgloss.NewStyle().Bold(true),
		status: lipgloss.NewStyle().Faint(true),
	}
}

type model struct {
	panes                 []pane
	focusedPaneIndex      int
	width, height         int
	styles                styles
	help                  help.Model
	statusText            string
	cfg                   config.Config
	showAuth              bool
	authMenu              list.Model
	signedIn              bool
	showShortcuts         bool
	shortcuts             list.Model
	horizontalMarginCells int
	moduleList            list.Model
	activeModuleIndex     int
}

type authMenuItem struct{ title, desc, action string }

func (i authMenuItem) Title() string       { return i.title }
func (i authMenuItem) Description() string { return i.desc }
func (i authMenuItem) FilterValue() string { return i.title }

// Items for shortcuts list
type shortcutItem struct{ title, desc string }

func (i shortcutItem) Title() string       { return i.title }
func (i shortcutItem) Description() string { return i.desc }
func (i shortcutItem) FilterValue() string { return i.title + " " + i.desc }

func formatKeyLabel(keys []string) string {
	if len(keys) == 0 {
		return ""
	}
	s := keys[0]
	// Normalize common patterns
	if s == "shift+tab" {
		return "S-tab"
	}
	if strings.HasPrefix(s, "ctrl+") && len(s) == len("ctrl+")+1 {
		return "C-" + s[len("ctrl+"):]
	}
	if strings.HasPrefix(s, "shift+") && len(s) == len("shift+")+1 {
		return strings.ToUpper(s[len("shift+"):])
	}
	if len(s) == 1 {
		return s
	}
	// arrows
	switch s {
	case "left":
		return "←"
	case "right":
		return "→"
	case "up":
		return "↑"
	case "down":
		return "↓"
	}
	return s
}

func (m *model) buildShortcutsItems() []list.Item {
	items := []list.Item{}
	// Global
	items = append(items,
		shortcutItem{title: formatKeyLabel(keys.NextPane.Keys()), desc: "Next pane [Global]"},
		shortcutItem{title: formatKeyLabel(keys.PrevPane.Keys()), desc: "Prev pane [Global]"},
		shortcutItem{title: formatKeyLabel(keys.FocusLeft.Keys()), desc: "Focus left [Global]"},
		shortcutItem{title: formatKeyLabel(keys.FocusRight.Keys()), desc: "Focus right [Global]"},
		shortcutItem{title: formatKeyLabel(keys.AuthMenu.Keys()), desc: "Auth menu [Global]"},
		shortcutItem{title: formatKeyLabel(keys.Shortcuts.Keys()), desc: "Show shortcuts [Global]"},
		shortcutItem{title: formatKeyLabel(keys.Quit.Keys()), desc: "Quit [Global]"},
	)
	// Resources (table) common keys
	items = append(items,
		shortcutItem{title: "↑/↓", desc: "Move selection [Resources]"},
		shortcutItem{title: "PgUp/PgDn", desc: "Page [Resources]"},
		shortcutItem{title: "Home/End", desc: "Top/Bottom [Resources]"},
		shortcutItem{title: "Enter", desc: "Open details [Resources]"},
	)
	// Incidents (viewport) scrolling
	items = append(items,
		shortcutItem{title: "↑/↓", desc: "Scroll [Incidents]"},
		shortcutItem{title: "PgUp/PgDn", desc: "Page [Incidents]"},
	)
	res := make([]list.Item, 0, len(items))
	for _, it := range items {
		res = append(res, it)
	}
	return res
}

func initialModel(cfg config.Config) model {
	// Left pane: table of resources
	columns := []table.Column{
		{Title: "Name", Width: 24},
		{Title: "Type", Width: 24},
		{Title: "Location", Width: 16},
	}
	rows := []table.Row{
		{"vm-prod-01", "Microsoft.Compute/virtualMachines", "westeurope"},
		{"stlogs01", "Microsoft.Storage/storageAccounts", "eastus"},
		{"pip-web-01", "Microsoft.Network/publicIPAddresses", "westeurope"},
		{"sql-core", "Microsoft.Sql/servers", "swedencentral"},
	}
	tbl := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	// Right pane: scrollable viewport for details
	right := viewport.New(40, 10)
	right.SetContent("Incidents / Details…\n(Tab to switch focus)")
	items := []list.Item{
		authMenuItem{title: "Sign in (Azure CLI)", desc: "Run az login and refresh status", action: "cli"},
		authMenuItem{title: "Sign in (Device Code)", desc: "Interactive device code flow", action: "devicecode"},
	}
	menu := list.New(items, list.NewDefaultDelegate(), 0, 0)
	menu.Title = "Authentication"

	// Shortcuts overlay list (empty for now; populated on open)
	shorts := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	shorts.Title = "Shortcuts (type to filter, enter/esc/? to close)"

	// Module sidebar list
	mods := list.New([]list.Item{
		shortcutItem{title: "resources", desc: "Azure resources"},
		shortcutItem{title: "incidents", desc: "Security incidents"},
	}, list.NewDefaultDelegate(), 0, 0)
	mods.Title = "Modules"

	return model{
		panes: []pane{
			{title: "Sidebar", kind: paneTable, table: tbl, focused: true},
			{title: "Main", kind: paneViewport, vp: right, focused: false},
		},
		styles:                newStyles(),
		help:                  help.New(),
		statusText:            "Authenticating…",
		cfg:                   cfg,
		showAuth:              false,
		authMenu:              menu,
		signedIn:              false,
		showShortcuts:         false,
		shortcuts:             shorts,
		horizontalMarginCells: 2,
		moduleList:            mods,
		activeModuleIndex:     0,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, authenticateCmd(m.cfg))
}

type authResultMsg struct {
	info iauth.Info
	err  error
}

func authenticateCmd(cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		info, err := iauth.Authenticate(context.Background(), cfg)
		return authResultMsg{info: info, err: err}
	}
}

type azLoginResultMsg struct{ err error }

func azLoginCmd() tea.Cmd {
	return func() tea.Msg {
		// Attempt normal az login flow
		err := exec.Command("az", "login").Run()
		return azLoginResultMsg{err: err}
	}
}

// layout(), Update(), and View() moved to layout.go, input.go, and view.go

func main() {
	// CLI flags
	flagHelpShort := flag.Bool("h", false, "Show help")
	flagHelpLong := flag.Bool("help", false, "Show help")
	flagAuth := flag.String("auth", "", "Authentication method: cli (default) or devicecode")
	flagTenant := flag.String("tenant", "", "Azure tenant ID (GUID)")
	flagClientID := flag.String("client-id", "", "Client (application) ID for device code auth")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "lazyms - Azure security TUI\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: lazyms [flags]\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nKeys: Tab/Shift+Tab switch panes, Ctrl+h/l focus, q to quit.\n")
	}
	flag.Parse()

	if *flagHelpShort || *flagHelpLong {
		flag.Usage()
		return
	}

	// Load config and apply flag overrides
	cfg, _ := config.Load()
	if *flagAuth != "" {
		cfg.PreferredAuth = *flagAuth
	}
	if *flagTenant != "" {
		cfg.TenantID = *flagTenant
	}
	if *flagClientID != "" {
		cfg.ClientID = *flagClientID
	}

	p := tea.NewProgram(initialModel(cfg), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	_ = version
	_ = commit
	_ = date
}
