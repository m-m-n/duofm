package ui

import (
	"testing"
	"time"
)

// fakeClock is an injectable time source for double-click detector tests.
type fakeClock struct {
	t time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{t: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)}
}

func (c *fakeClock) now() time.Time { return c.t }

func (c *fakeClock) advance(d time.Duration) { c.t = c.t.Add(d) }

func sampleKey() clickKey {
	return clickKey{pane: LeftPane, dir: "/tmp/dir", name: "file.txt"}
}

// --- AC-5 ---

func TestDoubleClickDetector_SameKeyWithinWindow(t *testing.T) {
	clock := newFakeClock()
	d := newDoubleClickDetector(clock.now)
	key := sampleKey()

	if got := d.press(key); got != false {
		t.Fatalf("first press = %v, want false", got)
	}
	clock.advance(200 * time.Millisecond)
	if got := d.press(key); got != true {
		t.Fatalf("second press 200ms later = %v, want true", got)
	}
}

func TestDoubleClickDetector_WindowBoundaryInclusiveAt500(t *testing.T) {
	clock := newFakeClock()
	d := newDoubleClickDetector(clock.now)
	key := sampleKey()

	d.press(key)
	clock.advance(500 * time.Millisecond)
	if got := d.press(key); got != true {
		t.Errorf("500ms apart = %v, want true (inclusive boundary)", got)
	}
}

func TestDoubleClickDetector_PastWindowIsFalse(t *testing.T) {
	for _, elapsed := range []time.Duration{501 * time.Millisecond, 600 * time.Millisecond} {
		t.Run(elapsed.String(), func(t *testing.T) {
			clock := newFakeClock()
			d := newDoubleClickDetector(clock.now)
			key := sampleKey()

			d.press(key)
			clock.advance(elapsed)
			if got := d.press(key); got != false {
				t.Errorf("%v apart = %v, want false", elapsed, got)
			}
		})
	}
}

func TestDoubleClickDetector_KeyMustMatchAllFields(t *testing.T) {
	base := sampleKey()
	variants := []clickKey{
		{pane: base.pane, dir: base.dir, name: "other.txt"},   // differs in name
		{pane: RightPane, dir: base.dir, name: base.name},     // differs in pane
		{pane: base.pane, dir: "/tmp/other", name: base.name}, // differs in dir
	}

	for _, v := range variants {
		t.Run(v.name+"|"+v.dir, func(t *testing.T) {
			clock := newFakeClock()
			d := newDoubleClickDetector(clock.now)

			d.press(base)
			clock.advance(100 * time.Millisecond)
			if got := d.press(v); got != false {
				t.Errorf("mismatched key 100ms later = %v, want false", got)
			}
		})
	}
}

func TestDoubleClickDetector_SameKeySequenceAlternatesTrueFalse(t *testing.T) {
	clock := newFakeClock()
	d := newDoubleClickDetector(clock.now)
	key := sampleKey()

	want := []bool{false, true, false, true}
	times := []time.Duration{0, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond}

	for i, dt := range times {
		clock.advance(dt)
		got := d.press(key)
		if got != want[i] {
			t.Errorf("press #%d = %v, want %v", i, got, want[i])
		}
	}
}

func TestDoubleClickDetector_ABASequenceNeverTriggers(t *testing.T) {
	clock := newFakeClock()
	d := newDoubleClickDetector(clock.now)
	a := clickKey{pane: LeftPane, dir: "/d", name: "A"}
	b := clickKey{pane: LeftPane, dir: "/d", name: "B"}

	keys := []clickKey{a, b, a}
	for i, k := range keys {
		if i > 0 {
			clock.advance(100 * time.Millisecond)
		}
		if got := d.press(k); got != false {
			t.Errorf("press #%d (A,B,A sequence) = %v, want false", i, got)
		}
	}
}

func TestDoubleClickDetector_ResetRestartsSequence(t *testing.T) {
	clock := newFakeClock()
	d := newDoubleClickDetector(clock.now)
	key := sampleKey()

	d.press(key)
	d.reset()
	clock.advance(100 * time.Millisecond)
	if got := d.press(key); got != false {
		t.Errorf("press after reset = %v, want false", got)
	}
}

func TestDoubleClickDetector_BackwardsClockNeverTriggers(t *testing.T) {
	clock := newFakeClock()
	d := newDoubleClickDetector(clock.now)
	key := sampleKey()

	clock.advance(1 * time.Second)
	d.press(key)
	clock.advance(-500 * time.Millisecond) // clock moves backwards
	if got := d.press(key); got != false {
		t.Errorf("press with backwards clock = %v, want false", got)
	}
}

func TestDoubleClickDetector_NoTimeSourceUsesSystemClock(t *testing.T) {
	d := newDoubleClickDetector(nil)
	key := sampleKey()

	if got := d.press(key); got != false {
		t.Fatalf("first press = %v, want false", got)
	}
	if got := d.press(key); got != true {
		t.Errorf("second immediate press = %v, want true", got)
	}
}
