# cliui progress 后续增强实现计划

> **给 agentic workers 的要求：** 按任务执行本计划时，必须使用 `superpowers:subagent-driven-development`（推荐）或 `superpowers:executing-plans`。任务使用 checkbox（`- [ ]`）格式跟踪进度。

**目标：** 实现 `progress` 第二、第三阶段增强：render mode、TTY 检测、format 对齐/截断、动态 bar 管理和状态 helper。

**架构：** 在 `MultiProgress` 内保留现有 dynamic renderer 作为默认行为，新增 plain/disabled 渲染分支但不改变 standalone `Progress` 的输出模型。动态管理通过 `Progress` 内部 hidden/removed 状态和 manager 锁完成；状态 helper 复用现有 message、finishManaged 和 update 路径，不引入业务状态机。

**关键约束：** 先重构 renderer 和 manager update 事件语义，再实现 plain/disabled 和动态管理。`RenderPlain` 不能靠布尔参数控制输出；使用 `updateEvent` 明确区分普通进度、关键状态和终态。`removed` 判断必须在 manager 锁内完成，避免并发下载路径出现 race。

**技术栈：** Go，标准库 `fmt` / `io` / `os` / `regexp` / `time`，现有 `progress` 包测试和中英文文档。

---

## 范围

本计划实现 `docs/superpowers/specs/2026-05-09-cliui-progress-followup-enhancement-design.md` 中的第二阶段和第三阶段。

包含：

- `RenderMode` 类型。
- `MultiProgress.RenderMode` 字段。
- `RenderDynamic` / `RenderPlain` / `RenderDisabled`。
- `IsTerminal(w io.Writer) bool`。
- format token 支持 `{@name:-12s}` 和 `{@name:.20s}`。
- `Hide()` / `Show()` / `Remove()`。
- `Progress.Done()` / `Fail()` / `Skip()`。
- 中英文文档更新。

暂不包含：

- `ByteTracker`。
- `NewConcurrentWriter()`。
- `Slot` 类型。
- `UseAutoRenderMode()` 便捷方法。
- missing token 默认渲染为空的行为。
- `SpinnerFactory` 纳入 `MultiProgress`。

## 文件职责

- 修改：`progress/multi.go`
  - 增加 `RenderMode` 类型和 `MultiProgress.RenderMode`。
  - 增加 `updateEvent` 类型，替代模糊的 plain 输出布尔参数。
  - 将现有 ANSI 多行刷新明确为 dynamic renderer。
  - 抽出 dynamic block 移动、清理和渲染 helper。
  - 增加 plain/disabled 渲染分支。
  - 增加 `Hide()` / `Show()` / `Remove()`。
  - 调整 `Len()` / `VisibleLen()`。

- 修改：`progress/progress.go`
  - 增加 `hidden` / `removed` 内部字段。
  - 让 managed update 对 removed progress no-op。
  - 增加 `Done()` / `Fail()` / `Skip()`。
  - 增强 `widgetMatch` 正则。

- 新建或修改：`progress/terminal.go`
  - 提供 `IsTerminal(w io.Writer) bool`。

- 修改：`progress/multi_test.go`
  - 增加 render mode、dynamic management 测试。

- 修改：`progress/progress_test.go`
  - 增加 format token 和 status helper 测试。

- 修改：`docs/zh-CN/progress.md`
  - 说明 render mode、TTY 检测、format 截断、动态管理、状态 helper。

- 修改：`docs/progress.md`
  - 增加对应英文说明。

---

### Task 1: Renderer 结构和 RenderMode 基础模型

**Files:**
- Modify: `progress/multi.go`
- Test: `progress/multi_test.go`

- [x] **Step 1: 添加 RenderDynamic 兼容测试**

在 `progress/multi_test.go` 添加：

```go
func TestMultiProgressRenderDynamicKeepsANSIBlock(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.RenderMode = RenderDynamic

	p := mp.New(2)
	p.AddMessage("message", " task")
	mp.Start()
	p.Advance()
	mp.Finish()

	out := buf.String()
	is.Contains(out, "\x1B[2K")
	is.Contains(out, " task")
}
```

- [x] **Step 2: 运行测试并确认失败**

运行：

```bash
go test ./progress -run TestMultiProgressRenderDynamicKeepsANSIBlock -count=1
```

