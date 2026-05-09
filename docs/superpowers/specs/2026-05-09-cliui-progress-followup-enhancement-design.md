# cliui progress 后续增强阶段设计

## 背景

`progress` 第一阶段已经完成并发下载 UI 的核心能力：

- `MultiProgress.AutoRefresh` / `RefreshInterval` 支持后台节流刷新。
- 托管模式下 `Progress.RedrawFreq` 生效。
- `Progress.Reset()` / `ResetWith()` 支持固定 worker slot 复用。
- `SetMaxSteps()` / `SetMessage()` / `SetMessages()` / `SetFormat()` 支持托管状态安全更新。
- `MultiProgress.RunExclusive()` / `Println()` / `Printf()` 支持安全日志输出。
- `Progress` 和 `MultiProgress` 已有基础生命周期查询。

这些能力已经能支撑 `eget` 的第一版固定 worker slot 多包下载 UI。后续增强主要面向更完整的生产 CLI 场景：

- 非 TTY、CI、pipe、重定向输出不能使用 ANSI 动态多行刷新。
- package 名、asset 名过长时需要对齐和截断，避免撑破终端行。
- 动态任务列表需要隐藏、显示、移除 progress。
- 批量任务需要更直接的 done、failed、skipped 状态 helper。
- chunk 并发下载可能需要更轻量的 byte 聚合工具。

本文将后续能力拆成三个阶段，避免一次性扩展过大，也避免过早冻结不确定 API。

## 总体目标

第二阶段目标：

1. 增加 render mode，让调用方可以选择动态渲染、plain fallback 或完全禁用 progress 渲染。
2. 提供 TTY 检测工具，帮助调用方决定 render mode。
3. 增强 format token 支持左对齐和截断。

第三阶段目标：

1. 支持 `Hide()` / `Show()` / `Remove()` 动态管理 progress。
2. 让 `VisibleLen()` 真正反映参与渲染的 bar 数量。
3. 提供 `Done()` / `Fail()` / `Skip()` 状态 helper，减少调用方重复样板代码。

第四阶段目标：

1. 根据 `eget` 实际接入反馈，决定是否提供并发 byte tracker 或 writer。
2. 如果需要，则提供多 goroutine 安全的轻量聚合能力，降低调用方 chunk 进度聚合复杂度。

## 非目标

后续增强仍不引入通用 TUI 框架。

后续增强不把 `SpinnerFactory` 纳入 `MultiProgress`。

后续增强不引入 `Slot` 公开类型。固定 worker slot 已经可以通过 `Progress.Reset()` 实现。

后续增强不让 `progress` 包理解业务状态机。`resolving`、`downloading`、`verifying`、`extracting` 等业务状态仍由调用方通过 message 字段表达。

第二阶段不改变 unknown format token 的默认行为。未知 token 继续原样保留，避免破坏现有使用者。

第四阶段不在没有性能或易用性证据前实现复杂 tracker。第一阶段的 auto refresh 已经解决主要闪烁问题，byte tracker 需要由实际压测或接入反馈触发。

## 阶段拆分

### 第二阶段：Render Mode 和 Format 增强

第二阶段优先级最高。它解决的是 `eget` 作为 CLI 工具真正发布时必须面对的输出环境问题。

#### RenderMode

新增类型：

```go
type RenderMode int

const (
	RenderDynamic RenderMode = iota
	RenderPlain
	RenderDisabled
)
```

`MultiProgress` 新增字段：

```go
RenderMode RenderMode
```

默认值为 `RenderDynamic`，保持现有行为兼容。

#### RenderDynamic

`RenderDynamic` 使用当前 ANSI 光标控制方式渲染多行 progress block。

行为：

- `Start()` 渲染初始 block。
- `Refresh()` 立即重绘 block。
- `AutoRefresh` dirty flush 重绘 block。
- `RunExclusive()` 清除 block、输出日志、恢复 block。
- `Finish()` 最终渲染并换行。

这是现有默认行为。

#### RenderPlain

`RenderPlain` 不使用 ANSI 光标上移和清行，不维护动态多行 block。它适合 CI、普通 pipe、重定向文件等非交互环境。

