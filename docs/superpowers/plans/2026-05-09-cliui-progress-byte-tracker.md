# cliui progress ByteTracker / ConcurrentWriter 开发计划

> **给 agentic workers 的要求：** 按任务执行本计划时，必须使用 `superpowers:subagent-driven-development`（推荐）或 `superpowers:executing-plans`。任务使用 checkbox（`- [ ]`）格式跟踪进度。多阶段执行时，每个 Task 完成后必须更新 checkbox 并提交 Git commit。

**目标：** 实现 `progress` 第四阶段增强：为并发下载等高频字节更新场景提供 `ByteTracker` 和 `NewConcurrentWriter()`，减少调用方重复聚合逻辑，并降低高频 `Advance(n)` 带来的锁竞争和输出刷新压力。

**架构：** 新增一个独立的 byte tracking helper 文件，`ByteTracker` 负责并发安全地聚合字节增量，并按固定时间窗口 flush 到 `Progress.Advance(delta)`；`NewConcurrentWriter()` 只是 `ByteTracker` 的 `io.Writer` 适配器。`Progress`、`MultiProgress` 的现有渲染模型不改变，tracker 只通过公开的 `Progress.Advance()` 更新进度。

**技术栈：** Go，标准库 `io` / `sync` / `time`，现有 `progress` 包测试和中英文文档。

---

## 范围

本计划实现 `docs/superpowers/specs/2026-05-09-cliui-progress-followup-enhancement-design.md` 中第四阶段的轻量版本。

包含：

- `ByteTracker` 类型。
- `NewByteTracker(p *Progress) *ByteTracker`。
- `NewByteTrackerWithInterval(p *Progress, interval time.Duration) *ByteTracker`。
- `(*ByteTracker).Add(n int64)`。
- `(*ByteTracker).Close()`。
- `NewConcurrentWriter(p *Progress) io.Writer`。
- 多 goroutine 并发 `Add()` / `Write()` 的测试。
- 中英文文档更新。

暂不包含：

- `Slot` 类型或 `NewSlot()`。
- missing token 默认渲染为空。
- `SpinnerFactory` 纳入 `MultiProgress`。
- 复杂 backpressure、队列长度配置或动态 interval 调整。
- 对 `Progress` 或 `MultiProgress` 渲染行为的破坏性改变。

## 文件职责

- 新建：`progress/byte_tracker.go`
  - 定义 `ByteTracker`。
  - 定义默认 flush interval。
  - 实现并发安全聚合、ticker flush、幂等 `Close()`。
  - 实现 `NewConcurrentWriter()` 返回的 writer adapter。

- 修改：`progress/progress_test.go`
  - 增加 `ByteTracker` 和 `NewConcurrentWriter()` 的行为测试。
  - 覆盖并发 Add、Close flush、Close 幂等、Write 返回长度、interval flush。

- 修改：`docs/zh-CN/progress.md`
  - 在 IO Progress 或 Multi Progress 相关章节补充 byte tracker / concurrent writer 用法。

- 修改：`docs/progress.md`
  - 增加对应英文说明。

---

### Task 1: ByteTracker 基础 API 和 Close flush

**Files:**
- Create: `progress/byte_tracker.go`
- Modify: `progress/progress_test.go`

- [x] **Step 1: 添加 `NewByteTracker` 编译失败测试**

在 `progress/progress_test.go` 添加：

```go
func TestByteTrackerCloseFlushesPendingBytes(t *testing.T) {
	is := assert.New(t)
	p := New(10)
	p.Out = new(bytes.Buffer)
	p.Start()

	tracker := NewByteTrackerWithInterval(p, time.Hour)
	tracker.Add(3)
	is.Eq(int64(0), p.Step())

	tracker.Close()

	is.Eq(int64(3), p.Step())
}
```

说明：使用很长 interval，确保不是 ticker 自动 flush，而是 `Close()` flush 剩余字节。

- [x] **Step 2: 添加 `Add(n <= 0)` no-op 测试**

在 `progress/progress_test.go` 添加：

