# Implementation Plan: narrow-pane-render-panic

## Overview

Stop the pane views from panicking at pane widths 0–4 and stop the minibuffer view from panicking when its input display width is zero or negative, and add `internal/ui` unit tests that detect a recurrence. Delivered as a single task (task0001).

## Technology Stack

- **Language**: Go (existing module; toolchain and container invocation per workflow.yaml `project.components.duofm`)
- **Framework / Key libraries**: the TUI framework and styling library already used by `internal/ui` — unchanged
- **New dependencies**: none (nothing to check against `project.license` MIT)

## Layer Structure

Only the rendering layer of `internal/ui` is touched. Dependency direction is unchanged: model view composition → pane rendering → minibuffer rendering.

| Layer | File | Role in this feature |
|-------|------|----------------------|
| Model view composition | `internal/ui/model_view.go` | Composes title, two panes and status bar. Not modified (SPEC A1). Exercised by the model-level test only. |
| Pane rendering | `internal/ui/pane_render.go` | Renders one pane as a fixed number of rows. The header separator row is fixed here. |
| Minibuffer rendering | `internal/ui/minibuffer.go` | Renders the single input line shown at the bottom of a pane. The input-window computation is fixed here. |

## Shared Components

No component is shared across tasks: this feature has a single task (task0001). The rendering contracts that task0001 must keep are stated in its task plan.

## Conventions

- **Width clamping**: any repeat count, content width or slice bound derived by subtracting fixed margins from a pane or minibuffer width is clamped at zero before use. This is the convention already followed by the header line 1 / header line 2 rendering, the output-area header of the background-output split view, and the ellipsis truncation helper in `internal/ui/pane_render.go`.
- **Row-structure invariant**: a narrow width may change the content of a row, never the number of rows. The pane header stays `paneHeaderRows` (3) rows; the minibuffer stays one row.
- **Error handling**: no new error type and no logging. Degenerate widths render degraded (possibly empty) content silently.
- **Panic regression tests**: each rendering call under test is wrapped so that a panic is captured and reported as a test failure for that case (the style of the existing narrow-width header line 1 test in `internal/ui/pane_render_test.go`), with one subtest per case so a failure names its width / input combination.

## Cross-task Design Decisions

### D1: Fix at the render sites, not in the model view

The fix lives in the pane and minibuffer rendering functions. No minimum-terminal-size check and no "screen too small" display is added to `internal/ui/model_view.go` (SPEC A1). Affects: task0001 and any rework task on this feature.

### D2: Output at non-degenerate widths is byte-identical

For pane widths ≥ 2 the three pane views (normal, background-output split, dimmed) produce the same output as before (NFR1). For a minibuffer input display width ≥ 1 the minibuffer output is the same as before (NFR3). The existing width-40 row-alignment / hit-test tests and the existing minibuffer view tests are the regression guard. Affects: task0001 and any rework task.

### D3: Single task

The minibuffer-shown pane view and the model-level test exercise both the separator fix and the minibuffer fix at once, so the work is not split: separate tasks could not make their own tests pass independently. Affects: task0001.

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Another narrow-width panic source exists on the covered render paths (FR2 also covers such sources) | Low — during planning, header line 1 / 2, entry formatting, background-output split heights, output-area header and status bar were found already clamped | Medium | Table-driven tests cover the four pane views at widths 0–4 and the model view at terminal widths 0–3. A source found inside task0001's file set is fixed under the width-clamping convention; one outside it is reported as a plan deviation. |
| A styled row wraps into several lines at a narrow width, changing the row count | Low — at widths 0–4 the styled rows' content wrap limit is ≤ 0, which the pinned styling library treats as "do not wrap" | Medium | Row-count test (pane widths 0 and 1 against width 40) and the minibuffer single-line test. |
| Output changes at normal widths and breaks row alignment used by mouse hit-testing | Low | High | D2; the existing width-40 alignment / hit-test tests stay green. |

## Open Questions

- None
