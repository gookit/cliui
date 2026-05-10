# cliui progress 增强需求

## 背景

`eget` 计划同时支持两类并发：

- `chunk_concurrency`：单个文件的 HTTP Range 分片并发下载。
- `batch_concurrency`：`install --all` / `update --all` 的多包任务并发。

这会带来新的终端进度展示需求：

```text
单包下载:
  一个文件可能由多个 chunk 并发下载，但用户只需要看到一个聚合进度条。

批量安装/更新:
  多个 package 可能同时下载，每个活跃 package 应该有独立进度行。
```

当前 `github.com/gookit/cliui/progress` 已有 `MultiProgress`，并且具备基础的多进度条渲染能力。但如果要让 `eget` 的并发下载体验稳定、低闪烁、可复用 worker slot、能安全输出日志，需要 `cliui/progress` 提供更完整的多任务进度 API。

本文描述的是理想能力，不是最小支持范围。

## 目标体验

### 单文件 chunk 并发

chunk 并发对用户隐藏为一个文件级进度条：

```text
fd       [===================>      ] 73.2% 8.1MiB/11.0MiB chunks:5
```

多个 chunk worker 同时下载时，进度条按聚合字节数刷新，不展示 chunk 级别子进度。

### batch 并发

`batch_concurrency > 1` 时，显示固定数量的 worker slot，而不是为所有 package 一次性创建进度条：

```text
#1 fd       [===================>      ] 73.2% 8.1MiB/11.0MiB downloading
#2 ripgrep  [=========>                ] 34.0% 4.4MiB/13.0MiB downloading
#3 jq       resolving release...
```

当 `fd` 完成后，slot `#1` 可以复用给下一个 package：

```text
#1 bat      [=====>                    ] 18.8% 2.2MiB/11.7MiB downloading
#2 ripgrep  [========================>] 100.0% 13.0MiB/13.0MiB done
#3 jq       extracting...
```

普通日志、warning、error 可以安全插入，不破坏多行进度块：

```text
Warning: sourceforge range download unsupported for winmerge, falling back to single connection

#1 bat      [=====>                    ] 18.8% 2.2MiB/11.7MiB downloading
#2 ripgrep  [========================>] 100.0% 13.0MiB/13.0MiB done
#3 jq       extracting...
```

## 需要的能力

## 1. MultiProgress 自动刷新与节流

### 问题

当前 managed `Progress.Advance()` 会触发 `MultiProgress.update()`，每次更新都可能重绘整个多行 block。

下载场景中，多个 goroutine 会频繁上报字节增量。如果每次 `Read()` 都导致整块重绘，会产生：

- 终端刷新过于频繁。
- CPU 和 IO 开销偏高。
- 多包多 chunk 下载时闪烁明显。

当前 `MultiProgress` 结构已有字段：

```go
AutoRefresh     bool
RefreshInterval time.Duration
```

但理想上需要让它们真正成为刷新策略的一部分。

### 期望 API

```go
mp := progress.NewMulti()
mp.AutoRefresh = true
mp.RefreshInterval = 100 * time.Millisecond
mp.Start()
defer mp.Finish()
```

语义：

- `AutoRefresh = true` 时，`Advance()` / `AddMessage()` / `Reset()` 只更新状态并标记 dirty。
- 后台 ticker 按 `RefreshInterval` 刷新。
- 如果没有 dirty 状态，不刷新。
- `Finish()` 停止 ticker，最终刷新一次并退出多行 block。

### 细节要求

- `RefreshInterval <= 0` 时使用默认值，例如 `100ms`。
- `Finish()` 必须等待刷新 goroutine 退出，避免结束后仍写终端。
- `Refresh()` 仍可手动触发立即刷新。
- `AutoRefresh = false` 时保留同步刷新行为，但要尊重 `Progress.RedrawFreq`。

### 补充要求：RedrawFreq 在 MultiProgress 下生效

即使不启用 `AutoRefresh`，managed `Progress` 也应该只在 `applyStep()` 判断需要刷新时才触发 `MultiProgress.refreshLocked()`。

期望内部语义类似：

```go
changed := p.applyStep(step)
if changed {
    mp.refreshLocked()
}
```

## 2. Progress 可安全重置和复用

### 问题

`eget install --all --batch 3` 最理想的 UI 是固定 3 个 worker slot。每个 slot 完成一个 package 后复用给下一个 package。

当前 `Progress` 没有明确的 reset API。直接修改公开字段如 `MaxSteps`、`Messages`、`step` 不安全，也不适合 managed progress。

### 期望 API