```go
func TestByteTrackerAddIgnoresNonPositiveValues(t *testing.T) {
	is := assert.New(t)
	p := New(10)
	p.Out = new(bytes.Buffer)
	p.Start()

	tracker := NewByteTrackerWithInterval(p, time.Hour)
	tracker.Add(0)
	tracker.Add(-3)
	tracker.Close()

	is.Eq(int64(0), p.Step())
}
```

- [x] **Step 3: 运行测试并确认失败**

运行：

```bash
go test ./progress -run "TestByteTracker(CloseFlushesPendingBytes|AddIgnoresNonPositiveValues)" -count=1
```

预期：编译失败，因为 `NewByteTrackerWithInterval` 和 `ByteTracker` 还不存在。

- [x] **Step 4: 新建 `progress/byte_tracker.go`**

创建文件并实现最小 API：

```go
package progress

import (
	"io"
	"sync"
	"time"
)

const DefaultByteTrackerInterval = 100 * time.Millisecond

type ByteTracker struct {
	progress *Progress
	interval time.Duration

	mu      sync.Mutex
	pending int64
	closed  bool
	stopCh  chan struct{}
	doneCh  chan struct{}
}

func NewByteTracker(p *Progress) *ByteTracker {
	return NewByteTrackerWithInterval(p, DefaultByteTrackerInterval)
}

func NewByteTrackerWithInterval(p *Progress, interval time.Duration) *ByteTracker {
	if interval <= 0 {
		interval = DefaultByteTrackerInterval
	}

	t := &ByteTracker{
		progress: p,
		interval: interval,
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}
	go t.loop()
	return t
}

func (t *ByteTracker) Add(n int64) {
	if n <= 0 || t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return
	}
	t.pending += n
}

func (t *ByteTracker) Close() {
	if t == nil {
		return
	}

	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return
	}
	t.closed = true
	close(t.stopCh)
	t.mu.Unlock()

	<-t.doneCh
	t.flush()
}

func (t *ByteTracker) loop() {
	defer close(t.doneCh)

	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.flush()
		case <-t.stopCh:
			return
		}
	}
}

func (t *ByteTracker) flush() {
	t.mu.Lock()
	n := t.pending
	t.pending = 0
	p := t.progress
	t.mu.Unlock()

	if n > 0 && p != nil {
		p.Advance(n)
	}
}

func NewConcurrentWriter(p *Progress) io.Writer {
	return byteTrackerWriter{tracker: NewByteTracker(p)}
}

type byteTrackerWriter struct {
	tracker *ByteTracker
}

func (w byteTrackerWriter) Write(bs []byte) (int, error) {
	n := len(bs)
	if n > 0 {
		w.tracker.Add(int64(n))
	}
	return n, nil
}
```

实现注意：

- `Close()` 必须幂等；Task 2 会补幂等测试，如果 Step 4 的实现还不够，需要在 Task 2 修正。
- `flush()` 不要持有 `ByteTracker.mu` 调用 `Progress.Advance()`，避免锁顺序耦合。
- `progress == nil` 时 `Add()` / `Close()` 不 panic，但不推进任何进度。

- [x] **Step 5: 验证 Task 1**

运行：

```bash
go test ./progress -run "TestByteTracker(CloseFlushesPendingBytes|AddIgnoresNonPositiveValues)" -count=1
go test ./progress
```

预期：新增测试和 `progress` 包测试通过。

- [x] **Step 6: 提交 Task 1**

运行：

```bash
git add progress/byte_tracker.go progress/progress_test.go
git commit -m "feat(progress): add byte tracker"
```

---

### Task 2: 并发 Add、幂等 Close 和 interval flush

**Files:**
- Modify: `progress/byte_tracker.go`
- Modify: `progress/progress_test.go`

- [x] **Step 1: 添加并发 Add 测试**

在 `progress/progress_test.go` 添加：

