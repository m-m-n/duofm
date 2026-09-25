package ui

// Layout constants shared by the hit test and the drag-row mapping. These
// mirror the renderer's layout (title row, pane header rows, first entry
// row) without refactoring the rendering files themselves. The pane header
// row count itself is not repeated here — it comes from paneHeaderRows
// (pane.go), the single shared definition also used by the renderer's
// visible-line and bg-split height computations.
const (
	// mouseTitleRow is the screen row occupied by the title bar.
	mouseTitleRow = 0
	// mouseFirstEntryRow is the first screen row an entry can occupy.
	mouseFirstEntryRow = mouseTitleRow + 1 + paneHeaderRows
)

// hitKind classifies what a screen coordinate hits.
type hitKind int

const (
	// hitNone is outside both panes entirely (title bar, status bar, the
	// leftover column of an odd width, or out-of-screen coordinates).
	hitNone hitKind = iota
	// hitNonEntry is inside a pane but not on a displayed entry row.
	hitNonEntry
	// hitEntry is on a displayed entry row.
	hitEntry
)

// mouseHit is the result of a hit test. pane is meaningful unless kind is
// hitNone; index is meaningful only when kind is hitEntry.
type mouseHit struct {
	kind  hitKind
	pane  PanePosition
	index int
}

// paneHitLayout is a read-only layout snapshot of one pane, used for hit
// testing and drag-row mapping.
type paneHitLayout struct {
	scrollOffset int
	visibleLines int
	entryCount   int
}

// hitTest classifies a screen coordinate against the current layout,
// following rules H1-H5. It reads nothing but its arguments, so it can be
// tested without a terminal.
func hitTest(x, y, width, height int, left, right paneHitLayout) mouseHit {
	paneWidth := width / 2

	// H1: title bar, status bar (or beyond), and out-of-pane columns
	// (including the leftover column of an odd width) give hitNone.
	if y <= mouseTitleRow || y >= height-1 || x < 0 || x >= 2*paneWidth {
		return mouseHit{kind: hitNone}
	}

	// H2: pick the pane from x.
	pane := LeftPane
	layout := left
	if x >= paneWidth {
		pane = RightPane
		layout = right
	}

	// H3: pane header rows are non-entry.
	if y < mouseFirstEntryRow {
		return mouseHit{kind: hitNonEntry, pane: pane}
	}

	// H4/H5: entry rows vs. everything else inside the pane (blank rows,
	// "(No matches)", the bg split separator/output rows, the trailing
	// row).
	r := y - mouseFirstEntryRow
	if r < layout.visibleLines && layout.scrollOffset+r < layout.entryCount {
		return mouseHit{kind: hitEntry, pane: pane, index: layout.scrollOffset + r}
	}
	return mouseHit{kind: hitNonEntry, pane: pane}
}

// clampedEntryIndex maps a screen row to a displayed entry index, clamped
// into the pane's currently displayed range. It has no x input and accepts
// any integer y. ok is false iff the pane displays no entry.
func clampedEntryIndex(y int, layout paneHitLayout) (index int, ok bool) {
	if layout.visibleLines <= 0 || layout.entryCount <= layout.scrollOffset {
		return 0, false
	}

	maxIndex := layout.scrollOffset + layout.visibleLines
	if layout.entryCount < maxIndex {
		maxIndex = layout.entryCount
	}
	maxIndex--

	index = layout.scrollOffset + (y - mouseFirstEntryRow)
	if index < layout.scrollOffset {
		index = layout.scrollOffset
	}
	if index > maxIndex {
		index = maxIndex
	}
	return index, true
}

// hitLayout is Pane's read-only adapter to paneHitLayout. It never
// re-derives heights on its own: visible lines come from getVisibleLines(),
// which already reflects bgSplitHeights() when background output is
// active. While a directory is loading, the currently displayed entries are
// used, the same as keyboard cursor movement.
func (p *Pane) hitLayout() paneHitLayout {
	return paneHitLayout{
		scrollOffset: p.scrollOffset,
		visibleLines: p.getVisibleLines(),
		entryCount:   len(p.entries),
	}
}

// paneAt returns the pane at the given position, or nil if it has not been
// initialized yet (before the first WindowSizeMsg).
func (m *Model) paneAt(pos PanePosition) *Pane {
	if pos == LeftPane {
		return m.leftPane
	}
	return m.rightPane
}

// mouseHitAt is Model's read-only adapter combining the Model's current
// width/height with both panes' layouts. If either pane is absent, the kind
// is hitNone.
func (m *Model) mouseHitAt(x, y int) mouseHit {
	if m.leftPane == nil || m.rightPane == nil {
		return mouseHit{kind: hitNone}
	}
	return hitTest(x, y, m.width, m.height, m.leftPane.hitLayout(), m.rightPane.hitLayout())
}

// mouseDragIndexAt is Model's read-only adapter for drag-row mapping,
// applying clampedEntryIndex to the named pane's current layout. Absent
// pane -> ok is false.
func (m *Model) mouseDragIndexAt(pane PanePosition, y int) (int, bool) {
	p := m.paneAt(pane)
	if p == nil {
		return 0, false
	}
	return clampedEntryIndex(y, p.hitLayout())
}
