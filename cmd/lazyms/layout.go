package main

import "github.com/charmbracelet/lipgloss"

func (m *model) layout() {
	// Horizontal margins
	m.horizontalMarginCells = 1
	innerWidthCells := m.width - m.horizontalMarginCells*2
	if innerWidthCells < 10 { // enforce minimum inner width for two panes + gap
		innerWidthCells = 10
	}
	columnGapCells := 1
	// Sidebar/main split (default 0.34/0.66)
	sidebarWidthCells := innerWidthCells * 34 / 100
	if sidebarWidthCells < 16 {
		sidebarWidthCells = 16
	}
	mainWidthCells := innerWidthCells - columnGapCells - sidebarWidthCells
	if mainWidthCells < 16 {
		mainWidthCells = 16
	}
	// Footer: help height (may be >1) + status (1)
	helpHeight := lipgloss.Height(m.help.View(keys))
	if helpHeight < 1 {
		helpHeight = 1
	}
	footerLines := helpHeight + 1
	paneAreaHeightCells := m.height - footerLines
	if paneAreaHeightCells < 5 { // title + borders + 1 content line
		paneAreaHeightCells = 5
	}
	contentAreaHeightCells := paneAreaHeightCells - 3
	if contentAreaHeightCells < 1 {
		contentAreaHeightCells = 1
	}
	// Left: sidebar (use moduleList dimensions)
	m.panes[0].posX, m.panes[0].posY = m.horizontalMarginCells, 0
	m.panes[0].widthCells, m.panes[0].heightCells = sidebarWidthCells, paneAreaHeightCells
	m.moduleList.SetSize(sidebarWidthCells-2, contentAreaHeightCells)
	// Right: main (active module)
	m.panes[1].posX, m.panes[1].posY = m.horizontalMarginCells+sidebarWidthCells+columnGapCells, 0
	m.panes[1].widthCells, m.panes[1].heightCells = mainWidthCells, paneAreaHeightCells
	rw := mainWidthCells - 2
	if rw < 1 {
		rw = 1
	}
	m.panes[1].vp.Width, m.panes[1].vp.Height = rw, contentAreaHeightCells
}