```go
func TestByteTrackerConcurrentAdd(t *testing.T) {
	is := assert.New(t)
	p := New(1000)
	p.Out = new(bytes.Buffer)
	p.Start()

	tracker := NewByteTrackerWithInterval(p, time.Hour)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				tracker.Add(1)
			}
		}()
	}
	wg.Wait()
	tracker.Close()

	is.Eq(int64(1000), p.Step())
}
```

同时给 `progress/progress_test.go` 增加 `sync` import。

- [x] **Step 2: 添加 Close 幂等测试**

在 `progress/progress_test.go` 添加：

```go
func TestByteTrackerCloseIsIdempotent(t *testing.T) {
	is := assert.New(t)
	p := New(10)
	p.Out = new(bytes.Buffer)
	p.Start()

	tracker := NewByteTrackerWithInterval(p, time.Hour)
	tracker.Add(2)
	tracker.Close()
	tracker.Close()

	is.Eq(int64(2), p.Step())
}
```

- [x] **Step 3: 添加 interval flush 测试**

在 `progress/progress_test.go` 添加：

```go
func TestByteTrackerFlushesOnInterval(t *testing.T) {
	is := assert.New(t)
	p := New(10)
	p.Out = new(bytes.Buffer)
	p.Start()

	tracker := NewByteTrackerWithInterval(p, 10*time.Millisecond)
	defer tracker.Close()

	tracker.Add(4)
	time.Sleep(40 * time.Millisecond)

	is.Eq(int64(4), p.Step())
}
```

- [x] **Step 4: 运行测试并确认失败或发现缺陷**

运行：

```bash
go test ./progress -run "TestByteTracker(ConcurrentAdd|CloseIsIdempotent|FlushesOnInterval)" -count=1
```

预期：如果 Task 1 的最小实现未正确处理幂等 Close，`TestByteTrackerCloseIsIdempotent` 可能失败或 deadlock；如果已经通过，继续执行 Step 5 做清理。

- [x] **Step 5: 修正 Close 幂等和锁顺序**

确保 `Close()` 满足：

- 第一次调用关闭 `stopCh`，等待 `doneCh`，flush 剩余字节。
- 后续调用直接返回，不重复 close channel，不重复 flush。
- `Close()` 不持有 `ByteTracker.mu` 等待 `doneCh`。

推荐实现：

```go
func (t *ByteTracker) Close() {
	if t == nil {
		return
	}

	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return
	}
	t.closed = true
	close(t.stopCh)
	doneCh := t.doneCh
	t.mu.Unlock()

	<-doneCh
	t.flush()
}
```

如果需要保证“Close 后 Add no-op 且不会丢掉 Close 前已经进入 pending 的数据”，`Add()` 必须在同一把 `mu` 下检查 `closed` 并累加 pending。

- [x] **Step 6: 验证 Task 2**

运行：

```bash
go test ./progress -run "TestByteTracker" -count=1
go test ./progress
```

预期：所有 ByteTracker 测试和 `progress` 包测试通过。

- [x] **Step 7: 提交 Task 2**

运行：

```bash
git add progress/byte_tracker.go progress/progress_test.go
git commit -m "test(progress): cover byte tracker concurrency"
```

---

### Task 3: NewConcurrentWriter 和 io.Copy 场景

**Files:**
- Modify: `progress/byte_tracker.go`
- Modify: `progress/progress_test.go`

- [x] **Step 1: 添加 Write 返回长度测试**

在 `progress/progress_test.go` 添加：

```go
func TestConcurrentWriterWriteReturnsLength(t *testing.T) {
	is := assert.New(t)
	p := New(10)
	p.Out = new(bytes.Buffer)
	p.Start()

	writer := NewConcurrentWriter(p)
	n, err := writer.Write([]byte("hello"))

	is.NoErr(err)
	is.Eq(5, n)
}
```

- [x] **Step 2: 添加并发 Write 测试**

在 `progress/progress_test.go` 添加：

