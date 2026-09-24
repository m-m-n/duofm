package ui

import "time"

// doubleClickWindow is the maximum elapsed time between two presses on the
// same entry for the second press to count as a double-click. The
// comparison is inclusive (elapsed == doubleClickWindow qualifies).
const doubleClickWindow = 500 * time.Millisecond

// clickKey identifies the press target for double-click detection. Two keys
// are equal iff all three fields are equal.
type clickKey struct {
	pane PanePosition
	dir  string
	name string
}

// doubleClickDetector tracks press-sequence timing with an injectable time
// source.
type doubleClickDetector struct {
	now      func() time.Time
	hasPrev  bool
	prevKey  clickKey
	prevTime time.Time
}

// newDoubleClickDetector creates a detector. now is the time source used for
// every press; when nil, the system clock is used.
func newDoubleClickDetector(now func() time.Time) *doubleClickDetector {
	if now == nil {
		now = time.Now
	}
	return &doubleClickDetector{now: now}
}

// press returns true iff a previous press is recorded, its key equals key,
// and the elapsed time since that press is within [0, doubleClickWindow].
// On true, the record is cleared. On false, this press becomes the new
// record (whether or not a record existed before).
func (d *doubleClickDetector) press(key clickKey) bool {
	now := d.now()

	if d.hasPrev && d.prevKey == key {
		elapsed := now.Sub(d.prevTime)
		if elapsed >= 0 && elapsed <= doubleClickWindow {
			d.hasPrev = false
			return true
		}
	}

	d.prevKey = key
	d.prevTime = now
	d.hasPrev = true
	return false
}

// reset clears the recorded press, restarting the sequence.
func (d *doubleClickDetector) reset() {
	d.hasPrev = false
}