```go
bar.Reset(maxSteps int64)
bar.ResetWith(func(p *progress.Progress) {
    p.MaxSteps = size
    p.Format = format
    p.Messages = map[string]string{
        "slot": "#1",
        "name": "bat",
        "phase": "downloading",
    }
})
```

或者更简单：

```go
bar.Reset(size)
bar.SetMessage("slot", "#1")
bar.SetMessage("name", "bat")
bar.SetMessage("phase", "downloading")
bar.SetFormat(format)
```

### 语义

`Reset()` 应重置：

- `step = 0`
- `percent = 0`
- `MaxSteps`
- `startedAt = time.Now()`
- `finishedAt = zero`
- 可选清理/替换 messages

`Reset()` 不应该让 bar 离开 `MultiProgress`，也不应该要求重新 `Start()`。

在 manager 模式下，`Reset()` 必须通过 manager 的锁执行，并触发一次安全刷新或标记 dirty。

## 3. 安全日志输出

### 问题

`MultiProgress` 使用 ANSI 光标上移和清行控制多行 block。如果其他代码直接写同一个 writer，显示会被破坏。

`eget` 的 batch 任务需要输出：

- warning
- fallback notice
- selected asset
- checksum verified
- extracted result
- error detail

这些输出不能和多进度块抢 writer。

### 期望 API

```go
mp.Println("Warning: fallback to single connection")
mp.Printf("Package %s failed: %v\n", name, err)

mp.RunExclusive(func(w io.Writer) {
    fmt.Fprintf(w, "Checksum verified with %s\n", sum)
})
```

### 语义

调用安全日志 API 时：

1. 如果多进度块已渲染，先清除或移开当前 block。
2. 执行日志输出。
3. 重新渲染当前进度 block。

这组 API 必须和 progress update 共用同一把锁。

### 可选能力

支持日志输出等级或颜色不属于 progress 核心职责，可以只提供 writer 独占能力，让调用方使用 `fmt` / `ccolor`。

## 4. 动态显示、隐藏、移除 progress

### 问题

固定 worker slot 是最推荐的 UI，但动态任务场景仍然需要隐藏或移除已完成的 bar：

- 小批量时可以每个 package 一行，完成后保留或折叠。
- 大批量时只显示活跃任务，完成任务转为摘要日志。
- 失败任务可以保留一行错误状态，其它完成任务隐藏。

### 期望 API

```go
mp.Hide(bar)
mp.Show(bar)
mp.Remove(bar)
```

### 语义

- `Hide(bar)`：bar 仍在 manager 中，但不参与渲染。
- `Show(bar)`：重新参与渲染。
- `Remove(bar)`：从 manager 中移除，不再更新和渲染。

`Remove()` 后如果调用 `bar.Advance()`，建议返回 no-op 或明确 panic。更安全的是 no-op，并提供状态查询。

## 5. Progress 状态与生命周期查询

### 问题

调用方需要根据 bar 状态决定是否 reset、finish、hide、复用。

当前 `Progress` 有部分 getter，但缺少明确生命周期状态。

### 期望 API

```go
bar.Started() bool
bar.Finished() bool
bar.Step() int64
bar.Max() int64
bar.Percent() float32
```

已有的 `Step()` / `Percent()` 可以保留。建议补充 `Started()` / `Finished()` / `Max()`。

`MultiProgress` 也需要：

```go
mp.Started() bool
mp.Finished() bool
mp.Len() int
mp.VisibleLen() int
```

## 6. Writer 与 TTY 策略

### 问题

动态多行 progress 适合交互式终端，不适合 CI 日志、重定向文件、普通 pipe。

当前 `MultiProgress` 会无条件输出 ANSI 控制符。

### 期望 API

```go
mp.Dynamic = true
mp.FallbackMode = progress.FallbackPlain
```

或者提供检测工具：

```go
progress.IsTerminal(w io.Writer) bool
```

### 期望行为

当 writer 不是 TTY 时，调用方可以选择：

- 禁用动态刷新，只输出最终摘要。
- 使用 plain fallback，每个关键状态输出一行。
- 保留 ANSI 输出，由调用方显式选择。

理想上 `MultiProgress` 支持模式：

```go
type RenderMode int

const (
    RenderDynamic RenderMode = iota
    RenderPlain
    RenderDisabled
)
```

## 7. 固定 slot 模型支持

### 背景

`eget` 最理想的 batch UI 是固定 worker slot，而不是 package 数量等于 progress bar 数量。

### 期望 API

可以通过已有 `Reset()` 实现，也可以在 `cliui` 提供更直接的 slot API。

示例：

