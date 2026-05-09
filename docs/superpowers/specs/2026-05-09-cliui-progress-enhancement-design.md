# cliui progress 增强设计

## 背景

`eget` 后续会同时支持两类并发下载：

- 单文件 `chunk_concurrency`：一个文件拆成多个 HTTP Range chunk 并发下载。
- 批量 `batch_concurrency`：`install --all` / `update --all` 同时处理多个 package。

这两类并发对终端进度展示的要求不同：

- chunk 并发应该对用户隐藏，只展示一个文件级聚合进度条。
- batch 并发应该展示固定数量的活跃 worker slot，每个 slot 一行，并且完成后可以复用给下一个 package。

当前 `progress.MultiProgress` 已经具备基础多行渲染能力，但仍存在几个限制：

- 托管模式下每次 `Progress.Advance()` 都会触发整块重绘，高频下载进度会导致刷新过密。
- `AutoRefresh` 和 `RefreshInterval` 字段已经存在，但还没有参与刷新策略。
- 托管模式下没有明确的 `Reset()` API，调用方难以安全复用固定 worker slot。
- 普通日志如果直接写同一个 writer，会破坏 `MultiProgress` 的多行 block。
- 缺少隐藏、移除、生命周期查询、TTY 策略等后续增强能力。

本文定义 `progress` 包的增强设计。目标是先完成能支撑 `eget` 稳定并发下载 UI 的核心能力，同时为后续更完整的多任务终端渲染保留扩展空间。

## 目标

首轮目标：

1. 让 `MultiProgress.AutoRefresh` 和 `RefreshInterval` 成为真实刷新策略。
2. 让托管模式下的 `Progress.RedrawFreq` 生效，避免无意义整块重绘。
3. 提供 `Progress.Reset()` 与常用 setter，使固定 worker slot 可以安全复用。
4. 提供 `MultiProgress.RunExclusive()` / `Println()` / `Printf()`，支持安全插入日志。
5. 补充基础生命周期查询，方便调用方判断进度条状态。

后续目标：

1. 支持 `Hide()` / `Show()` / `Remove()` 动态管理 progress。
2. 支持 render mode 或 TTY 检测，避免非交互式环境输出 ANSI 控制符。
3. 支持 done / fail / skip 等完成态 helper。
4. 支持并发 byte tracker 或 concurrent writer。
5. 增强 format token 的对齐、截断与缺省值行为。

## 非目标

首轮不引入通用 TUI 框架。

首轮不把 `SpinnerFactory` 纳入 `MultiProgress`。

首轮不新增复杂状态机。`progress` 包只负责渲染和基础生命周期，`resolving`、`downloading`、`verifying`、`extracting`、`failed` 等业务状态仍通过 message 字段表达。

首轮不强制做 TTY 自动降级。TTY 策略会单独设计，避免影响现有 ANSI 动态输出行为。

## 当前代码结构

相关文件：

- `progress/multi.go`：`MultiProgress` 管理器，多行 block 渲染，统一锁。
- `progress/progress.go`：`Progress` 状态、格式构建、单条输出、托管模式更新路径。
- `progress/widgets.go`：内置 widget，例如 `bar`、`percent`、`curSize`、`maxSize`。
- `progress/multi_test.go`：多进度条基础测试。
- `progress/progress_test.go`：单条 progress、IO wrapper、widget 测试。
- `docs/zh-CN/progress.md` 与 `docs/progress.md`：公开文档。

现有关键行为：

- `MultiProgress.update(fn)` 持有 manager 锁，执行状态更新后立即调用 `refreshLocked()`。
- `Progress.AdvanceTo()` 在托管模式下调用 `p.manager.update(func() { p.applyStep(step) })`。
- `Progress.applyStep(step)` 已经返回 `bool`，表示是否跨过 `RedrawFreq` 周期或到达最大进度。
- `Progress.Line()` 基于 `Format`、`Widgets` 和 `Messages` 返回当前渲染行。

因此首轮实现应该尽量复用已有结构，重点改造刷新策略和托管状态变更 API。

