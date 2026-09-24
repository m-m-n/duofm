package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// dragSession holds the state of an in-progress mouse drag gesture.
type dragSession struct {
	armed    bool
	active   bool
	pane     PanePosition
	anchor   int
	baseline map[string]bool
}

// isMouseModal reports whether the Model is in a state that suppresses all
// mouse handling. Every check is nil-safe.
func (m *Model) isMouseModal() bool {
	if m.dialog != nil {
		return true
	}
	if m.sortDialog != nil && m.sortDialog.IsActive() {
		return true
	}
	if m.searchState.IsActive {
		return true
	}
	if m.shellCommandMode {
		return true
	}
	if m.bgOutputFocused {
		return true
	}
	return false
}

// mouseDetector returns the Model's double-click detector, lazily creating
// one with the system clock when the Model was built without the
// constructor (as several existing tests do).
func (m *Model) mouseDetector() *doubleClickDetector {
	if m.detector == nil {
		m.detector = newDoubleClickDetector(nil)
	}
	return m.detector
}

// handleMouse is the mouse entry point: it gates on readiness and modal
// state, filters out wheel events and non-left buttons, and routes the
// remaining press/motion/release events to the gesture state machine.
// Modifier keys never change handling.
func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Either pane is absent (before the first WindowSizeMsg) -> ignore.
	if m.leftPane == nil || m.rightPane == nil {
		return m, nil
	}

	// Modal state -> disarm any session and ignore the event.
	if m.isMouseModal() {
		m.mouseDrag = dragSession{}
		return m, nil
	}

	// Wheel events, and right- or middle-button events of any action, are
	// ignored. An armed session is left as it is.
	event := tea.MouseEvent(msg)
	if event.IsWheel() || msg.Button == tea.MouseButtonRight || msg.Button == tea.MouseButtonMiddle {
		return m, nil
	}

	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button != tea.MouseButtonLeft {
			return m, nil
		}
		cmd := m.handleMousePress(msg.X, msg.Y)
		return m, cmd

	case tea.MouseActionMotion:
		if msg.Button != tea.MouseButtonLeft || !m.mouseDrag.armed {
			return m, nil
		}
		m.updateDrag(msg.Y)
		return m, nil

	case tea.MouseActionRelease:
		if !m.mouseDrag.armed {
			return m, nil
		}
		m.updateDrag(msg.Y)
		m.mouseDrag = dragSession{}
		return m, nil
	}

	return m, nil
}

// handleMousePress evaluates the hit test before any state change, discards
// any existing drag session, then acts on the hit kind.
func (m *Model) handleMousePress(x, y int) tea.Cmd {
	hit := m.mouseHitAt(x, y)
	m.mouseDrag = dragSession{}

	switch hit.kind {
	case hitNone:
		m.mouseDetector().reset()
		return nil

	case hitNonEntry:
		m.mouseDetector().reset()
		if hit.pane != m.activePane {
			m.switchToPane(hit.pane)
		}
		return nil

	case hitEntry:
		return m.handleEntryPress(hit.pane, hit.index)
	}

	return nil
}

// handleEntryPress applies the click effect at once (activate the pane if
// needed, move its cursor), then consults the double-click detector to
// either run the existing Enter action or arm a drag session.
func (m *Model) handleEntryPress(pane PanePosition, index int) tea.Cmd {
	if pane != m.activePane {
		m.switchToPane(pane)
	}

	p := m.paneAt(pane)
	p.SetCursor(index)

	// If activation shrank the pane's visible lines (the bg split appears
	// on the bg command's pane) and the cursor fell outside them, apply the
	// existing ensure-visible adjustment.
	if index < p.scrollOffset || index >= p.scrollOffset+p.getVisibleLines() {
		p.EnsureCursorVisible()
	}

	entry := p.entries[index]
	key := clickKey{pane: pane, dir: p.Path(), name: entry.Name}

	if m.mouseDetector().press(key) {
		newModel, cmd := m.handleEnter()
		*m = newModel.(Model)
		return cmd
	}

	m.mouseDrag = dragSession{
		armed:    true,
		pane:     pane,
		anchor:   index,
		baseline: p.snapshotMarks(),
	}
	return nil
}

// updateDrag applies a motion or release row to the armed drag session. x is
// never used, and the other pane is never touched.
func (m *Model) updateDrag(y int) {
	target, ok := m.mouseDragIndexAt(m.mouseDrag.pane, y)
	if !ok {
		return
	}

	// Still a click: the session hasn't activated and the target is back on
	// the anchor.
	if !m.mouseDrag.active && target == m.mouseDrag.anchor {
		return
	}

	if !m.mouseDrag.active {
		m.mouseDrag.active = true
		m.mouseDetector().reset()
	}

	p := m.paneAt(m.mouseDrag.pane)
	p.applyDragMarks(m.mouseDrag.anchor, target, m.mouseDrag.baseline)
}