预期：编译失败，原因是 `RenderDynamic` 和 `MultiProgress.RenderMode` 还不存在。

- [x] **Step 3: 增加 RenderMode 类型和字段**

在 `progress/multi.go` 添加：

```go
type RenderMode int

const (
	RenderDynamic RenderMode = iota
	RenderPlain
	RenderDisabled
)
```

给 `MultiProgress` 增加公开字段：

```go
RenderMode RenderMode
```

默认零值 `RenderDynamic` 保持现有行为。

- [x] **Step 4: 重构 dynamic renderer**

将当前 `refreshLocked()` 的 ANSI 多行逻辑移动到：

```go
func (mp *MultiProgress) refreshDynamicLocked()
```

`refreshLocked()` 暂时只分发默认 dynamic 分支：

```go
func (mp *MultiProgress) refreshLocked() {
	switch mp.RenderMode {
	default:
		mp.refreshDynamicLocked()
	}
}
```

- [x] **Step 5: 抽出 dynamic block helper**

在 `progress/multi.go` 中抽出：

```go
func (mp *MultiProgress) moveToBlockStartLocked()
func (mp *MultiProgress) clearRenderedBlockLocked()
func (mp *MultiProgress) renderDynamicLinesLocked(lines []string)
func (mp *MultiProgress) currentLinesLocked() []string
```

实现要求：

- `moveToBlockStartLocked()` 使用 `lastLines` 移动到旧 block 起点。
- `clearRenderedBlockLocked()` 使用 `lastLines` 清理完整旧 block。
- `renderDynamicLinesLocked(lines)` 渲染传入行并更新 `rendered` / `lastLines`。
- `currentLinesLocked()` 基于当前可见 bars 生成 `Line()` 文本；Task 1 暂时所有 bars 都可见。

这一步是为了后续 hide/remove 时能清理旧 block 多余行，避免残留。

- [x] **Step 6: 验证 Task 1**

运行：

```bash
go test ./progress -run TestMultiProgressRenderDynamicKeepsANSIBlock -count=1
go test ./progress
```

预期：新增测试和 `progress` 包测试通过。

- [ ] **Step 7: 提交 Task 1**

运行：

```bash
git add progress/multi.go progress/multi_test.go
git commit -m "refactor(progress): prepare multi progress render modes"
```

预期：如果当前会话允许提交，commit 成功。

---

### Task 2: UpdateEvent、RenderPlain 和 RenderDisabled

**Files:**
- Modify: `progress/multi.go`
- Modify: `progress/progress.go`
- Test: `progress/multi_test.go`

- [ ] **Step 1: 添加 RenderPlain 不输出 ANSI 的失败测试**

在 `progress/multi_test.go` 添加：

```go
func TestMultiProgressRenderPlainDoesNotUseANSIBlock(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.RenderMode = RenderPlain

	p := mp.New(3)
	p.AddMessage("message", " task")
	mp.Start()
	p.Advance()
	mp.Refresh()
	mp.Finish()

	out := buf.String()
	is.NotContains(out, "\x1B[2K")
	is.NotContains(out, "\x1B[")
	is.Contains(out, " task")
}
```

- [ ] **Step 2: 添加 RenderPlain 下 Advance 不刷屏的失败测试**

在 `progress/multi_test.go` 添加：

```go
func TestMultiProgressRenderPlainAdvanceDoesNotPrint(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.RenderMode = RenderPlain

	p := mp.New(10)
	mp.Start()
	size := buf.Len()
	p.Advance()
	p.Advance()

	is.Eq(size, buf.Len())
	mp.Refresh()
	is.True(buf.Len() > size)
}
```

- [ ] **Step 3: 添加 RenderPlain Reset 输出关键事件的失败测试**

在 `progress/multi_test.go` 添加：

```go
func TestMultiProgressRenderPlainResetPrintsKeyState(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.RenderMode = RenderPlain

	p := mp.New(10)
	p.SetFormat("{@name} {@percent}%")
	p.SetMessage("name", "fd")
	mp.Start()
	size := buf.Len()

	p.Reset(20)
	p.SetMessage("name", "bat")

	out := buf.String()[size:]
	is.Contains(out, "fd 0.0%")
	is.NotContains(out, "\x1B[")
}
```

说明：`Reset()` 是关键状态事件，plain mode 应输出该 bar 的当前行；后续 `SetMessage()` 不要求立即输出。

