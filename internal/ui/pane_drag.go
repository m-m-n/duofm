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

// snapshotDisplayedNames returns a new ordered sequence holding the name of
// every displayed (post-filter) entry, in display order, including the
// parent entry when it is shown. Later changes to the pane's entries do not
// change the returned sequence, and changes to the returned sequence do not
// change the pane.
func (p *Pane) snapshotDisplayedNames() []string {
	names := make([]string, len(p.entries))
	for i, entry := range p.entries {
		names[i] = entry.Name
	}
	return names
}

// displayedNamesMatch reports whether names has the same length as the
// pane's current displayed entries and every position holds the same name.
// It modifies neither the pane nor names.
func (p *Pane) displayedNamesMatch(names []string) bool {
	if len(names) != len(p.entries) {
		return false
	}
	for i, entry := range p.entries {
		if entry.Name != names[i] {
			return false
		}
	}
	return true
}

// existingMarks returns a new map holding exactly the baseline entries whose
// names appear in the pane's current full (pre-filter) entry list -- the
// same list RefreshDirectoryPreserveCursor uses to drop marks of vanished
// files. Names hidden by an active filter but still present in the full
// list are kept. It modifies neither the pane nor baseline.
func (p *Pane) existingMarks(baseline map[string]bool) map[string]bool {
	existing := make(map[string]bool, len(p.allEntries))
	for _, entry := range p.allEntries {
		existing[entry.Name] = true
	}

	result := make(map[string]bool, len(baseline))
	for name, marked := range baseline {
		if existing[name] {
			result[name] = marked
		}
	}
	return result
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