## 设计概览

`MultiProgress` 继续作为唯一终端写入者。所有托管 `Progress` 的状态变更都通过 manager 锁串行化。

新增刷新策略：

- 同步刷新模式：`AutoRefresh == false`。状态变更后，如果确实需要重绘，立即刷新多行 block。
- 自动刷新模式：`AutoRefresh == true`。状态变更只标记 dirty，由后台 ticker 按 `RefreshInterval` 执行刷新。
- 手动刷新：`Refresh()` 始终立即刷新，不依赖 dirty。
- 结束刷新：`Finish()` 停止后台 ticker，等待退出，最终刷新一次并换行。

新增 slot 复用能力：

- `Progress.Reset()` 重置一个 progress 的运行状态，但不解除它与 `MultiProgress` 的关系。
- `SetMessage()` / `SetMessages()` / `SetFormat()` / `SetMaxSteps()` 在托管模式下走 manager 锁，并触发刷新或标记 dirty。

新增安全日志能力：

- `RunExclusive()` 在 manager 锁内暂时清除当前 progress block，执行调用方输出，然后重新渲染 block。
- `Println()` / `Printf()` 是 `RunExclusive()` 的便捷封装。

## API 设计

### MultiProgress 自动刷新

现有字段继续保留：

```go
type MultiProgress struct {
	Overwrite       bool
	AutoRefresh     bool
	RefreshInterval time.Duration
	Writer          io.Writer
}
```

新增内部字段：

```go
type MultiProgress struct {
	mu       sync.Mutex
	bars     []*Progress
	started  bool
	finished bool
	rendered bool

	dirty  bool
	stopCh chan struct{}
	doneCh chan struct{}

	lastLines int
}
```

默认刷新间隔：

```go
const DefaultRefreshInterval = 100 * time.Millisecond
```

语义：

- `AutoRefresh = false` 时，保持同步刷新模式。
- `AutoRefresh = true` 时，`Start()` 启动后台刷新 goroutine。
- `RefreshInterval <= 0` 时使用 `DefaultRefreshInterval`。
- `Finish()` 关闭 `stopCh`，等待 `doneCh`，再执行最终刷新。
- `Finish()` 后不允许后台 goroutine 继续写 writer。
- `Refresh()` 可在任意模式下立即刷新。

### MultiProgress 状态更新

现有：

```go
func (mp *MultiProgress) update(fn func())
```

调整为内部方法：

```go
func (mp *MultiProgress) update(fn func() bool)
```

语义：

- `fn()` 返回 `false` 表示状态没有导致可见变化，不刷新也不标记 dirty。
- `fn()` 返回 `true` 表示需要刷新。
- `AutoRefresh = true` 时设置 `mp.dirty = true`。
- `AutoRefresh = false` 时立即调用 `refreshLocked()`。

托管模式下 `Progress.AdvanceTo()` 使用 `applyStep()` 的返回值：

```go
p.manager.update(func() bool {
	return p.applyStep(step)
})
```

这样 `RedrawFreq` 在 `MultiProgress` 下也会生效。

### Progress reset 和 setter

新增公开 API：

```go
func (p *Progress) Reset(maxSteps ...int64)
func (p *Progress) ResetWith(fn func(p *Progress))
func (p *Progress) SetMaxSteps(maxSteps int64) *Progress
func (p *Progress) SetMessage(name, message string) *Progress
func (p *Progress) SetMessages(messages map[string]string) *Progress
func (p *Progress) SetFormat(format string) *Progress
```

`Reset()` 重置：

- `step = 0`
- `percent = 0`
- `MaxSteps`
- `started = true`
- `firstRun = true`
- `startedAt = time.Now()`
- `finishedAt = time.Time{}`
- 默认 widgets

`Reset()` 不清空 `Messages`，因为固定 slot 通常会先设置固定字段，再覆盖任务字段。需要替换全部 messages 时使用 `SetMessages()`。

`ResetWith()` 用于需要一次性修改多个字段的场景：

