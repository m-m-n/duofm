package ui

// snapshotMarks returns an independent copy of the pane's current mark set,
// in the same name->marked representation the pane uses internally. Later
// changes to either the pane's marks or the returned copy do not affect the
// other.
func (p *Pane) snapshotMarks() map[string]bool {
	snapshot := make(map[string]bool, len(p.markedFiles))
	for name, marked := range p.markedFiles {
		snapshot[name] = marked
	}
	return snapshot
}

// applyDragMarks computes the mark set for a drag range on top of baseline.
//
// baseline must have come from snapshotMarks(). If anchor or target is
// outside [0, entry count), it returns false and changes nothing. Otherwise
// the pane's mark set becomes baseline plus the names of every displayed
// entry between anchor and target inclusive (excluding the parent entry
// ".."), the cursor moves to target, the scroll offset is left untouched,
// baseline itself is never mutated, and it returns true.
func (p *Pane) applyDragMarks(anchor, target int, baseline map[string]bool) bool {
	n := len(p.entries)
	if anchor < 0 || anchor >= n || target < 0 || target >= n {
		return false
	}

	marks := make(map[string]bool, len(baseline))
	for name, marked := range baseline {
		marks[name] = marked
	}

	lo, hi := anchor, target
	if lo > hi {
		lo, hi = hi, lo
	}
	for i := lo; i <= hi; i++ {
		entry := p.entries[i]
		if entry.IsParentDir() {
			continue
		}
		marks[entry.Name] = true
	}

	p.markedFiles = marks
	p.cursor = target
	return true
}
