package main

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()
	case tea.KeyMsg:
		// Global auth menu toggle
		if key.Matches(msg, keys.AuthMenu) {
			m.showAuth = !m.showAuth
			if m.showAuth {
				m.showShortcuts = false
			}
			return m, nil
		}
		// Global shortcuts menu toggle
		if key.Matches(msg, keys.Shortcuts) || msg.String() == "?" {
			m.showShortcuts = !m.showShortcuts
			if m.showShortcuts {
				m.showAuth = false
				m.shortcuts.SetItems(m.buildShortcutsItems())
			}
			return m, nil
		}
		// Numeric shortcuts like Lazygit: 0 focuses main; 1..9 select module
		if s := msg.String(); len(s) == 1 && s[0] >= '0' && s[0] <= '9' {
			if s == "0" {
				m.setFocus(1)
				return m, nil
			}
			idx := int(s[0] - '1')
			total := len(m.moduleList.Items())
			if idx >= 0 && idx < total {
				// select in sidebar and activate
				m.moduleList.Select(idx)
				m.activeModuleIndex = idx
				if m.activeModuleIndex == 0 { // resources
					m.panes[1].kind = paneViewport
					m.panes[1].vp.SetContent("Resources module (main view placeholder)")
				} else { // incidents or others
					m.panes[1].kind = paneViewport
					m.panes[1].vp.SetContent("Incidents module (main view placeholder)")
				}
				return m, nil
			}
		}
		// If auth menu is open, route input to it
		if m.showAuth {
			var c tea.Cmd
			m.authMenu, c = m.authMenu.Update(msg)
			switch msg.String() {
			case "enter":
				if it, ok := m.authMenu.SelectedItem().(authMenuItem); ok {
					switch it.action {
					case "cli":
						m.statusText = "Starting az login…"
						m.showAuth = false
						return m, azLoginCmd()
					case "devicecode":
						m.statusText = "Starting device code auth…"
						m.showAuth = false
						cfg := m.cfg
						cfg.PreferredAuth = "devicecode"
						return m, authenticateCmd(cfg)
					}
				}
			case "esc":
				m.showAuth = false
			}
			return m, c
		}
		// If shortcuts overlay is open, route input
		if m.showShortcuts {
			var c tea.Cmd
			m.shortcuts, c = m.shortcuts.Update(msg)
			switch msg.String() {
			case "enter", "esc", "?":
				m.showShortcuts = false
			}
			return m, c
		}
		// Sidebar navigation and module selection
		prevIndex := m.moduleList.Index()
		m.moduleList, _ = m.moduleList.Update(msg)
		if msg.String() == "enter" || msg.String() == "tab" {
			m.activeModuleIndex = m.moduleList.Index()
			// Swap main pane content depending on module selection (placeholder behavior)
			if m.activeModuleIndex == 0 { // resources
				m.panes[1].kind = paneViewport
				m.panes[1].vp.SetContent("Resources module (main view placeholder)")
			} else { // incidents
				m.panes[1].kind = paneViewport
				m.panes[1].vp.SetContent("Incidents module (main view placeholder)")
			}
			return m, nil
		}
		// If sidebar index changed via arrows, update focus but don’t switch yet
		if m.moduleList.Index() != prevIndex {
			return m, nil
		}
		// Other global focus keys
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.NextPane):
			m.setFocus((m.focusedPaneIndex + 1) % len(m.panes))
		case key.Matches(msg, keys.PrevPane):
			m.setFocus((m.focusedPaneIndex - 1 + len(m.panes)) % len(m.panes))
		case key.Matches(msg, keys.FocusLeft):
			m.setFocus(0)
		case key.Matches(msg, keys.FocusRight):
			if len(m.panes) > 1 {
				m.setFocus(1)
			}
		}
		fp := &m.panes[m.focusedPaneIndex]
		switch fp.kind {
		case paneTable:
			fp.table, cmd = fp.table.Update(msg)
		case paneViewport:
			fp.vp, cmd = fp.vp.Update(msg)
		}
	case authResultMsg:
		if msg.err != nil {
			m.statusText = "Auth error: " + msg.err.Error()
			m.signedIn = false
		} else {
			who := msg.info.UserPrincipalName
			if who == "" {
				who = "?"
			}
			ten := msg.info.TenantID
			if ten == "" {
				ten = "?"
			}
			m.statusText = "Logged in as " + who + " (tenant " + ten + ")"
			m.signedIn = true
		}
	case azLoginResultMsg:
		if msg.err != nil {
			m.statusText = "az login failed: " + msg.err.Error()
			m.signedIn = false
		} else {
			m.statusText = "az login complete; refreshing…"
			return m, authenticateCmd(m.cfg)
		}
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			// Focus pane on mouse press inside its bounds
			for i := range m.panes {
				if pointInPaneBounds(msg.X, msg.Y, m.panes[i]) {
					m.setFocus(i)
					break
				}
			}
		}
		// Always forward mouse events to focused component (wheel, motion, etc.)
		fp := &m.panes[m.focusedPaneIndex]
		switch fp.kind {
		case paneTable:
			fp.table, cmd = fp.table.Update(msg)
		case paneViewport:
			fp.vp, cmd = fp.vp.Update(msg)
		}
	}
	return m, cmd
}