关键设计原则：plain mode 不能在每次 `Advance()` 时输出一行，否则高频下载会污染日志。

建议 plain mode 只在关键状态变化时输出：

- `Start()`：输出当前已注册 progress 的初始行。
- `Reset()`：输出该 bar 的新任务行。
- `Finish()` / `Done()` / `Fail()` / `Skip()`：输出最终状态行。
- `RunExclusive()` / `Println()` / `Printf()`：正常输出日志。
- `Refresh()`：调用方显式要求时输出所有可见 bar 当前行。

`AutoRefresh` 在 `RenderPlain` 下不启动后台 ticker，或者 ticker 不产生输出。plain mode 的重点是事件输出，不是持续刷新。

`Progress.Advance()` 在 `RenderPlain` 下仍更新内部状态，但默认不输出。

#### RenderDisabled

`RenderDisabled` 完全禁用 progress 渲染，但不禁用安全日志 API。

行为：

- `Start()` 不输出 progress。
- `Refresh()` 不输出 progress。
- `AutoRefresh` 不启动后台 ticker。
- `Reset()` / `Advance()` / `Finish()` 只更新状态，不输出 progress。
- `RunExclusive()` / `Println()` / `Printf()` 仍写日志。

这适合 `quiet` 模式或调用方只想复用 progress 状态但不想显示 UI 的场景。

#### TTY 检测

新增工具函数：

```go
func IsTerminal(w io.Writer) bool
```

设计要求：

- 支持常见 `*os.File` writer。
- 对无法识别的 writer 返回 `false`。
- 不在 `NewMulti()` 中自动改变 mode，避免破坏兼容。

调用方推荐用法：

```go
mp := progress.NewMulti()
mp.Writer = stderr
if !progress.IsTerminal(stderr) {
	mp.RenderMode = progress.RenderPlain
}
```

后续如果需要，可以再提供便捷方法：

```go
func (mp *MultiProgress) UseAutoRenderMode()
```

但第二阶段不实现这个便捷方法。

#### Format token 增强

当前 token 正则只支持简单格式：

```go
var widgetMatch = regexp.MustCompile(`{@([\w_]+)(?::([\w-]+))?}`)
```

第二阶段改为：

```go
var widgetMatch = regexp.MustCompile(`{@([\w_]+)(?::([^}]+))?}`)
```

这样可以支持：

```text
{@name:-12s}
{@name:.20s}
{@percent:5s}
```

实现仍使用：

```go
fmt.Sprintf("%"+fmtArg, text)
```

兼容性要求：

- `{@percent:4s}` 继续工作。
- `{@estimated:-7s}` 继续工作。
- unknown token 继续原样返回。
- 不新增 missing token 默认空字符串行为。

如果后续确实需要 missing token 渲染为空，应该单独设计显式开关，而不是改变默认行为。

### 第三阶段：动态管理和状态 Helper

第三阶段用于支持更复杂的 batch UI。固定 worker slot 不依赖这些能力，但动态任务列表、完成折叠、失败保留等场景需要它们。

#### Hide / Show / Remove

新增 API：

```go
func (mp *MultiProgress) Hide(p *Progress)
func (mp *MultiProgress) Show(p *Progress)
func (mp *MultiProgress) Remove(p *Progress)
```

`Progress` 增加内部字段：

```go
hidden  bool
removed bool
```

语义：

- `Hide(p)`：bar 仍归 manager 管理，但不参与渲染。
- `Show(p)`：bar 重新参与渲染。
- `Remove(p)`：bar 从渲染集合移除，后续对这个 bar 的更新 no-op。

`Remove()` 后不 panic。原因是并发任务收尾时可能有迟到的 `Advance()`、`SetMessage()` 或 `Finish()`，panic 对 CLI 工具不友好。

`Remove()` 后也不应该退化成 standalone progress 输出，否则会破坏当前终端区域。

#### VisibleLen

第一阶段 `VisibleLen()` 与 `Len()` 相同。第三阶段后：