```go
func TestConcurrentWriterConcurrentWrite(t *testing.T) {
	is := assert.New(t)
	p := New(1000)
	p.Out = new(bytes.Buffer)
	p.Start()

	writer := NewConcurrentWriterWithInterval(p, time.Hour)
	closer := writer.(io.Closer)

	var wg sync.WaitGroup
	errCh := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				n, err := writer.Write([]byte("x"))
				if err != nil {
					errCh <- err
					return
				}
				if n != 1 {
					errCh <- fmt.Errorf("unexpected write length: %d", n)
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		is.NoErr(err)
	}
	is.NoErr(closer.Close())

	is.Eq(int64(1000), p.Step())
}
```

同时给 `progress/progress_test.go` 增加 `fmt` import。说明：为了让测试不等待默认 interval，Task 3 推荐同时提供未在第四阶段草案中强制要求的测试辅助/公开 API：

```go
func NewConcurrentWriterWithInterval(p *Progress, interval time.Duration) io.WriteCloser
```

如果不希望公开该 API，则测试可以通过类型断言到内部类型并调用 `Close()`；但公开 `io.WriteCloser` 版本对调用方更实用。

- [x] **Step 3: 添加 io.Copy 测试**

在 `progress/progress_test.go` 添加：

```go
func TestConcurrentWriterWorksWithCopy(t *testing.T) {
	is := assert.New(t)
	p := New(11)
	p.Out = new(bytes.Buffer)
	p.Start()

	writer := NewConcurrentWriterWithInterval(p, time.Hour)
	n, err := io.Copy(writer, strings.NewReader("hello world"))
	is.NoErr(err)
	is.Eq(int64(11), n)
	is.NoErr(writer.Close())

	is.Eq(int64(11), p.Step())
}
```

- [x] **Step 4: 运行测试并确认失败**

运行：

```bash
go test ./progress -run "TestConcurrentWriter" -count=1
```

预期：如果 Task 1 只提供了 `NewConcurrentWriter(p) io.Writer`，新增的 `NewConcurrentWriterWithInterval` 和 `Close()` 能力尚不存在，测试失败。

- [x] **Step 5: 实现 closeable writer**

将 writer adapter 改为：

```go
func NewConcurrentWriter(p *Progress) io.Writer {
	return NewConcurrentWriterWithInterval(p, DefaultByteTrackerInterval)
}

func NewConcurrentWriterWithInterval(p *Progress, interval time.Duration) io.WriteCloser {
	return byteTrackerWriter{tracker: NewByteTrackerWithInterval(p, interval)}
}

type byteTrackerWriter struct {
	tracker *ByteTracker
}

func (w byteTrackerWriter) Write(bs []byte) (int, error) {
	n := len(bs)
	if n > 0 {
		w.tracker.Add(int64(n))
	}
	return n, nil
}

func (w byteTrackerWriter) Close() error {
	w.tracker.Close()
	return nil
}
```

兼容性说明：

- `NewConcurrentWriter(p)` 返回 `io.Writer`，满足原草案。
- `NewConcurrentWriterWithInterval` 返回 `io.WriteCloser`，便于调用方显式 flush。
- 如果调用方只用 `io.Writer` 且不 Close，后台 ticker 仍会按 interval flush，但最后不足一个 interval 的 pending bytes 可能需要等待 ticker；文档必须建议在下载完成后 Close closeable writer 或直接使用 `ByteTracker`。

- [x] **Step 6: 验证 Task 3**

运行：

```bash
go test ./progress -run "TestConcurrentWriter" -count=1
go test ./progress
```

预期：新增 writer 测试和 `progress` 包测试通过。

- [x] **Step 7: 提交 Task 3**

运行：

```bash
git add progress/byte_tracker.go progress/progress_test.go
git commit -m "feat(progress): add concurrent progress writer"
```

---

### Task 4: MultiProgress 集成回归和文档

**Files:**
- Modify: `progress/progress_test.go`
- Modify: `docs/zh-CN/progress.md`
- Modify: `docs/progress.md`

- [ ] **Step 1: 添加 managed AutoRefresh 集成测试**

在 `progress/progress_test.go` 添加：

