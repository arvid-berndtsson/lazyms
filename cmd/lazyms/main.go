package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
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

type pane struct {
	title      string
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
	left := viewport.New(40, 10)
	left.SetContent("Resources…\n(click to focus; wheel to scroll)")
	right := viewport.New(40, 10)
	right.SetContent("Incidents / Details…\n(Tab to switch focus)")
	return model{
		panes: []pane{
			{title: "Resources", vp: left, focused: true},
			{title: "Incidents", vp: right, focused: false},
		},
		styles: newStyles(),
		help:   help.New(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		fp.vp, _ = fp.vp.Update(msg)
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			// Focus pane on mouse press inside its bounds
			for i := range m.panes {
				if inside(msg.X, msg.Y, m.panes[i]) {
					m.setFocus(i)
					break
				}
			}
			// Forward wheel events to focused viewport
			if msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown {
				fp := &m.panes[m.focusIdx]
				fp.vp, _ = fp.vp.Update(msg)
			}
		}
	}
	return m, nil
}

func (m *model) layout() {
	gap := 1
	leftW := (m.width - gap) / 2
	rightW := m.width - gap - leftW
	h := m.height - 2
	m.panes[0].x, m.panes[0].y, m.panes[0].w, m.panes[0].h = 0, 0, leftW, h
	m.panes[0].vp.Width, m.panes[0].vp.Height = leftW-2, h-2
	m.panes[1].x, m.panes[1].y, m.panes[1].w, m.panes[1].h = leftW+gap, 0, rightW, h
	m.panes[1].vp.Width, m.panes[1].vp.Height = rightW-2, h-2
}

func (m *model) setFocus(i int) {
	for j := range m.panes {
		m.panes[j].focused = false
	}
	m.panes[i].focused = true
	m.focusIdx = i
}

func inside(mx, my int, p pane) bool {
	return mx >= p.x && mx < p.x+p.w && my >= p.y && my < p.y+p.h
}

func (m model) View() string {
	var leftBox, rightBox string
	if m.panes[0].focused {
		leftBox = m.styles.focus.Render(m.styles.title.Render(" "+m.panes[0].title+" ") + "\n" + m.panes[0].vp.View())
		rightBox = m.styles.blur.Render(m.styles.title.Render(" "+m.panes[1].title+" ") + "\n" + m.panes[1].vp.View())
	} else {
		leftBox = m.styles.blur.Render(m.styles.title.Render(" "+m.panes[0].title+" ") + "\n" + m.panes[0].vp.View())
		rightBox = m.styles.focus.Render(m.styles.title.Render(" "+m.panes[1].title+" ") + "\n" + m.panes[1].vp.View())
	}
	return leftBox + " " + rightBox + "\n" + m.help.View(keys)
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