```go
bar.ResetWith(func(p *progress.Progress) {
	p.MaxSteps = size
	p.Format = format
	p.Messages = map[string]string{
		"slot":  "#1",
		"name":  "bat",
		"phase": "downloading",
	}
})
```

托管模式下，`ResetWith()` 的回调在 manager 锁内执行。调用方不能在回调里调用会再次获取同一把锁的方法，例如 `SetMessage()` 或 `Advance()`。文档需要明确这一点。

### 安全日志输出

新增公开 API：

```go
func (mp *MultiProgress) RunExclusive(fn func(w io.Writer))
func (mp *MultiProgress) Println(args ...any)
func (mp *MultiProgress) Printf(format string, args ...any)
```

语义：

1. 获取 manager 锁。
2. 如果当前 block 已渲染，清除当前 block。
3. 使用 `mp.writer()` 执行日志输出。
4. 如果 manager 已启动且未结束，重新渲染当前 progress block。

`Println()` 等价于：

```go
mp.RunExclusive(func(w io.Writer) {
	fmt.Fprintln(w, args...)
})
```

`Printf()` 等价于：

```go
mp.RunExclusive(func(w io.Writer) {
	fmt.Fprintf(w, format, args...)
})
```

`RunExclusive()` 不负责日志等级和颜色。调用方可以在回调里使用 `fmt`、`color` 或 `ccolor`。

### 生命周期查询

新增 `Progress` 查询 API：

```go
func (p *Progress) Started() bool
func (p *Progress) Finished() bool
func (p *Progress) Max() int64
```

`Finished()` 基于 `finishedAt` 是否为零值判断，不用 `step == MaxSteps` 判断。这样可以兼容未知总大小、reset 复用、最大值增长等场景。

新增 `MultiProgress` 查询 API：

```go
func (mp *MultiProgress) Started() bool
func (mp *MultiProgress) Finished() bool
func (mp *MultiProgress) Len() int
func (mp *MultiProgress) VisibleLen() int
```

首轮没有隐藏/移除能力时，`VisibleLen()` 与 `Len()` 相同。后续实现 `Hide()` / `Remove()` 后，`VisibleLen()` 只统计参与渲染的 bar。

## 内部渲染设计

### Dirty 刷新循环

`Start()` 在 `AutoRefresh = true` 时启动后台 goroutine：

```go
func (mp *MultiProgress) startAutoRefreshLocked() {
	interval := mp.RefreshInterval
	if interval <= 0 {
		interval = DefaultRefreshInterval
	}

	mp.stopCh = make(chan struct{})
	mp.doneCh = make(chan struct{})

	go func() {
		defer close(mp.doneCh)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				mp.flushDirty()
			case <-mp.stopCh:
				return
			}
		}
	}()
}
```

`flushDirty()`：

```go
func (mp *MultiProgress) flushDirty() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if !mp.dirty {
		return
	}

	mp.dirty = false
	mp.refreshLocked()
}
```

`refreshLocked()` 只负责渲染，不负责判断 dirty。

### Finish 顺序

`Finish()` 需要避免持锁等待 goroutine，同时避免关闭 channel 后重复关闭。

建议流程：

1. 获取锁。
2. 如果已经 finished，释放锁并返回。
3. 标记 `finished = true`，取出 `stopCh` / `doneCh`。
4. 释放锁。
5. 如果存在 `stopCh`，关闭它并等待 `doneCh`。
6. 再次获取锁。
7. 临时允许最终渲染，调用 `refreshLocked()` 或专用 `renderFinalLocked()`。
8. 输出换行。
9. 释放锁。

也可以在关闭 goroutine 前先最终刷新，但必须确保 goroutine 不会在 `Finish()` 后再次刷新。实现时需要测试覆盖。

推荐实现方式是增加内部 `stopping` 或在关闭 goroutine 后最终刷新，避免并发写。

### 清除当前 block

安全日志需要清除已渲染 block。新增内部方法：

```go
func (mp *MultiProgress) clearLocked()
```

语义：