- `Len()` 返回 manager 中仍被跟踪的 bar 数量。
- `VisibleLen()` 返回未 hidden 且未 removed 的 bar 数量。

是否让 `Remove()` 物理删除 `bars` 中的元素，需要根据实现复杂度决定：

- 如果物理删除，需要维护 index。
- 如果逻辑删除，`Len()` 是否包含 removed bar 会产生语义争议。

推荐语义：

- `Remove()` 后从 `bars` slice 中移除。
- `p.removed = true` 保留在 progress 自身，用于后续 no-op。
- 对 remaining bars 重建 index。

这样 `Len()` 表示当前 manager 有效 bar 数量，`VisibleLen()` 表示当前参与渲染 bar 数量。

#### 状态 Helper

新增 API：

```go
func (p *Progress) Done(message ...string)
func (p *Progress) Fail(message ...string)
func (p *Progress) Skip(message ...string)
```

语义：

- 只更新 progress 行，不直接输出额外日志。
- 通过现有 manager update 路径刷新或标记 dirty。
- `Done()` 推进到 `MaxSteps`，设置完成时间。
- `Fail()` 设置完成时间，但不强制推进到 100%。
- `Skip()` 设置完成时间，但不强制推进到 100%。

默认 message 字段：

- `Done("done")` 设置 `message=done` 和 `status=done`。
- `Fail("failed")` 设置 `message=failed` 和 `status=failed`。
- `Skip("skipped")` 设置 `message=skipped` 和 `status=skipped`。

如果没有传入 message，使用默认文本：

- done
- failed
- skipped

这些 helper 不理解业务 phase。调用方仍可继续使用：

```go
bar.SetMessage("phase", "verifying")
bar.SetMessage("extra", "chunks:5")
```

#### Render mode 交互

第三阶段 API 需要兼容第二阶段 render mode：

- `RenderDynamic`：hide/show/remove 后刷新 block 或标记 dirty。
- `RenderPlain`：hide/show 默认不输出；remove 默认不输出；done/fail/skip 输出最终状态行。
- `RenderDisabled`：hide/show/remove/done/fail/skip 只更新状态，不输出 progress。

### 第四阶段：ByteTracker / ConcurrentWriter

第四阶段只在实际需要时实现。触发条件包括：

- `eget` 接入第一阶段后，chunk worker 高频调用 `Progress.Advance(n)` 导致明显锁竞争。
- 调用侧需要重复编写 byte 聚合逻辑，API 使用成本明显偏高。
- 压测显示通过聚合 flush 可以显著降低 CPU 或锁竞争。

#### API 草案

轻量 tracker：

```go
type ByteTracker struct {
	// internal
}

func NewByteTracker(p *Progress) *ByteTracker
func (t *ByteTracker) Add(n int64)
func (t *ByteTracker) Close()
```

writer 适配：

```go
func NewConcurrentWriter(p *Progress) io.Writer
```

#### 语义

- 多 goroutine 可安全调用 `Add()` 或 `Write()`。
- 内部聚合字节增量，按时间窗口 flush 到 `Progress.Advance(delta)`。
- `Close()` flush 剩余字节，并停止内部 goroutine。
- `Close()` 幂等。
- `Add(n <= 0)` no-op。

#### Flush 策略

默认 flush interval 可以是 `50ms` 或 `100ms`。如果 `Progress` 已由 `MultiProgress.AutoRefresh` 托管，tracker 主要减少锁竞争；如果 progress 是 standalone，tracker 还可以减少单条 progress 输出频率。

是否允许配置 interval，后续实现前再决定。第一版可以提供：

```go
func NewByteTrackerWithInterval(p *Progress, interval time.Duration) *ByteTracker
```

但如果没有明确需求，先不增加配置 API。

## 兼容性

默认行为必须保持兼容：

- `MultiProgress.RenderMode` 默认是 `RenderDynamic`。
- 未设置 render mode 的现有代码继续使用 ANSI 多行动态刷新。
- unknown format token 继续原样保留。
- `AutoRefresh` 默认仍为 `false`。
- `Progress` standalone 行为不因 render mode 改变；render mode 只属于 `MultiProgress`。

