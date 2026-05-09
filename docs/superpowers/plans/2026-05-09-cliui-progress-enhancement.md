# cliui progress 增强开发计划

> **给 agentic workers 的要求：** 按任务执行本计划时，必须使用 `superpowers:subagent-driven-development`（推荐）或 `superpowers:executing-plans`。任务使用 checkbox（`- [ ]`）格式跟踪进度。

**目标：** 实现 `progress` 第一阶段增强能力，用于支撑稳定的并发下载终端 UI：`MultiProgress` 刷新节流、可复用 progress slot、安全日志输出和生命周期查询。

**架构：** `MultiProgress` 继续作为托管进度条的唯一 writer。所有托管状态变更都经过 manager 锁，并由状态更新函数返回是否产生可见变化；`AutoRefresh` 决定同步刷新还是 dirty + ticker 刷新。`Progress` 增加 reset/setter API，但第一阶段不引入 `Slot` 抽象。

**技术栈：** Go，标准库 `sync` / `time` / `io` / `fmt`，现有 `progress` 包测试和文档。

---

## 范围

本计划只实现 `docs/superpowers/specs/2026-05-09-cliui-progress-enhancement-design.md` 中定义的第一阶段能力。

包含：

- `MultiProgress.AutoRefresh` 和 `RefreshInterval`。
- 托管模式下 `Progress.RedrawFreq` 生效。
- `Progress.Reset()` / `ResetWith()` / setter API。
- `Progress` 和 `MultiProgress` 生命周期 getter。
- `MultiProgress.RunExclusive()` / `Println()` / `Printf()`。
- 新增 API 的中英文文档。

暂不包含：

- `Hide()` / `Show()` / `Remove()`。
- TTY 检测和 render mode。
- `Done()` / `Fail()` / `Skip()`。
- `NewConcurrentWriter()` / `ByteTracker`。
- format token 截断或 missing token 行为调整。

## 文件职责

- 修改：`progress/multi.go`
  - 增加 dirty/ticker 刷新状态。
  - 将托管 update callback 改为返回 `bool`。
  - 增加安全日志输出 helper 和 `MultiProgress` getter。

- 修改：`progress/progress.go`
  - 让托管模式下 `AdvanceTo()` 使用 `applyStep()` 返回值控制刷新。
  - 增加 `Reset()` / `ResetWith()` / setter API。
  - 增加生命周期 getter。
  - 更新已有 message/widget 托管更新调用，适配新的 update 签名。

- 修改：`progress/multi_test.go`
  - 增加 redraw frequency、auto refresh、安全日志和 manager getter 测试。

- 修改：`progress/progress_test.go`
  - 增加 reset、setter 和 progress 生命周期 getter 测试。

- 修改：`docs/zh-CN/progress.md`
  - 说明 auto refresh、slot 复用、安全日志和生命周期方法。

- 修改：`docs/progress.md`
  - 增加对应英文说明。

---

### Task 1: 托管刷新频率和自动刷新

**Files:**
- Modify: `progress/multi.go`
- Modify: `progress/progress.go`
- Test: `progress/multi_test.go`

- [ ] **Step 1: 添加托管模式 `RedrawFreq` 的失败测试**

在 `progress/multi_test.go` 添加：

```go
func TestMultiProgressRedrawFreqInManagedMode(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf

	p := mp.New(10)
	p.RedrawFreq = 3
	mp.Start()

	initial := strings.Count(buf.String(), "\x1B[2K")
	p.Advance()
	p.Advance()
	is.Eq(initial, strings.Count(buf.String(), "\x1B[2K"))

	p.Advance()
	is.True(strings.Count(buf.String(), "\x1B[2K") > initial)
}
```

- [ ] **Step 2: 添加 auto refresh dirty 刷新的失败测试**

在 `progress/multi_test.go` 添加：

```go
func TestMultiProgressAutoRefreshMarksDirty(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.AutoRefresh = true
	mp.RefreshInterval = 20 * time.Millisecond

	p := mp.New(10)
	mp.Start()
	initial := strings.Count(buf.String(), "\x1B[2K")
	p.AdvanceTo(5)
	is.Eq(initial, strings.Count(buf.String(), "\x1B[2K"))

	time.Sleep(60 * time.Millisecond)
	is.Contains(buf.String(), "50.0%")
	mp.Finish()
}
```