- 如果 `rendered == false` 或 `lastLines == 0`，直接返回。
- 光标回到 block 第一行。
- 逐行 `\x1B[2K` 清除。
- 光标回到清除区域起点。
- 设置 `rendered = false`。

`RunExclusive()` 通过 `clearLocked()` 给普通日志让出干净区域，然后重新 `refreshLocked()`。

## 固定 slot 用法

`eget` 侧目标用法：

```go
mp := progress.NewMulti()
mp.Writer = stderr
mp.AutoRefresh = true
mp.RefreshInterval = 100 * time.Millisecond
mp.Start()
defer mp.Finish()

slots := make([]*progress.Progress, batch)
for i := range slots {
	slots[i] = mp.New()
	slots[i].SetFormat("{@slot} {@name:-12s} [{@bar}] {@percent:5s}% {@curSize}/{@maxSize} {@phase} {@extra}")
	slots[i].SetMessage("slot", fmt.Sprintf("#%d", i+1))
}

slot := slots[workerID]
slot.Reset(size)
slot.SetMessages(map[string]string{
	"name":  pkgName,
	"phase": "downloading",
	"extra": fmt.Sprintf("chunks:%d", chunks),
})

_, err := io.Copy(dst, slot.WrapReader(src))
if err != nil {
	slot.SetMessage("phase", "failed")
	mp.Printf("Package %s failed: %v\n", pkgName, err)
	return err
}

slot.Finish("done")
```

chunk 并发时，多个 worker 可以共同更新同一个 file-level progress。首轮可以直接使用 `Progress.Advance(n)`，依赖 manager 锁和 `AutoRefresh` 节流。后续如果锁竞争明显，再增加 byte tracker。

## 后续扩展设计

### Hide / Show / Remove

后续新增：

```go
func (mp *MultiProgress) Hide(p *Progress)
func (mp *MultiProgress) Show(p *Progress)
func (mp *MultiProgress) Remove(p *Progress)
```

推荐在 `Progress` 增加内部字段：

```go
hidden  bool
removed bool
```

`refreshLocked()` 只渲染未隐藏、未移除的 bar。

`Remove()` 后继续调用 `bar.Advance()` 应该 no-op，不应该退化成单条 progress 输出。这样对并发任务收尾更安全。

### Render mode 和 TTY

后续新增：

```go
type RenderMode int

const (
	RenderDynamic RenderMode = iota
	RenderPlain
	RenderDisabled
)
```

并提供：

```go
func IsTerminal(w io.Writer) bool
```

首轮不实现 render mode，原因是 plain fallback 的输出频率和摘要语义需要单独定义，否则容易污染 CI 日志。

### Done / Fail / Skip helper

后续可以新增：

```go
func (p *Progress) Done(message ...string)
func (p *Progress) Fail(message ...string)
func (p *Progress) Skip(message ...string)
```

这些方法只更新 message 和完成状态，不直接输出额外日志。业务方仍通过 `phase`、`status` 等 message 字段控制展示。

### ConcurrentWriter / ByteTracker

后续可以新增：

```go
func NewConcurrentWriter(p *Progress) io.Writer
func NewByteTracker(p *Progress) *ByteTracker
```

首轮不做的原因是 `AutoRefresh` 已经解决主要闪烁问题。是否需要额外聚合锁竞争，应该根据 `eget` 实际压测决定。

### Format token 增强

当前 token 正则只支持 `{@name}` 和类似 `{@percent:4s}` 的简单格式。后续建议改为：

```go
var widgetMatch = regexp.MustCompile(`{@([\w_]+)(?::([^}]+))?}`)
```

这样可以支持：

- `{@name:-12s}` 左对齐。
- `{@name:.20s}` 截断。
- `{@percent:5s}` 保持兼容。

未知 token 是否渲染为空有兼容风险，后续需要单独决定。首轮不改变未知 token 行为。

## 兼容性

单条 `Progress` 行为保持兼容：

- `Start()` 重复调用仍然 panic。
- `Finish()` 仍然补齐进度并换行。
- `Out`、`Overwrite`、`Newline`、`Format`、`Widgets`、`Messages` 继续工作。