- [ ] **Step 4: 添加 RenderPlain 下 Progress.Finish 输出且 manager Finish 不重复输出的失败测试**

在 `progress/multi_test.go` 添加：

```go
func TestMultiProgressRenderPlainProgressFinishPrintsOnce(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.RenderMode = RenderPlain

	p := mp.New(2)
	p.SetFormat("{@message} {@percent}%")
	mp.Start()
	p.Finish("done")
	afterProgressFinish := strings.Count(buf.String(), "done 100.0%")
	mp.Finish()
	afterManagerFinish := strings.Count(buf.String(), "done 100.0%")

	is.Eq(1, afterProgressFinish)
	is.Eq(1, afterManagerFinish)
}
```

- [ ] **Step 5: 添加 RenderDisabled 的失败测试**

在 `progress/multi_test.go` 添加：

```go
func TestMultiProgressRenderDisabledSuppressesProgress(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.RenderMode = RenderDisabled

	p := mp.New(3)
	p.AddMessage("message", " hidden")
	mp.Start()
	p.Advance()
	mp.Refresh()
	p.Finish()
	mp.Finish()

	is.Eq("", buf.String())
}
```

- [ ] **Step 6: 添加 RenderDisabled 仍允许日志输出的失败测试**

在 `progress/multi_test.go` 添加：

```go
func TestMultiProgressRenderDisabledStillPrintsLogs(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.RenderMode = RenderDisabled
	mp.New(1)
	mp.Start()

	mp.Println("warning: fallback")
	mp.Printf("package %s failed\n", "fd")
	mp.Finish()

	out := buf.String()
	is.Contains(out, "warning: fallback")
	is.Contains(out, "package fd failed")
	is.NotContains(out, "\x1B[")
}
```

- [ ] **Step 7: 运行测试并确认失败**

运行：

```bash
go test ./progress -run "TestMultiProgressRender(Plain|Disabled)" -count=1
```

预期：测试失败，因为 plain/disabled 还没有专用渲染分支。

- [ ] **Step 8: 增加 updateEvent**

在 `progress/multi.go` 增加：

```go
type updateEvent int

const (
	updateSilent updateEvent = iota
	updateProgress
	updateKeyState
	updateFinalState
)
```

将：

```go
func (mp *MultiProgress) update(fn func() bool)
```

改成：

```go
func (mp *MultiProgress) update(event updateEvent, fn func() bool)
```

事件规则：

- `AdvanceTo()` 使用 `updateProgress`。
- `Reset()` / `ResetWith()` 使用 `updateKeyState`。
- `Progress.Finish()` 使用 `updateFinalState`。
- setter/message/widget 更新使用 `updateKeyState`，但 plain mode 可以选择不立即输出，避免频繁日志；实现时以 `Reset()` 和 final events 为 plain 输出重点。

- [ ] **Step 9: 改造 `refreshLocked()` mode 分发**

将 `refreshLocked()` 改成按 mode 分发：

```go
func (mp *MultiProgress) refreshLocked() {
	switch mp.RenderMode {
	case RenderDisabled:
		return
	case RenderPlain:
		mp.refreshPlainLocked()
	default:
		mp.refreshDynamicLocked()
	}
}
```

新增或完善：

```go
func (mp *MultiProgress) refreshPlainLocked()
```

plain refresh 行为：

- `!mp.started || mp.finished` 时返回。
- 遍历可见 bars。
- 每个 bar 输出 `p.Line()` + newline。
- 不输出任何 ANSI 控制符。
- 不设置 `rendered = true`。

- [ ] **Step 10: 调整 update 事件处理**

`MultiProgress.update(event, fn)` 的行为：

- 先在锁内执行 `changed := fn()`。
- `changed == false` 直接返回。
- `RenderDisabled`：不输出。
- `RenderDynamic`：按现有逻辑刷新或 dirty。
- `RenderPlain`：
  - `updateProgress`：不输出。
  - `updateKeyState`：只对 `Reset()` 这类调用方明确需要的事件输出当前 bar 行；如果实现无法识别具体 bar，允许调用方通过专用 `updateBar(event, p, fn)` 传入 bar。
  - `updateFinalState`：输出当前 bar 行。

推荐实现：