```go
mp := progress.NewMulti()
slot := mp.NewSlot("#1")

slot.Run("fd", size, func(p *progress.Progress) error {
    // download and advance p
    return nil
})

slot.Run("bat", nextSize, func(p *progress.Progress) error {
    // same slot reused
    return nil
})
```

如果不想引入 Slot 概念，`Progress.Reset()` 足够支撑 eget 侧实现 slot。

## 8. 聚合进度事件支持

### 背景

chunk 并发下载时，多个 worker 会上报字节增量。调用方可以自己聚合，但 `cliui/progress` 如果提供轻量工具，会更方便。

### 可选 API

```go
tracker := progress.NewByteTracker(bar)
tracker.Add(n)
tracker.Close()
```

或者：

```go
writer := progress.NewConcurrentWriter(bar)
```

语义：

- 多 goroutine 可安全调用。
- 内部按时间窗口聚合更新，避免每次 `Write()` 都刷新。

这项不是必须，但对下载类工具很有价值。

## 9. 完成态与失败态渲染

### 问题

批量任务中，package 可能处于：

- resolving
- downloading
- verifying
- extracting
- done
- failed
- skipped

`Progress.Finish(message...)` 可以设置 message，但没有明确状态模型。

### 期望能力

不一定需要 progress 包理解所有状态，但至少需要方便更新 message 和格式：

```go
bar.SetMessage("phase", "done")
bar.SetMessage("status", "OK")
bar.Finish("done")

bar.SetMessage("phase", "failed")
bar.SetMessage("status", "ERR")
bar.Finish("failed")
```

如果提供状态样式会更好：

```go
bar.Done("done")
bar.Fail("failed")
bar.Skip("skipped")
```

这些方法应只更新 progress 行，不直接打印额外日志。

## 10. 推荐格式能力

`eget` 需要在进度行中展示自定义字段：

```text
{@slot} {@name:-12s} [{@bar}] {@percent:5s}% {@curSize}/{@maxSize} {@phase} {@extra}
```

当前 format 已支持 message 和 widget，基本够用。希望增强：

- 支持左对齐格式，例如 `{@name:-12s}`。
- message 不存在时可以渲染为空，而不是保留原 token。
- 支持截断长文本，例如 `{@name:.20s}` 或内置 truncate widget。

长 package name / asset name 不截断时，多行 progress 很容易横向溢出。

## 建议优先级

如果按理想体验排序，建议优先实现：

1. `MultiProgress` 自动刷新与节流，且 `RedrawFreq` 在 manager 模式下生效。
2. `Progress.Reset()` / `SetMaxSteps()` / `SetMessage()` / `SetFormat()`。
3. `MultiProgress.RunExclusive()` / `Printf()` / `Println()` 安全日志输出。
4. `Hide()` / `Show()` / `Remove()` 动态管理 bar。
5. TTY 检测或 render mode。
6. 完成态、失败态 helper。
7. 并发 byte writer / tracker。
8. format 截断和更完整对齐能力。

前三项完成后，`eget` 就可以实现稳定的固定 worker slot 多包下载 UI。

## eget 侧目标用法

理想情况下，`eget` 侧代码可以接近这样：

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
}

slot := slots[workerID]
slot.Reset(size)
slot.SetMessages(map[string]string{
    "slot":  fmt.Sprintf("#%d", workerID+1),
    "name":  pkgName,
    "phase": "downloading",
    "extra": fmt.Sprintf("chunks:%d", chunks),
})

writer := progress.NewConcurrentWriter(slot)
err := downloadWithChunks(url, writer)
if err != nil {
    slot.SetMessage("phase", "failed")
    mp.Printf("Package %s failed: %v\n", pkgName, err)
    return err
}

slot.Finish("done")
```

这个 API 形态能让应用层专注调度和下载逻辑，`cliui/progress` 负责终端渲染一致性。

## 对 eget 当前设计的影响

如果 `cliui/progress` 增强到上述能力，`eget` 的并发下载实现可以更简单：

- chunk worker 只写并发安全的 byte tracker。
- batch worker 复用固定 progress slot。
- 普通日志通过 `MultiProgress.RunExclusive()` 输出。
- `quiet` 模式直接使用 disabled renderer。
- 非 TTY 环境使用 plain renderer 或完全关闭动态 progress。

如果 `cliui` 暂时不增强，`eget` 仍可先实现保守版本：

- chunk 层自行聚合进度事件。
- 小批量 batch 使用现有 `MultiProgress`。
- 大批量 batch 使用摘要输出。
- 普通日志缓冲到任务结束后输出。

但理想长期方案还是把刷新节流、slot 复用、安全日志和 TTY 策略放到 `cliui/progress`，这样其它 CLI 工具也能复用。
