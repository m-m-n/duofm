# Threat Model: minibuffer-grapheme-bounds-linear

## Verdict
threats-identified

## Rationale
Inspected SPEC.md (FR1, FR2, NFR1–NFR3) and `internal/ui/minibuffer.go`, the only production file this feature changes; tier is reduced. The change is confined to how the minibuffer locates the grapheme cluster around the cursor. That computation runs on text the terminal user types or pastes into the minibuffer and on text set programmatically by history search, and pasted text can originate outside the project. task0001 declares the domains `ui` and `input-handling`; `input-handling` sets TB-1's depth to deep. The one realistic threat is algorithmic-complexity denial of service driven by crafted text — the defect this feature fixes. No other STRIDE category applies: the change adds no persistence, no external I/O and no privilege change, and it does not alter what text is rendered or how.

## Trust Boundaries

### TB-1: Text entering the minibuffer
Crossing: text from the terminal user (typed or pasted key events) and text set programmatically from history search enters the minibuffer input, and is segmented by the grapheme cluster locator on every redraw while the cursor is not at the end and on every Left/Ctrl+B and Right/Ctrl+F key press. Pasted text may come from a source outside the project.
Boundary files: `internal/ui/minibuffer.go`
Depth: deep (input-handling)

| STRIDE category | Threat | Mitigation ID | Mitigation | Implemented by | Verified by |
|---|---|---|---|---|---|
| Denial of service | Text shaped as a sentence terminator followed by a long run of spaces and a letter makes each locator call quadratic in the input length, stalling every such redraw and cursor-movement key press (about 0.4 s at 8000 runes, about 60 s at 100000 runes) — relates to FR1, NFR1 | TM-1 | The locator computes grapheme cluster boundaries only, in a single forward pass from the start of the input that carries segmentation state and stops at the cluster containing the requested position, so one call is linear in the input length; time-bounded regression tests fail on a return to quadratic behavior | task0001 AC-1, AC-2, AC-3, AC-4 | VERIFICATION.md Performance / Security Verification item TM-1 (TS1, TS2, TS3) |