```go
func TestByteTrackerWithManagedAutoRefreshProgress(t *testing.T) {
	is := assert.New(t)
	buf := new(bytes.Buffer)
	mp := NewMulti()
	mp.Writer = buf
	mp.AutoRefresh = true
	mp.RefreshInterval = 10 * time.Millisecond

	p := mp.New(12)
	mp.Start()

	tracker := NewByteTrackerWithInterval(p, time.Hour)
	tracker.Add(5)
	tracker.Add(7)
	tracker.Close()
	mp.Finish()

	is.Eq(int64(12), p.Step())
	is.Contains(buf.String(), "100.0%(12/12)")
}
```

该测试确认 tracker 和 managed progress 通过现有 `Progress.Advance()` 集成，不直接依赖 `MultiProgress` 内部锁。

- [ ] **Step 2: 更新中文文档**

在 `docs/zh-CN/progress.md` 的 `IO Progress` 部分补充：

- `ByteTracker` 用于多 goroutine 上报字节增量。
- `NewConcurrentWriter()` 用于 download writer 适配。
- 建议下载结束后调用 `Close()` flush 剩余字节。

示例：

```go
tracker := progress.NewByteTracker(bar)
defer tracker.Close()

go func() {
	tracker.Add(n)
}()
```

writer 示例：

```go
writer := progress.NewConcurrentWriterWithInterval(bar, 100*time.Millisecond)
defer writer.Close()

_, err := io.Copy(writer, resp.Body)
```

- [ ] **Step 3: 更新英文文档**

在 `docs/progress.md` 添加对应英文说明。术语保持一致：

- byte tracker
- concurrent writer
- flush interval
- close to flush pending bytes

- [ ] **Step 4: 验证 Task 4**

运行：

```bash
go test ./progress -run "TestByteTracker|TestConcurrentWriter" -count=1
go test ./progress
go test ./...
```

预期：新增 tests、`progress` 包和全仓库测试通过。

- [ ] **Step 5: 检查 diff 范围**

运行：

```bash
git diff --stat
git diff --name-only
```

预期变更文件：

- `progress/byte_tracker.go`
- `progress/progress_test.go`
- `docs/zh-CN/progress.md`
- `docs/progress.md`
- `docs/superpowers/plans/2026-05-09-cliui-progress-byte-tracker.md`

- [ ] **Step 6: 提交 Task 4**

运行：

```bash
git add progress/progress_test.go docs/zh-CN/progress.md docs/progress.md
git commit -m "docs(progress): document byte tracker usage"
```

---

## 最终验收标准

- `go test ./progress` 通过。
- `go test ./...` 通过，或只存在已记录且与本次改动无关的失败。
- `ByteTracker.Add()` 可被多个 goroutine 并发调用。
- `Add(n <= 0)` no-op。
- `Close()` flush 剩余字节。
- `Close()` 幂等，不 panic，不重复推进 progress。
- ticker interval 能自动 flush pending bytes。
- `NewConcurrentWriter()` 返回可用于 `io.Copy` / download writer 的 writer。
- closeable writer 能显式 flush 最后不足一个 interval 的 pending bytes。
- 不改变 `Progress.Write()` / `WrapReader()` / `WrapWriter()` 现有行为。
- 不改变 `MultiProgress` render mode、auto refresh 或 plain/disabled 语义。
- 中英文文档包含 byte tracker 和 concurrent writer 用法。

## 实施注意事项

- 严格 TDD：每个 Task 先写失败测试，再实现。
- 不在 `ByteTracker.mu` 内调用 `Progress.Advance()`。
- `Close()` 不能持锁等待 goroutine 退出。
- 不引入全局 goroutine 池或复杂调度器。
- 不用 `ByteTracker` 替换现有 `Progress.Write()` 行为；这是新增可选 API。
- 如果执行环境支持 race 检测，可额外运行：

```bash
racedetector test ./progress -run "TestByteTracker|TestConcurrentWriter" -count=1
```

当前本机 `racedetector v0.8.5` 需要 Go 1.24+；如果 Go 版本不足，记录为环境限制，不阻塞常规测试。
