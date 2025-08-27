package main

func (m *model) setFocus(i int) {
	for j := range m.panes {
		m.panes[j].focused = false
		if m.panes[j].kind == paneTable {
			m.panes[j].table.Blur()
		}
	}
	m.panes[i].focused = true
	m.focusedPaneIndex = i
	if m.panes[i].kind == paneTable {
		m.panes[i].table.Focus()
	}
}

func pointInPaneBounds(mx, my int, p pane) bool {
	return mx >= p.posX && mx < p.posX+p.widthCells && my >= p.posY && my < p.posY+p.heightCells
}