- [ ] **Step 3: 添加 `Finish()` 停止 auto refresh 的失败测试**

在 `progress/multi_test.go` 添加：

```go
func TestMultiProgressFinishStopsAutoRefresh(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.AutoRefresh = true
	mp.RefreshInterval = 10 * time.Millisecond

	p := mp.New(10)
	mp.Start()
	p.AdvanceTo(3)
	mp.Finish()

	size := buf.Len()
	time.Sleep(40 * time.Millisecond)
	is.Eq(size, buf.Len())
}
```

- [ ] **Step 4: 运行测试并确认失败**

运行：

```bash
go test ./progress -run "TestMultiProgress(RedrawFreq|AutoRefresh|FinishStops)" -count=1
```

预期：新增测试至少有失败项，因为当前托管更新仍然同步刷新，`AutoRefresh` 还没有 ticker。

- [ ] **Step 5: 给 `MultiProgress` 增加刷新状态**

在 `progress/multi.go` 添加默认刷新间隔：

```go
const DefaultRefreshInterval = 100 * time.Millisecond
```

给 `MultiProgress` 增加内部字段：

```go
dirty  bool
stopCh chan struct{}
doneCh chan struct{}
```

- [ ] **Step 6: 修改 manager update callback 签名**

将：

```go
func (mp *MultiProgress) update(fn func())
```

改为：

```go
func (mp *MultiProgress) update(fn func() bool)
```

实现语义：

- 持有 `mp.mu` 时执行 `changed := fn()`。
- `changed == false` 时直接返回。
- `mp.AutoRefresh == true` 时设置 `mp.dirty = true` 并返回。
- 否则调用 `mp.refreshLocked()` 同步刷新。

- [ ] **Step 7: 更新 `progress.go` 中的托管调用**

将所有 `p.manager.update` 调用改为返回 `bool`。

`AdvanceTo()` 使用 `applyStep()` 的返回值：

```go
p.manager.update(func() bool {
	return p.applyStep(step)
})
```

message/widget/config 变更在更新状态后返回 `true`。

托管 `Finish()` 在 `p.finishManaged(message...)` 后返回 `true`。

- [ ] **Step 8: 实现 auto refresh loop**

在 `progress/multi.go` 添加私有方法：

```go
func (mp *MultiProgress) startAutoRefreshLocked()
func (mp *MultiProgress) flushDirty()
func (mp *MultiProgress) stopAutoRefresh()
```

实现要求：

- `startAutoRefreshLocked()` 初始化 `stopCh` 和 `doneCh`；当 `RefreshInterval <= 0` 时使用 `DefaultRefreshInterval`；启动一个 ticker goroutine。
- `flushDirty()` 获取 `mp.mu`，没有 dirty 时跳过；有 dirty 时清理标记并调用 `refreshLocked()`。
- `stopAutoRefresh()` 关闭 `stopCh` 并等待 `doneCh`；等待时不能持有 `mp.mu`。

- [ ] **Step 9: 将 auto refresh 接入 `Start()` 和 `Finish()`**

`Start()` 中：

- 保持现有 bar 初始化逻辑。
- 设置 `started = true`。
- 首次调用 `refreshLocked()` 渲染初始状态。
- `AutoRefresh == true` 时启动 ticker。

`Finish()` 中：

- 已经 finished 时直接返回。
- 如果 `Finish()` 在 `Start()` 前被调用，先初始化 bars。
- 最终输出前停止 auto-refresh goroutine。
- 最终渲染一次。
- 有 bar 时输出结束换行。
- 设置 `finished = true`。

实现时需要注意锁顺序：等待 `doneCh` 时不能持有 `mp.mu`。

- [ ] **Step 10: 验证本任务测试**

运行：

```bash
go test ./progress -run "TestMultiProgress(RedrawFreq|AutoRefresh|FinishStops)" -count=1
```

预期：三个新增测试全部通过。

- [ ] **Step 11: 验证 progress 包现有测试**

运行：

```bash
go test ./progress
```

预期：`progress` 包全部测试通过。

- [ ] **Step 12: 提交 Task 1**

运行：