```go
func (mp *MultiProgress) update(event updateEvent, fn func() bool) {
	mp.updateBar(event, nil, fn)
}

func (mp *MultiProgress) updateBar(event updateEvent, p *Progress, fn func() bool) {
	// holds mp.mu, applies fn, then render behavior based on mode/event.
}
```

`Progress.Reset()`、`Progress.Finish()`、后续 status helper 走 `updateBar(..., p, ...)`，plain mode 才能只输出相关 bar。

- [ ] **Step 11: 调整 `Start()` / `Finish()` / `RunExclusive()`**

`Start()`：

- `RenderDynamic`：现有行为。
- `RenderPlain`：输出当前所有 bar 行。
- `RenderDisabled`：不输出，不启动 auto refresh。

`Finish()`：

- `RenderDynamic`：现有行为。
- `RenderPlain`：只结束 manager 生命周期，不重复输出全部最终行。
- `RenderDisabled`：不输出 progress。

`RunExclusive()`：

- `RenderDynamic`：现有 clear + fn + redraw。
- `RenderPlain`：直接执行 `fn(mp.writer())`，不自动 redraw。
- `RenderDisabled`：直接执行 `fn(mp.writer())`。

- [ ] **Step 12: AutoRefresh 与 render mode 交互**

`Start()` 只有在 `RenderMode == RenderDynamic && AutoRefresh` 时启动 ticker。

`update()` 在 `RenderPlain` / `RenderDisabled` 下不启动或依赖 auto refresh 输出。

- [ ] **Step 13: 验证 Task 2**

运行：

```bash
go test ./progress -run "TestMultiProgressRender(Plain|Disabled)" -count=1
go test ./progress
```

预期：新增 tests 和 `progress` 包测试通过。

- [ ] **Step 14: 提交 Task 2**

运行：

```bash
git add progress/multi.go progress/progress.go progress/multi_test.go
git commit -m "feat(progress): add plain and disabled render modes"
```

预期：如果当前会话允许提交，commit 成功。

---

### Task 3: TTY 检测

**Files:**
- Create: `progress/terminal.go`
- Test: `progress/multi_test.go`

- [ ] **Step 1: 添加 `IsTerminal` 失败测试**

在 `progress/multi_test.go` 添加：

```go
func TestIsTerminalReturnsFalseForBuffer(t *testing.T) {
	is := assert.New(t)
	is.False(IsTerminal(new(bytes.Buffer)))
}
```

- [ ] **Step 2: 添加普通文件检测测试**

在 `progress/multi_test.go` 添加：

```go
func TestIsTerminalReturnsFalseForRegularFile(t *testing.T) {
	is := assert.New(t)
	file, err := os.CreateTemp("", "cliui-progress-terminal-*")
	is.NoErr(err)
	defer os.Remove(file.Name())
	defer file.Close()

	is.False(IsTerminal(file))
}
```

同时给 `progress/multi_test.go` 增加 `os` import。

- [ ] **Step 3: 运行测试并确认失败**

运行：

```bash
go test ./progress -run TestIsTerminal -count=1
```

预期：编译失败，因为 `IsTerminal` 还不存在。

- [ ] **Step 4: 新增 `progress/terminal.go`**

创建 `progress/terminal.go`：

```go
package progress

import (
	"io"
	"os"

	"golang.org/x/term"
)

// IsTerminal reports whether w is an interactive terminal.
func IsTerminal(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}

	return term.IsTerminal(int(file.Fd()))
}
```

项目当前已经依赖 `golang.org/x/term`，直接复用该依赖，不需要新增 go.mod/go.sum 变更。

- [ ] **Step 5: 验证 Task 3**

运行：

```bash
go test ./progress -run TestIsTerminal -count=1
go test ./progress
```

预期：新增 tests 和 `progress` 包测试通过。

- [ ] **Step 6: 提交 Task 3**

运行：

```bash
git add progress/terminal.go progress/multi_test.go go.mod go.sum
git commit -m "feat(progress): add terminal detection helper"
```

预期：如果没有 go.mod/go.sum 变化，`git add` 会忽略无变化文件，commit 仍成功。

---

### Task 4: Format Token 对齐和截断

**Files:**
- Modify: `progress/progress.go`
- Test: `progress/progress_test.go`

- [ ] **Step 1: 添加左对齐和截断失败测试**

在 `progress/progress_test.go` 添加：

