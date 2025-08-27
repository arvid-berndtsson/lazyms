package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type keymap struct {
	NextPane, PrevPane, FocusLeft, FocusRight, Quit key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.NextPane, k.PrevPane, k.FocusLeft, k.FocusRight, k.Quit}
}
func (k keymap) FullHelp() [][]key.Binding { return [][]key.Binding{k.ShortHelp()} }

var keys = keymap{
	NextPane:   key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next pane")),
	PrevPane:   key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("S-tab", "prev pane")),
	FocusLeft:  key.NewBinding(key.WithKeys("ctrl+h", "left"), key.WithHelp("←/C-h", "focus left")),
	FocusRight: key.NewBinding(key.WithKeys("ctrl+l", "right"), key.WithHelp("→/C-l", "focus right")),
	Quit:       key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

type paneKind int

const (
	paneTable paneKind = iota
	paneViewport
)

type pane struct {
	title      string
	kind       paneKind
	table      table.Model
	vp         viewport.Model
	focused    bool
	x, y, w, h int
}

type styles struct {
	focus, blur, title lipgloss.Style
}

func newStyles() styles {
	return styles{
		focus: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("12")),
		blur:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")),
		title: lipgloss.NewStyle().Bold(true),
	}
}

type model struct {
	panes         []pane
	focusIdx      int
	width, height int
	styles        styles
	help          help.Model
}

func initialModel() model {
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
	return model{
		panes: []pane{
			{title: "Resources", kind: paneTable, table: tbl, focused: true},
			{title: "Incidents", kind: paneViewport, vp: right, focused: false},
		},
		styles: newStyles(),
		help:   help.New(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.NextPane):
			m.setFocus((m.focusIdx + 1) % len(m.panes))
		case key.Matches(msg, keys.PrevPane):
			m.setFocus((m.focusIdx - 1 + len(m.panes)) % len(m.panes))
		case key.Matches(msg, keys.FocusLeft):
			m.setFocus(0)
		case key.Matches(msg, keys.FocusRight):
			if len(m.panes) > 1 {
				m.setFocus(1)
			}
		}
		fp := &m.panes[m.focusIdx]
		switch fp.kind {
		case paneTable:
			fp.table, cmd = fp.table.Update(msg)
		case paneViewport:
			fp.vp, cmd = fp.vp.Update(msg)
		}
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			// Focus pane on mouse press inside its bounds
			for i := range m.panes {
				if inside(msg.X, msg.Y, m.panes[i]) {
					m.setFocus(i)
					break
				}
			}
		}
		// Always forward mouse events to focused component (wheel, motion, etc.)
		fp := &m.panes[m.focusIdx]
		switch fp.kind {
		case paneTable:
			fp.table, cmd = fp.table.Update(msg)
		case paneViewport:
			fp.vp, cmd = fp.vp.Update(msg)
		}
	}
	return m, cmd
}

func (m *model) layout() {
	gap := 1
	leftW := (m.width - gap) / 2
	rightW := m.width - gap - leftW
	// Reserve space for the help bar at the bottom
	helpHeight := lipgloss.Height(m.help.View(keys))
	if helpHeight < 1 {
		helpHeight = 1
	}
	totalPaneHeight := m.height - helpHeight
	if totalPaneHeight < 4 {
		totalPaneHeight = 4
	}
	// Content height inside borders minus title line
	contentH := totalPaneHeight - 3
	if contentH < 1 {
		contentH = 1
	}
	// Left pane (table)
	m.panes[0].x, m.panes[0].y, m.panes[0].w, m.panes[0].h = 0, 0, leftW, totalPaneHeight
	lw := leftW - 2
	if lw < 1 {
		lw = 1
	}
	m.panes[0].table.SetWidth(lw)
	m.panes[0].table.SetHeight(contentH)
	// Right pane (viewport)
	m.panes[1].x, m.panes[1].y, m.panes[1].w, m.panes[1].h = leftW+gap, 0, rightW, totalPaneHeight
	rw := rightW - 2
	if rw < 1 {
		rw = 1
	}
	m.panes[1].vp.Width, m.panes[1].vp.Height = rw, contentH
}

func (m *model) setFocus(i int) {
	for j := range m.panes {
		m.panes[j].focused = false
		if m.panes[j].kind == paneTable {
			m.panes[j].table.Blur()
		}
	}
	m.panes[i].focused = true
	m.focusIdx = i
	if m.panes[i].kind == paneTable {
		m.panes[i].table.Focus()
	}
}

func inside(mx, my int, p pane) bool {
	return mx >= p.x && mx < p.x+p.w && my >= p.y && my < p.y+p.h
}

func (m model) View() string {
	var leftBox, rightBox string
	// Compose inner views per pane kind
	leftInner := func() string {
		if m.panes[0].kind == paneTable {
			return m.panes[0].table.View()
		}
		return m.panes[0].vp.View()
	}()
	rightInner := func() string {
		if m.panes[1].kind == paneViewport {
			return m.panes[1].vp.View()
		}
		return m.panes[1].table.View()
	}()
	if m.panes[0].focused {
		leftBox = m.styles.focus.Render(m.styles.title.Render(" "+m.panes[0].title+" ") + "\n" + leftInner)
		rightBox = m.styles.blur.Render(m.styles.title.Render(" "+m.panes[1].title+" ") + "\n" + rightInner)
	} else {
		leftBox = m.styles.blur.Render(m.styles.title.Render(" "+m.panes[0].title+" ") + "\n" + leftInner)
		rightBox = m.styles.focus.Render(m.styles.title.Render(" "+m.panes[1].title+" ") + "\n" + rightInner)
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, " ", rightBox)
	return row + "\n" + m.help.View(keys)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	_ = version
	_ = commit
	_ = date
}