```bash
git add progress/multi.go progress/progress.go progress/multi_test.go
git commit -m "feat(progress): throttle managed multi progress refresh"
```

预期：如果当前会话允许提交，commit 成功；如果当前流程不提交，按协作者要求保留 staged 或 unstaged 状态。

---

### Task 2: Progress Reset、Setter 和生命周期 Getter

**Files:**
- Modify: `progress/progress.go`
- Test: `progress/progress_test.go`

- [ ] **Step 1: 添加 standalone `Reset()` 的失败测试**

在 `progress/progress_test.go` 添加：

```go
func TestProgressReset(t *testing.T) {
	is := assert.New(t)
	p := Txt(10)
	p.Out = new(bytes.Buffer)
	p.Start()
	p.AdvanceTo(7)
	p.Finish()

	p.Reset(20)
	is.Eq(int64(0), p.Step())
	is.Eq(float32(0), p.Percent())
	is.Eq(int64(20), p.Max())
	is.True(p.Started())
	is.False(p.Finished())
}
```

- [ ] **Step 2: 添加 managed `Reset()` 复用测试**

在 `progress/progress_test.go` 添加：

```go
func TestProgressResetManaged(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf

	p := mp.New(10)
	p.SetFormat("{@name} {@percent}%")
	p.SetMessage("name", "fd")
	mp.Start()
	p.AdvanceTo(10)
	p.Reset(20)
	p.SetMessage("name", "bat")
	p.AdvanceTo(5)
	mp.Finish()

	out := buf.String()
	is.Contains(out, "bat 25.0%")
	is.Eq(int64(5), p.Step())
	is.Eq(int64(20), p.Max())
}
```

- [ ] **Step 3: 添加 setter API 测试**

在 `progress/progress_test.go` 添加：

```go
func TestProgressSetters(t *testing.T) {
	is := assert.New(t)
	p := New()
	p.SetMaxSteps(8).
		SetFormat("{@name}:{@phase}:{@max}").
		SetMessage("name", "fd")
	p.SetMessages(map[string]string{"phase": "downloading"})
	p.Start()

	is.Eq("fd:downloading:8", p.Line())
}
```

- [ ] **Step 4: 运行测试并确认失败**

运行：

```bash
go test ./progress -run "TestProgress(Reset|Setters)" -count=1
```

预期：测试失败，因为新 API 还不存在。

- [ ] **Step 5: 实现 `Reset()` 和内部 `reset()`**

在 `progress/progress.go` 添加：

```go
func (p *Progress) Reset(maxSteps ...int64)
func (p *Progress) reset(maxSteps ...int64)
```

实现要求：

- 托管模式调用 `p.manager.update(func() bool { p.reset(maxSteps...); return true })`。
- standalone 模式直接调用 `p.reset(maxSteps...)`，如果已经 started，则显示当前行。
- `reset()` 设置 `step`、`percent`、`started`、`firstRun`、`startedAt`、`finishedAt` 和规范化后的 max steps。
- `reset()` 确保 `RedrawFreq` 默认值为 `1`。
- `reset()` 调用 `p.addWidgets(builtinWidgets)`。
- `reset()` 不清空已有 `Messages`。

- [ ] **Step 6: 实现 `ResetWith()`**

添加：

```go
func (p *Progress) ResetWith(fn func(p *Progress))
```

实现要求：

- 托管模式在同一个 manager update callback 内执行 `p.reset()` 和 `fn(p)`，然后规范化 `MaxSteps`。
- standalone 模式不经过 manager 锁，逻辑相同；如果已经 started，则显示当前行。
- `fn == nil` 时行为等同 `Reset()`。

- [ ] **Step 7: 实现 setter API**

添加方法：

```go
func (p *Progress) SetMaxSteps(maxSteps int64) *Progress
func (p *Progress) SetMessage(name, message string) *Progress
func (p *Progress) SetMessages(messages map[string]string) *Progress
func (p *Progress) SetFormat(format string) *Progress
```

实现要求：

- 托管模式在 `p.manager.update(...)` 内变更状态并返回 `true`。
- standalone 模式直接变更状态。
- `SetMessages` 使用现有 `addMessages` 合并 message。
- `SetMaxSteps` 使用 `normalizeMaxSteps`。

