package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	// Sidebar view
	sidebarInner := m.moduleList.View()
	// Main view
	mainInner := func() string {
		if m.panes[1].kind == paneViewport {
			return m.panes[1].vp.View()
		}
		return m.panes[1].table.View()
	}()

	leftStyle := m.styles.blur
	rightStyle := m.styles.blur
	if m.panes[0].focused {
		leftStyle = m.styles.focus
		rightStyle = m.styles.blur
	} else {
		leftStyle = m.styles.blur
		rightStyle = m.styles.focus
	}
	leftStyle = leftStyle.Width(m.panes[0].widthCells - 2).Height(m.panes[0].heightCells - 2)
	rightStyle = rightStyle.Width(m.panes[1].widthCells - 2).Height(m.panes[1].heightCells - 2)
	leftBox := leftStyle.Render(m.styles.title.Render(" Sidebar ") + "\n" + sidebarInner)
	rightBox := rightStyle.Render(m.styles.title.Render(" Main ") + "\n" + mainInner)
	row := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, " ", rightBox)

	// Clamp content row to inner width
	innerWidthCells := m.width - m.horizontalMarginCells*2
	if innerWidthCells < 1 {
		innerWidthCells = 1
	}
	row = lipgloss.NewStyle().Width(innerWidthCells).Render(row)

	// Footer: help left, version+signin right, single line
	helpLeft := m.help.View(keys)
	rightParts := []string{"v" + version}
	if m.signedIn {
		rightParts = append(rightParts, "signed in")
	} else {
		rightParts = append(rightParts, "signed out")
	}
	rightText := strings.Join(rightParts, " · ")
	leftWidth := (innerWidthCells * 2) / 3
	if leftWidth < 1 {
		leftWidth = 1
	}
	rightWidth := innerWidthCells - leftWidth
	if rightWidth < 0 {
		rightWidth = 0
	}
	footer := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(leftWidth).Render(helpLeft),
		lipgloss.NewStyle().Width(rightWidth).Align(lipgloss.Right).Render(rightText),
	)

	container := lipgloss.NewStyle().Width(m.width).PaddingLeft(m.horizontalMarginCells).PaddingRight(m.horizontalMarginCells)

	content := row + "\n" + m.styles.status.Render(m.statusText) + "\n" + footer
	return container.Render(content)
}