新增 API 都是 additive change。

## 测试计划

### 第二阶段测试

`progress/multi_test.go`：

- `RenderDynamic` 保持现有 ANSI 控制符行为。
- `RenderPlain` 不输出 ANSI 光标上移和清行。
- `RenderPlain` 下 `Advance()` 不刷屏，`Refresh()` 输出当前可见行。
- `RenderPlain` 下 `Reset()` 输出新任务行。
- `RenderDisabled` 下 `Start()` / `Refresh()` / `Finish()` 不输出 progress。
- `RenderDisabled` 下 `Println()` / `Printf()` 仍输出日志。

`progress/progress_test.go`：

- `{@name:-12s}` 支持左对齐。
- `{@name:.20s}` 支持截断。
- `{@percent:5s}` 保持工作。
- unknown token 保持原样。

TTY 检测测试：

- `bytes.Buffer` 返回 false。
- 普通文件返回 false。
- 如果测试环境可稳定构造 terminal file，再覆盖 true；否则 true 分支不强制单测。

### 第三阶段测试

`progress/multi_test.go`：

- `Hide()` 后渲染行数减少，`VisibleLen()` 变化。
- `Show()` 后重新参与渲染。
- `Remove()` 后 `Len()` 和 `VisibleLen()` 变化。
- `Remove()` 后 `Advance()` / `SetMessage()` / `Finish()` no-op，不写 standalone 输出。
- auto refresh 模式下 hide/show/remove 只标记 dirty。

`progress/progress_test.go`：

- `Done()` 推进到 max，设置 finishedAt，设置 message/status。
- `Fail()` 不强制推进到 max，设置 finishedAt，设置 message/status。
- `Skip()` 不强制推进到 max，设置 finishedAt，设置 message/status。

### 第四阶段测试

仅在实现时增加：

- 多 goroutine 并发 `Add()` 后总步数正确。
- `Close()` flush 剩余字节。
- `Close()` 幂等。
- `Write()` 返回写入长度。
- interval flush 不超过预期刷新次数。

## 推荐实施顺序

第二阶段：

1. 增加 `RenderMode` 类型和 `MultiProgress.RenderMode` 字段。
2. 将当前 `refreshLocked()` 明确作为 dynamic renderer。
3. 增加 plain/disabled 分支。
4. 增加 `IsTerminal()`。
5. 增强 format token 正则。
6. 更新中英文文档。

第三阶段：

1. 增加 hidden/removed 内部状态。
2. 实现 `Hide()` / `Show()` / `Remove()`。
3. 调整 `Len()` / `VisibleLen()`。
4. 让 update 路径识别 removed progress 并 no-op。
5. 实现 `Done()` / `Fail()` / `Skip()`。
6. 更新中英文文档。

第四阶段：

1. 先从 `eget` 接入和压测反馈确认是否需要。
2. 如果需要，先实现 `ByteTracker`。
3. 再用 `NewConcurrentWriter()` 作为 thin adapter 包装 tracker。

## 对 eget 的建议

`eget` 接入第一阶段时，可以先采用固定 worker slot：

- TTY 环境使用 `RenderDynamic`。
- CI 或 pipe 环境暂时由 `eget` 自己禁用 progress 或只输出摘要。
- 长 package name 暂时在 `eget` 侧截断。

第二阶段完成后：

- TTY 环境继续使用 `RenderDynamic`。
- 非 TTY 使用 `RenderPlain`。
- quiet 模式使用 `RenderDisabled`。
- package name / asset name 使用 `{@name:.20s}` 或类似格式截断。

第三阶段完成后：

- 如果 fixed slot 已足够，`eget` 不需要使用 hide/show/remove。
- 如果要实现“只显示活跃任务，完成项转摘要日志”，可以使用 `Hide()` 或 `Remove()`。
- 失败任务保留一行时，用 `Fail()` + `SetMessage("phase", "failed")`。

第四阶段完成后：

- 只有在 chunk 并发更新过于频繁或调用侧聚合复杂时，才切换到 `ByteTracker` 或 `NewConcurrentWriter()`。