- [ ] **Step 8: 实现生命周期 getter**

添加：

```go
func (p *Progress) Started() bool
func (p *Progress) Finished() bool
func (p *Progress) Max() int64
```

实现要求：

- `Started()` 返回 `p.started`。
- `Finished()` 返回 `!p.finishedAt.IsZero()`。
- `Max()` 返回 `p.MaxSteps`。

- [ ] **Step 9: 验证本任务测试**

运行：

```bash
go test ./progress -run "TestProgress(Reset|Setters)" -count=1
```

预期：新增 reset 和 setter 测试通过。

- [ ] **Step 10: 验证 progress 包**

运行：

```bash
go test ./progress
```

预期：`progress` 包全部测试通过。

- [ ] **Step 11: 提交 Task 2**

运行：

```bash
git add progress/progress.go progress/progress_test.go
git commit -m "feat(progress): add reusable progress reset APIs"
```

预期：如果当前会话允许提交，commit 成功。

---

### Task 3: MultiProgress 安全日志和 Manager Getter

**Files:**
- Modify: `progress/multi.go`
- Test: `progress/multi_test.go`

- [ ] **Step 1: 添加 `RunExclusive()` / `Println()` 失败测试**

在 `progress/multi_test.go` 添加：

```go
func TestMultiProgressRunExclusivePrintsBetweenBlocks(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf

	p := mp.New(10)
	p.AddMessage("message", " task")
	mp.Start()
	p.AdvanceTo(5)
	mp.Println("warning: fallback")
	p.AdvanceTo(10)
	mp.Finish()

	out := buf.String()
	is.Contains(out, "warning: fallback")
	is.Contains(out, "100.0%(10/10)")
	is.True(strings.Count(out, "\x1B[2K") >= 3)
}
```

- [ ] **Step 2: 添加 `Printf()` 失败测试**

在 `progress/multi_test.go` 添加：

```go
func TestMultiProgressPrintf(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.New(1)
	mp.Start()

	mp.Printf("package %s failed: %s\n", "fd", "network")
	mp.Finish()

	is.Contains(buf.String(), "package fd failed: network")
}
```

- [ ] **Step 3: 添加 manager getter 失败测试**

在 `progress/multi_test.go` 添加：

```go
func TestMultiProgressStateGetters(t *testing.T) {
	is := assert.New(t)
	mp := NewMulti()
	mp.Writer = new(bytes.Buffer)
	mp.New(1)

	is.False(mp.Started())
	is.False(mp.Finished())
	is.Eq(1, mp.Len())
	is.Eq(1, mp.VisibleLen())

	mp.Start()
	is.True(mp.Started())
	mp.Finish()
	is.True(mp.Finished())
}
```

- [ ] **Step 4: 运行测试并确认失败**

运行：

```bash
go test ./progress -run "TestMultiProgress(RunExclusive|Printf|StateGetters)" -count=1
```

预期：测试失败，因为安全日志和 getter API 还不存在。

- [ ] **Step 5: 实现 `clearLocked()`**

在 `progress/multi.go` 添加：

```go
func (mp *MultiProgress) clearLocked()
```

实现要求：

- `!mp.rendered` 或 `mp.lastLines == 0` 时直接返回。
- 将光标移动到已渲染 block 的第一行。
- 对每个已渲染行输出 `\x1B[2K` 清除。
- 将光标恢复到清除区域起点。
- 设置 `mp.rendered = false`。

- [ ] **Step 6: 实现 `RunExclusive()`**

添加：

```go
func (mp *MultiProgress) RunExclusive(fn func(w io.Writer))
```

实现要求：

- 获取 `mp.mu`。
- 调用 `mp.clearLocked()`。
- `fn != nil` 时调用 `fn(mp.writer())`。
- 如果 `mp.started && !mp.finished`，调用 `mp.refreshLocked()`。
- 释放锁。

- [ ] **Step 7: 实现 `Println()` 和 `Printf()`**

添加：

```go
func (mp *MultiProgress) Println(args ...any)
func (mp *MultiProgress) Printf(format string, args ...any)
```

两个方法都通过 `RunExclusive()` 实现，分别使用 `fmt.Fprintln` 和 `fmt.Fprintf`。

- [ ] **Step 8: 实现 manager getter**