```go
func TestProgressFormatSupportsAlignmentAndTruncation(t *testing.T) {
	is := assert.New(t)
	p := New()
	p.SetFormat("{@name:-8s}|{@name:.3s}|{@percent:5s}|{@missing:.2s}")
	p.SetMessage("name", "abcdef")
	p.Start()

	is.Eq("abcdef  |abc|  0.0|{@missing:.2s}", p.Line())
}
```

- [ ] **Step 2: 添加既有格式兼容测试**

在 `progress/progress_test.go` 添加：

```go
func TestProgressFormatKeepsExistingWidgetFormats(t *testing.T) {
	is := assert.New(t)
	p := New()
	p.SetFormat("{@percent:4s}|{@estimated:-7s}")
	p.Start()

	line := p.Line()
	is.Contains(line, " 0.0")
	is.Contains(line, "unknown")
}
```

- [ ] **Step 3: 运行测试并确认失败**

运行：

```bash
go test ./progress -run "TestProgressFormat" -count=1
```

预期：第一个测试失败，因为 `widgetMatch` 当前不支持 `.`。

- [ ] **Step 4: 修改 widget 正则**

在 `progress/progress.go` 将：

```go
var widgetMatch = regexp.MustCompile(`{@([\w_]+)(?::([\w-]+))?}`)
```

改为：

```go
var widgetMatch = regexp.MustCompile(`{@([\w_]+)(?::([^}]+))?}`)
```

不修改 unknown token 默认行为。

- [ ] **Step 5: 验证 Task 4**

运行：

```bash
go test ./progress -run "TestProgressFormat" -count=1
go test ./progress
```

预期：新增 tests 和 `progress` 包测试通过。

- [ ] **Step 6: 提交 Task 4**

运行：

```bash
git add progress/progress.go progress/progress_test.go
git commit -m "feat(progress): support wider progress format verbs"
```

预期：如果当前会话允许提交，commit 成功。

---

### Task 5: Hide / Show / Remove 动态管理

**Files:**
- Modify: `progress/multi.go`
- Modify: `progress/progress.go`
- Test: `progress/multi_test.go`

- [ ] **Step 1: 添加 Hide/Show 失败测试**

在 `progress/multi_test.go` 添加：

```go
func TestMultiProgressHideAndShow(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf

	p1 := mp.New(1)
	p1.SetMessage("message", " one")
	p2 := mp.New(1)
	p2.SetMessage("message", " two")
	mp.Start()

	mp.Hide(p2)
	is.Eq(2, mp.Len())
	is.Eq(1, mp.VisibleLen())
	buf.Reset()
	mp.Refresh()
	is.Contains(buf.String(), " one")
	is.NotContains(buf.String(), " two")

	mp.Show(p2)
	is.Eq(2, mp.Len())
	is.Eq(2, mp.VisibleLen())
	buf.Reset()
	mp.Refresh()
	is.Contains(buf.String(), " two")
}
```

这个测试通过 `buf.Reset(); mp.Refresh()` 检查新一次刷新输出，避免从历史 ANSI buffer 中截取最后一个清行导致误判。

- [ ] **Step 2: 添加 Remove 失败测试**

在 `progress/multi_test.go` 添加：

```go
func TestMultiProgressRemove(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf

	p1 := mp.New(1)
	p1.SetMessage("message", " one")
	p2 := mp.New(1)
	p2.SetMessage("message", " two")
	mp.Start()

	mp.Remove(p2)
	is.Eq(1, mp.Len())
	is.Eq(1, mp.VisibleLen())
	p2.Advance()
	p2.SetMessage("message", " removed")
	p2.Finish()
	buf.Reset()
	mp.Refresh()

	out := buf.String()
	is.Contains(out, " one")
	is.NotContains(out, " removed")
}
```

- [ ] **Step 3: 添加 dynamic 行数减少时清理旧 block 的失败测试**

在 `progress/multi_test.go` 添加：

```go
func TestMultiProgressHideClearsStaleDynamicLines(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf

	p1 := mp.New(1)
	p1.SetMessage("message", " one")
	p2 := mp.New(1)
	p2.SetMessage("message", " two")
	mp.Start()

	buf.Reset()
	mp.Hide(p2)
	out := buf.String()

	is.Contains(out, "\x1B[2K")
	is.Contains(out, " one")
	is.NotContains(out, " two")
}
```