`MultiProgress` 行为保持兼容：

- `AutoRefresh = false` 是默认模式，仍然同步刷新。
- `Start()`、`Refresh()`、`Finish()` 的公开语义保持不变。
- 已有 `Add()` / `New()` 用法不变。

首轮新增 API 不破坏现有 API。

唯一行为修正是：托管模式下会尊重 `RedrawFreq`。这会减少输出次数，但不会改变最终状态。

## 测试计划

新增或调整 `progress/multi_test.go`：

- `TestMultiProgressRedrawFreqInManagedMode`：设置 `RedrawFreq`，连续 `Advance()`，验证不会每次都刷新。
- `TestMultiProgressAutoRefreshMarksDirty`：启用 `AutoRefresh`，连续更新后立即检查输出次数较少，等待 interval 后检查最终内容出现。
- `TestMultiProgressFinishStopsAutoRefresh`：`Finish()` 后等待超过 interval，确认 writer 没有继续增长。
- `TestMultiProgressRunExclusivePrintsBetweenBlocks`：启动多进度，调用 `Println()`，验证日志内容存在且后续 progress block 仍可刷新。
- `TestMultiProgressStateGetters`：验证 `Started()`、`Finished()`、`Len()`、`VisibleLen()`。

新增或调整 `progress/progress_test.go`：

- `TestProgressReset`：验证 step、percent、max、startedAt、finishedAt。
- `TestProgressResetManaged`：托管 progress reset 后仍由 manager 渲染，不直接写单条输出。
- `TestProgressSettersManaged`：`SetMessage()`、`SetMessages()`、`SetFormat()`、`SetMaxSteps()` 在托管模式下刷新。
- `TestProgressLifecycleGetters`：验证 `Started()`、`Finished()`、`Max()`。

后续扩展对应测试：

- hide/show/remove 的可见行数和 no-op 更新。
- render mode 的动态、plain、disabled 行为。
- format token 左对齐和截断。
- byte tracker 并发写入。

## 文档计划

更新：

- `docs/zh-CN/progress.md`
- `docs/progress.md`

新增内容：

- `MultiProgress.AutoRefresh` 使用示例。
- 固定 worker slot 复用示例。
- 安全日志输出示例。
- `Reset()` 和 setter API 说明。
- 生命周期 getter 说明。

示例重点放在 batch download 场景：

```go
mp := progress.NewMulti()
mp.AutoRefresh = true
mp.RefreshInterval = 100 * time.Millisecond
mp.Start()
defer mp.Finish()

bar := mp.New()
bar.SetFormat("{@slot} {@name:-12s} [{@bar}] {@percent:5s}% {@curSize}/{@maxSize} {@phase}")
bar.Reset(size)
bar.SetMessages(map[string]string{
	"slot":  "#1",
	"name":  "fd",
	"phase": "downloading",
})
```

## 实施顺序

建议分成三个小批次实现。

第一批：刷新策略

1. 给 `MultiProgress` 增加 dirty、stopCh、doneCh 和默认刷新间隔。
2. 改造 `update(fn func() bool)`。
3. 调整 `Progress.AdvanceTo()`，让 `applyStep()` 的返回值控制托管刷新。
4. 实现 auto refresh goroutine。
5. 补自动刷新和 `RedrawFreq` 测试。

第二批：slot 复用 API

1. 实现 `Progress.Reset()` 和内部 `reset()`。
2. 实现 `ResetWith()`。
3. 实现 `SetMaxSteps()`、`SetMessage()`、`SetMessages()`、`SetFormat()`。
4. 实现 `Started()`、`Finished()`、`Max()`。
5. 补 reset、setter、getter 测试。

第三批：安全日志

1. 实现 `clearLocked()`。
2. 实现 `RunExclusive()`。
3. 实现 `Println()` 和 `Printf()`。
4. 实现 `MultiProgress` getter。
5. 补安全日志和 manager 状态测试。

三个批次完成后，`eget` 可以实现稳定的固定 worker slot 多包下载 UI。其它扩展能力再按实际使用反馈继续推进。
