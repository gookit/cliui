# interact/ui Event Backend Filter Resize Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add optional Select/MultiSelect filtering, resize-aware list rendering, and richer backend event support for `interact/ui`.

**Architecture:** Keep the existing `Backend` and `Session` interfaces stable, extend `Event` and `View` with additive fields, and implement filtering/paging as focused helpers reused by Select and MultiSelect. `plain` remains line-based, `fake` gains testable size/resize behavior, and `readline` gains cursor visibility handling plus resize detection.

**Tech Stack:** Go 1.22, existing `interact/ui` and `interact/backend` packages, `golang.org/x/term`, standard library tests.

---

### Task 1: Backend Model And Fake Size Support

**Files:**
- Modify: `interact/backend/backend.go`
- Modify: `interact/backend/fake/fake.go`
- Test: `interact/backend/fake/fake_test.go`

- [ ] **Step 1: Write fake backend tests**

Add tests that construct `fake.NewWithOptions(nil, fake.WithSize(120, 40))`, assert `Size()` returns `120, 40`, then feed an `EventResize{Width: 90, Height: 20}` and assert the session size updates after `ReadEvent`.

- [ ] **Step 2: Extend backend structs**

Add `Width`, `Height` to `backend.Event`; add `Width`, `Height`, `HideCursor` to `backend.View`.

- [ ] **Step 3: Implement fake options**

Keep `fake.New(events ...backend.Event)` compatible, add `NewWithOptions(events []backend.Event, options ...Option)`, and add:

```go
type Option func(*Backend)

func WithSize(width, height int) Option
```

Store default size on `Backend`, copy it into each `Session`, and update session size when `ReadEvent` returns `EventResize`.

- [ ] **Step 4: Verify**

Run:

```bash
go test ./interact/backend/fake ./interact/ui
```

Expected: all tests pass.

### Task 2: Filtering, Paging, And Size Helpers

**Files:**
- Create: `interact/ui/filter.go`
- Create: `interact/ui/list_view.go`
- Create: `interact/ui/terminal_size.go`
- Test: `interact/ui/ui_test.go`

- [ ] **Step 1: Add helper tests**

Add tests for case-insensitive filtering by key/label/value, query editing with Backspace and Ctrl+U, page size calculation for zero/small terminal heights, and visible window calculation that keeps cursor visible.

- [ ] **Step 2: Implement filter helpers**

Implement a private `filterState` with query editing and `filterItems(items []Item) []int` returning original item indexes.

- [ ] **Step 3: Implement list view helpers**

Implement private helpers to calculate page size and visible indexes without changing original item order.

- [ ] **Step 4: Implement resize helpers**

Implement private `terminalSize` and `sizeFromEvent(session, ev, current)` helpers so components can handle resize events consistently.

- [ ] **Step 5: Verify**

Run:

```bash
go test ./interact/ui
```

Expected: all tests pass.

### Task 3: Select Filtering And Resize

**Files:**
- Modify: `interact/ui/select.go`
- Test: `interact/ui/ui_test.go`

- [ ] **Step 1: Add Select tests**

Cover these cases: filter `be` returns Beta; no match displays `no matched option`; Backspace restores matches; resize keeps filter query and selection; disabled filtered item cannot be submitted; `Filterable=false` existing behavior is unchanged.

- [ ] **Step 2: Add Select fields**

Add `Filterable`, `FilterPrompt`, and `PageSize` to `Select`.

- [ ] **Step 3: Update Select state machine**

Track current terminal size, filter query, filtered indexes, cursor original index, and visible window. Handle `EventResize`, filter editing keys, movement in filtered results, and Enter submission from current filtered item.

- [ ] **Step 4: Update Select rendering**

Render filter line when enabled, only render visible items when paging applies, include no-match state, and set `View.Width`, `View.Height`, and `HideCursor`.

- [ ] **Step 5: Verify**

Run:

```bash
go test ./interact/ui
```

Expected: all tests pass.

### Task 4: MultiSelect Filtering And Resize

**Files:**
- Modify: `interact/ui/multiselect.go`
- Test: `interact/ui/ui_test.go`

- [ ] **Step 1: Add MultiSelect tests**

Cover filtering then toggling a result, clearing filter while preserving selected keys, no-match Enter error, resize preserving selected keys and filter query, and disabled filtered item cannot be toggled.

- [ ] **Step 2: Add MultiSelect fields**

Add `Filterable`, `FilterPrompt`, and `PageSize` to `MultiSelect`.

- [ ] **Step 3: Update MultiSelect state machine**

Use the shared filter and paging helpers while preserving selected map semantics. Space toggles the current filtered item; Enter submits selected keys even if selected items are currently hidden by filter.

- [ ] **Step 4: Update MultiSelect rendering**

Render filter line, visible item window, current item status, selected summary, no-match state, and cursor-hidden view metadata.

- [ ] **Step 5: Verify**

Run:

```bash
go test ./interact/ui
```

Expected: all tests pass.

### Task 5: Readline Resize And Cursor Handling

**Files:**
- Modify: `interact/backend/readline/readline.go`
- Test: `interact/backend/readline/readline_test.go`

- [ ] **Step 1: Add readline tests**

Cover `Render` writing hide/show cursor escape codes when `HideCursor` changes, and a testable resize detection helper that returns an `EventResize` when size changes.

- [ ] **Step 2: Implement cursor visibility**

In `Render`, write `\x1B[?25l` for hidden cursor and `\x1B[?25h` when visible again or on `Close`.

- [ ] **Step 3: Implement resize detection**

Track last width/height in `Session`; before blocking for keyboard input, compare `Size()` with last size and return `EventResize` if changed. Keep implementation simple and avoid persistent goroutine leaks.

- [ ] **Step 4: Verify**

Run:

```bash
go test ./interact/backend/readline
```

Expected: all tests pass.

### Task 6: Documentation And Full Verification

**Files:**
- Modify: `docs/zh-CN/interact-ui.md`
- Modify: `docs/interact-ui.md`

- [ ] **Step 1: Update docs**

Document `Filterable`, `PageSize`, resize behavior, and backend differences.

- [ ] **Step 2: Run full tests**

Run:

```bash
go test ./...
```

Expected: all tests pass.

- [ ] **Step 3: Review diff**

Run:

```bash
git diff --stat
git diff --name-only
```

Expected: changes are limited to backend/ui implementation, tests, and docs/design/plan files.