这个测试要求 hide/remove 导致可见行减少时，dynamic renderer 清理旧 block，不能留下旧的第二行。

- [ ] **Step 4: 运行测试并确认失败**

运行：

```bash
go test ./progress -run "TestMultiProgress(HideAndShow|Remove|HideClearsStaleDynamicLines)" -count=1
```

预期：编译失败，因为 `Hide()` / `Show()` / `Remove()` 不存在。

- [ ] **Step 5: 给 `Progress` 增加内部状态**

在 `Progress` 结构体增加：

```go
hidden  bool
removed bool
```

- [ ] **Step 6: 实现 visible bar helper**

在 `progress/multi.go` 增加：

```go
func (mp *MultiProgress) visibleBarsLocked() []*Progress
```

返回 `!p.hidden && !p.removed` 的 bars。

`refreshDynamicLocked()` 和 `refreshPlainLocked()` 都改用 visible bars。

- [ ] **Step 7: 实现 Hide / Show**

在 `progress/multi.go` 添加：

```go
func (mp *MultiProgress) Hide(p *Progress)
func (mp *MultiProgress) Show(p *Progress)
```

实现要求：

- 持有 `mp.mu`。
- 只处理 `p != nil && p.manager == mp && !p.removed`。
- `Hide()` 设置 `p.hidden = true`。
- `Show()` 设置 `p.hidden = false`。
- dynamic mode 下刷新或 auto refresh dirty；如果可见行数减少，必须使用 `clearRenderedBlockLocked()` 清理旧 block。
- plain/disabled mode 下不自动输出。

- [ ] **Step 8: 实现 Remove**

在 `progress/multi.go` 添加：

```go
func (mp *MultiProgress) Remove(p *Progress)
```

实现要求：

- 持有 `mp.mu`。
- 只处理 `p != nil && p.manager == mp && !p.removed`。
- 从 `mp.bars` slice 中移除该 bar。
- 对 remaining bars 重建 `index`。
- 设置 `p.removed = true`。
- 保留 `p.manager = mp`，让后续 update 能 no-op，不退化成 standalone 输出。
- dynamic mode 下刷新或 auto refresh dirty；如果可见行数减少，必须清理旧 block。

- [ ] **Step 9: 让 removed progress 更新 no-op**

调整 `Progress` 的托管更新路径：

- `SetMaxSteps`
- `SetFormat`
- `SetMessage`
- `SetMessages`
- `AddMessage`
- `AddMessages`
- `AddWidget`
- `SetWidget`
- `AddWidgets`
- `Reset`
- `ResetWith`
- `AdvanceTo`
- `Finish`
- `Display`

共同规则：

```go
// must run while holding mp.mu
if p.removed {
	return
}
```

对于返回 `*Progress` 的方法，直接返回 `p`。

不要在未持有 `mp.mu` 时读取 `p.removed`。这些方法的 managed 分支应先进入 `mp.update(...)` 或 `mp.updateBar(...)`，在 closure 中检查 `p.removed` 并返回 `false`。

- [ ] **Step 10: 调整 Len / VisibleLen**

`Len()` 返回 `len(mp.bars)`。

`VisibleLen()` 返回 `len(mp.visibleBarsLocked())`。

- [ ] **Step 11: 验证 Task 5**

运行：

```bash
go test ./progress -run "TestMultiProgress(HideAndShow|Remove|HideClearsStaleDynamicLines)" -count=1
go test ./progress
```

预期：新增 tests 和 `progress` 包测试通过。

- [ ] **Step 12: 提交 Task 5**

运行：

```bash
git add progress/multi.go progress/progress.go progress/multi_test.go
git commit -m "feat(progress): support dynamic multi progress bars"
```

预期：如果当前会话允许提交，commit 成功。

---

### Task 6: Done / Fail / Skip 状态 Helper

**Files:**
- Modify: `progress/progress.go`
- Test: `progress/progress_test.go`

- [ ] **Step 1: 添加 Done 失败测试**

在 `progress/progress_test.go` 添加：

```go
func TestProgressDone(t *testing.T) {
	is := assert.New(t)
	p := New(10)
	p.SetFormat("{@message}|{@status}|{@current}/{@max}")
	p.Start()
	p.AdvanceTo(3)

	p.Done()

	is.True(p.Finished())
	is.Eq(int64(10), p.Step())
	is.Eq("done|done|10/10", p.Line())
}
```

