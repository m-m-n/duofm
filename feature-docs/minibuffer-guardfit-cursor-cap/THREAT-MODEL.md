# Threat Model: minibuffer-guardfit-cursor-cap

## Verdict
threats-identified

## Rationale
Inspected: SPEC.md (FR1-FR4, NFR1-NFR3; REQUIREMENTS.md is not produced at
the reduced tier, and the design step was skipped), the minibuffer's View
and its width-fitting helpers in `internal/ui/minibuffer.go`, and the
channels that fill the minibuffer input (keyboard input, command history,
and shell-command tab completion, which inserts file-system entry names from
the active pane's directory). Tier: reduced. The single task declares the
domains `ui` and `input-handling`; `input-handling` sets the analysis depth
of TB-1 to deep.

The feature changes only how View picks the displayed input range when the
width-fitting guard gives up. The text it renders can contain file-system
entry names that another party chose (tab completion), so the rendering
path is a trust boundary. Two denial-of-service threats apply to the
changed fallback: per-redraw cost and single-line layout integrity. No
spoofing, tampering, repudiation, information disclosure or elevation of
privilege applies to the changed code: it only narrows the displayed range
of text already present in the input, reads nothing new, and hands the
result to the same terminal as before. The handling of control characters
in input text is unchanged by this feature (NFR1) and is not part of this
analysis.

## Trust Boundaries

### TB-1: Minibuffer input text to the rendered minibuffer line
Crossing: text typed by the user or inserted by shell-command tab
completion (file-system entry names chosen by whoever created those
entries) flows into View, which measures and renders it to the terminal.
Boundary files: `internal/ui/minibuffer.go`
Depth: deep (input-handling)

| STRIDE category | Threat | Mitigation ID | Mitigation | Implemented by | Verified by |
|---|---|---|---|---|---|
| Denial of service | Input whose grapheme width estimates undershoot the actual rendered width (e.g. a crafted entry name inserted by tab completion) drives View into the re-measurement cap; a cursor-only fallback that re-measures repeatedly, or in proportion to the input length, would make every redraw slower as the input grows (relates to NFR2). | TM-1 | The cursor-only fallback adds at most one width measurement per View call on top of guardFit's bounded measurements, so the number of measurements per View call does not depend on input length; the existing linear-time tests keep passing within their limit. | task0001 AC-5 | VERIFICATION.md Performance / Security Verification: TM-1 |
| Denial of service | The same input makes the displayed minibuffer line wider than its content width, so the line wraps and breaks the single-line TUI layout (relates to FR2, FR3). | TM-2 | The cursor-only line is measured with the width basis before it is displayed and is displayed only when it fits the content width; otherwise neither input nor cursor is rendered, so the single-line, width <= m.width-2 postcondition holds on every fallback path. | task0001 AC-4 | VERIFICATION.md Performance / Security Verification: TM-2 |