添加：

```go
func (mp *MultiProgress) Started() bool
func (mp *MultiProgress) Finished() bool
func (mp *MultiProgress) Len() int
func (mp *MultiProgress) VisibleLen() int
```

实现要求：

- 所有方法都获取 `mp.mu`。
- 第一阶段 `VisibleLen()` 返回 `len(mp.bars)`。

- [ ] **Step 9: 验证本任务测试**

运行：

```bash
go test ./progress -run "TestMultiProgress(RunExclusive|Printf|StateGetters)" -count=1
```

预期：Task 3 新增测试全部通过。

- [ ] **Step 10: 验证 progress 包**

运行：

```bash
go test ./progress
```

预期：`progress` 包全部测试通过。

- [ ] **Step 11: 提交 Task 3**

运行：

```bash
git add progress/multi.go progress/multi_test.go
git commit -m "feat(progress): add safe multi progress logging"
```

预期：如果当前会话允许提交，commit 成功。

---

### Task 4: 文档和完整验证

**Files:**
- Modify: `docs/zh-CN/progress.md`
- Modify: `docs/progress.md`
- Verify: `progress/multi.go`
- Verify: `progress/progress.go`
- Verify: `progress/multi_test.go`
- Verify: `progress/progress_test.go`

- [ ] **Step 1: 更新中文 progress 文档**

在 `docs/zh-CN/progress.md` 的 `Multi Progress` 部分补充：

- `AutoRefresh` / `RefreshInterval` 说明。
- 使用 `Reset()` 复用固定 worker slot 的示例。
- 使用 `Println()` 和 `Printf()` 安全输出日志的示例。
- 生命周期 getter 列表。

核心示例：

```go
mp := progress.NewMulti()
mp.AutoRefresh = true
mp.RefreshInterval = 100 * time.Millisecond
mp.Start()
defer mp.Finish()
```

slot 复用示例：

```go
bar := mp.New()
bar.SetFormat("{@slot} {@name} {@percent}% {@phase}")
bar.SetMessage("slot", "#1")
bar.Reset(100)
bar.SetMessages(map[string]string{"name": "fd", "phase": "downloading"})
```

- [ ] **Step 2: 更新英文 progress 文档**

在 `docs/progress.md` 添加对应英文说明和示例。

术语保持一致：

- auto refresh
- reusable worker slot
- exclusive logging
- lifecycle getters

- [ ] **Step 3: 运行 progress 测试**

运行：

```bash
go test ./progress
```

预期：`progress` 测试全部通过。

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
- `progress/multi_test.go`
- `progress/progress_test.go`
- `docs/zh-CN/progress.md`
- `docs/progress.md`
- `docs/superpowers/specs/2026-05-09-cliui-progress-enhancement-design.md`
- `docs/superpowers/plans/2026-05-09-cliui-progress-enhancement.md`

- [ ] **Step 6: 提交 Task 4**

运行：

```bash
git add docs/zh-CN/progress.md docs/progress.md
git commit -m "docs(progress): document multi progress enhancements"
```

预期：如果当前会话允许提交，commit 成功。

---

## 最终验收标准

- `go test ./progress` 通过。
- `go test ./...` 通过，或只存在已记录且与本次改动无关的失败。
- 托管模式下 `Progress.AdvanceTo()` 尊重 `RedrawFreq`。
- `MultiProgress.AutoRefresh` 能批量刷新，并在 `Finish()` 时干净停止。
- `Progress.Reset()` 能复用 managed bar，不需要重新 `Start()`。
- `SetMessage()` / `SetMessages()` / `SetFormat()` / `SetMaxSteps()` 在 standalone 和 managed 模式都可用。
- `RunExclusive()` / `Println()` / `Printf()` 能安全输出日志，不会永久破坏 progress block。
- 公开文档说明新增 API，并包含 batch download 风格示例。

## 实施注意事项

- 不实现本计划明确延期的能力。
- 不修改 `widgetMatch` 或 format token 行为。
- 不引入 `Slot` 类型。
- 不把 `AutoRefresh` 设为默认行为；默认仍然是同步刷新。
- 等待 auto-refresh goroutine 退出时不要持有 `mp.mu`。
- `Finish()` 必须保持幂等。