- [ ] **Step 2: 添加 Fail / Skip 失败测试**

在 `progress/progress_test.go` 添加：

```go
func TestProgressFailAndSkip(t *testing.T) {
	is := assert.New(t)

	failed := New(10)
	failed.SetFormat("{@message}|{@status}|{@current}/{@max}")
	failed.Start()
	failed.AdvanceTo(4)
	failed.Fail("network failed")

	is.True(failed.Finished())
	is.Eq(int64(4), failed.Step())
	is.Eq("network failed|failed| 4/10", failed.Line())

	skipped := New(10)
	skipped.SetFormat("{@message}|{@status}|{@current}/{@max}")
	skipped.Start()
	skipped.AdvanceTo(2)
	skipped.Skip()

	is.True(skipped.Finished())
	is.Eq(int64(2), skipped.Step())
	is.Eq("skipped|skipped| 2/10", skipped.Line())
}
```

- [ ] **Step 3: 添加 managed helper 测试**

在 `progress/progress_test.go` 添加：

```go
func TestProgressStatusHelpersManaged(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf

	p := mp.New(5)
	p.SetFormat("{@message}|{@status}|{@current}/{@max}")
	mp.Start()
	p.AdvanceTo(2)
	p.Fail()
	mp.Finish()

	is.Contains(buf.String(), "failed|failed|2/5")
	is.True(p.Finished())
	is.Eq(int64(2), p.Step())
}
```

- [ ] **Step 4: 添加 RenderPlain status helper 输出测试**

在 `progress/progress_test.go` 添加：

```go
func TestProgressStatusHelpersRenderPlainPrintFinalState(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.RenderMode = RenderPlain

	p := mp.New(5)
	p.SetFormat("{@message}|{@status}|{@current}/{@max}")
	mp.Start()
	size := buf.Len()
	p.Fail("network failed")

	out := buf.String()[size:]
	is.Contains(out, "network failed|failed|0/5")
	is.NotContains(out, "\x1B[")
}
```

- [ ] **Step 5: 运行测试并确认失败**

运行：

```bash
go test ./progress -run "TestProgress(Done|FailAndSkip|StatusHelpersManaged|StatusHelpersRenderPlain)" -count=1
```

预期：编译失败，因为 `Done()` / `Fail()` / `Skip()` 不存在。

- [ ] **Step 6: 实现内部 status helper**

在 `progress/progress.go` 增加：

```go
func (p *Progress) finishStatus(status string, complete bool, message ...string)
```

语义：

- 如果 `p.removed`，直接返回。
- 设置 `finishedAt = time.Now()`。
- 如果 `complete == true`，推进到 `MaxSteps`；`MaxSteps == 0` 时设置为当前 step。
- 设置 `message` 和 `status` messages。
- 通过 manager update 路径刷新或标记 dirty。
- managed 模式下 removed 判断必须在 manager 锁内执行。
- `RenderPlain` 下应作为 `updateFinalState` 输出当前 bar 行。

- [ ] **Step 7: 实现 Done / Fail / Skip**

添加：

```go
func (p *Progress) Done(message ...string)
func (p *Progress) Fail(message ...string)
func (p *Progress) Skip(message ...string)
```

默认文本：

- `Done()` 默认 `done`。
- `Fail()` 默认 `failed`。
- `Skip()` 默认 `skipped`。

完成策略：

- `Done()` 调用 `finishStatus("done", true, message...)`。
- `Fail()` 调用 `finishStatus("failed", false, message...)`。
- `Skip()` 调用 `finishStatus("skipped", false, message...)`。

- [ ] **Step 8: 验证 Task 6**

运行：

```bash
go test ./progress -run "TestProgress(Done|FailAndSkip|StatusHelpersManaged|StatusHelpersRenderPlain)" -count=1
go test ./progress
```

预期：新增 tests 和 `progress` 包测试通过。

- [ ] **Step 9: 提交 Task 6**

运行：

```bash
git add progress/progress.go progress/progress_test.go
git commit -m "feat(progress): add progress status helpers"
```

预期：如果当前会话允许提交，commit 成功。

---

### Task 7: 文档和完整验证

**Files:**
- Modify: `docs/zh-CN/progress.md`
- Modify: `docs/progress.md`
- Verify: `progress/multi.go`
- Verify: `progress/progress.go`
- Verify: `progress/terminal.go`
- Verify: `progress/multi_test.go`
- Verify: `progress/progress_test.go`

- [ ] **Step 1: 更新中文文档**

在 `docs/zh-CN/progress.md` 的 `Multi Progress` 部分补充：

- `RenderMode` 三种模式说明。
- `IsTerminal()` 推荐用法。
- `RenderPlain` 不适合高频逐步输出，只输出关键事件。
- `RenderDisabled` 仍允许安全日志输出。
- `Hide()` / `Show()` / `Remove()`。
- `Done()` / `Fail()` / `Skip()`。
- format token 左对齐和截断示例。

核心示例：

```go
mp := progress.NewMulti()
mp.Writer = os.Stderr
if !progress.IsTerminal(os.Stderr) {
	mp.RenderMode = progress.RenderPlain
}
```

format 示例：

```go
bar.SetFormat("{@slot} {@name:-12s} {@name:.20s} {@percent:5s}%")
```

- [ ] **Step 2: 更新英文文档**

在 `docs/progress.md` 添加对应英文说明。

术语保持一致：

- render mode
- dynamic rendering
- plain rendering
- disabled rendering
- dynamic bar management
- status helpers

- [ ] **Step 3: 运行 progress 测试**

运行：

```bash
go test ./progress
```

预期：`progress` 包测试全部通过。

- [ ] **Step 4: 运行全仓库测试**

运行：

```bash
go test ./...
```

预期：全部测试通过。如果无关包失败，记录失败 package 和关键错误文本，再判断是否与本次改动相关。

- [ ] **Step 5: 检查 diff 范围**

运行：

```bash
git diff --stat
git diff --name-only
```

预期变更文件：

- `progress/multi.go`
- `progress/progress.go`
- `progress/terminal.go`
- `progress/multi_test.go`
- `progress/progress_test.go`
- `docs/zh-CN/progress.md`
- `docs/progress.md`
- `docs/superpowers/plans/2026-05-09-cliui-progress-followup-enhancement.md`

- [ ] **Step 6: 提交 Task 7**

运行：

```bash
git add docs/zh-CN/progress.md docs/progress.md
git commit -m "docs(progress): document followup progress enhancements"
```

预期：如果当前会话允许提交，commit 成功。

---

## 最终验收标准

- `go test ./progress` 通过。
- `go test ./...` 通过，或只存在已记录且与本次改动无关的失败。
- 默认 `MultiProgress` 行为仍是 dynamic ANSI 多行刷新。
- `RenderPlain` 不输出 ANSI 控制符，且不会因每次 `Advance()` 刷屏。
- `RenderPlain` 下 `Reset()` 输出关键状态行。
- `RenderPlain` 下 `Progress.Finish()` / status helper 输出终态行，`MultiProgress.Finish()` 不重复输出全部 bar。
- `RenderDisabled` 不输出 progress，但安全日志 API 仍输出日志。
- `IsTerminal()` 对 `bytes.Buffer` 和普通文件返回 false。
- `{@name:-12s}`、`{@name:.20s}`、`{@percent:5s}` 生效。
- unknown token 继续原样保留。
- `Hide()` / `Show()` / `Remove()` 正确影响渲染和 `VisibleLen()`。
- dynamic mode 下隐藏或移除导致行数减少时，旧 block 底部不会残留已隐藏/已移除行。
- `Remove()` 后迟到的 progress 更新 no-op，不 panic，不写 standalone 输出。
- `removed` 判断在 manager 锁内执行，避免并发访问数据竞争。
- `Done()` / `Fail()` / `Skip()` 更新状态和 message，不直接输出额外日志。
- 中英文文档包含新增 API 和推荐用法。

## 实施注意事项

- 不实现 `ByteTracker` 和 `NewConcurrentWriter()`。
- 不引入 `Slot` 类型。
- 不改变 unknown token 默认行为。
- 不让 `Progress` standalone 行为依赖 `RenderMode`。
- 不让 `RenderPlain` 在每次 `Advance()` 自动输出。
- 不让 `Remove()` 后的 progress 退化成 standalone 输出。
- 不在未持有 manager 锁时读取 managed progress 的 `removed` 状态。
- `golang.org/x/term` 已存在于 `go.mod`，实现 `IsTerminal()` 时直接复用，不主动改依赖。
