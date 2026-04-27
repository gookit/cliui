# interact/ui 事件 backend、resize 与选择过滤设计

## 背景

`interact/ui` 已经提供 `Input`、`Confirm`、`Select`、`MultiSelect` 四类组件，并通过 `backend.Session` 抽象出事件读取、视图渲染、终端尺寸和会话关闭能力。

当前抽象已经包含 `EventResize` 和 `Size()`，但组件尚未真正处理 resize，`readline` backend 也还不会主动发出 resize 事件。`Select` 和 `MultiSelect` 目前只能通过方向键移动或 item key 选择，缺少真实终端中常见的逐键过滤能力。

本设计是 `interact/ui` 第二阶段增强，目标是在不推翻现有抽象的前提下补齐事件驱动能力、resize 处理和选择过滤。

## 目标

- 保持现有 `backend.Backend` 和 `backend.Session` 主接口兼容。
- 扩展事件和视图模型，使 resize 和终端尺寸能被组件感知。
- 让 `readline` backend 能发出 resize 事件。
- 让 `fake` backend 支持固定尺寸和 resize 测试。
- 为 `Select` 和 `MultiSelect` 增加可选的过滤能力。
- resize 时重新计算可见列表，不中断交互、不清空选择状态。
- 更新中英文文档，说明不同 backend 下的行为差异。

## 非目标

- 不引入完整 TUI 框架。
- 不支持鼠标事件、alternate screen 或复杂布局树。
- 不重写旧 `interact` 包 API。
- 不让 `plain` backend 模拟 raw terminal。
- 不支持模糊匹配、正则匹配或匹配字符高亮。

## 推荐方案

采用渐进扩展方案：

- `Backend` / `Session` 接口保持不变。
- `Event` 增加 `Width` 和 `Height` 字段，用于 `EventResize`。
- `View` 增加 `Width`、`Height` 和 `HideCursor` 字段，供 backend 和测试使用。
- `Select` / `MultiSelect` 增加 `Filterable`、`FilterPrompt` 和 `PageSize` 配置。
- `fake` backend 增加固定尺寸配置，并在读取 `EventResize` 时更新尺寸。
- `readline` backend 在会话内维护 resize 通知，跨平台优先使用尺寸变化检测；Unix 可叠加 `SIGWINCH`。
- `plain` backend 保持行输入降级行为。

该方案复用当前抽象边界，兼容性最好，改动集中，适合当前项目规模。

## API 变更

### backend.Event

```go
type Event struct {
	Type   EventType
	Key    Key
	Text   string
	Width  int
	Height int
}
```

`Width` 和 `Height` 仅在 `EventResize` 时有意义。若 resize 事件没有携带尺寸，组件调用 `session.Size()` 获取当前尺寸。

### backend.View

```go
type View struct {
	Lines        []string
	CursorRow    int
	CursorColumn int
	Width        int
	Height       int
	HideCursor   bool
}
```

`Width` 和 `Height` 表示组件渲染时参考的尺寸。`HideCursor` 用于选择列表等不需要编辑光标的视图；backend 可以忽略该字段，但 `readline` backend 应处理它。

### ui.Select

```go
type Select struct {
	Prompt       string
	Items        []Item
	DefaultKey   string
	Filterable   bool
	FilterPrompt string
	PageSize      int
}
```

### ui.MultiSelect

```go
type MultiSelect struct {
	Prompt       string
	Items        []Item
	DefaultKeys  []string
	MinSelected  int
	Filterable   bool
	FilterPrompt string
	PageSize      int
}
```

兼容规则：

- `Filterable=false` 时保持当前行为。
- `PageSize=0` 时根据终端高度自动计算；无尺寸时使用默认页大小。
- `FilterPrompt=""` 时显示默认文案 `Filter`。
- `plain` backend 下仍按整行 item key 输入，不提供逐键过滤。

## 过滤行为

过滤只影响候选项可见性，不改变 `Items` 原始顺序。

匹配字段：

- `Item.Key`
- `Item.Label`
- `fmt.Sprint(Item.Value)`，仅当 `Value != nil`

匹配规则：

- case-insensitive contains
- query 为空时显示所有 item
- 禁用项可显示，但不能提交或勾选

按键规则：

- 普通文本：追加到 filter query。
- `Backspace`：删除 query 最后一个 rune。
- `Ctrl+U`：清空 query。
- `Up` / `Down`：在过滤后的可见项中移动。
- `PageUp` / `PageDown`：在过滤后的可见项中翻页或跳转。
- `Enter`：提交当前过滤结果中的高亮项。
- `Esc` / `Ctrl+C`：取消整个组件并返回 `ErrAborted`。

空结果行为：

- 显示 `No matches` 状态行。
- `Enter` 不提交，显示 `no matched option`。
- `MultiSelect` 已选项不因过滤消失而丢失。

## 分页和 resize 行为

组件每次渲染前根据终端尺寸计算可见列表：

- `availableHeight = terminalHeight - fixedLines`
- `PageSize > 0` 时使用配置值，但不超过可用高度。
- `PageSize == 0` 时使用可用高度。
- 无终端尺寸时默认页大小为 10。
- 最小 page size 为 1。

resize 事件处理：

- backend 发出 `EventResize{Width, Height}`。
- 组件更新内部 width/height。
- 重新计算 page size 和可见窗口。
- 不修改 filter query。
- 不修改 `MultiSelect` 已选项。
- 如果当前 cursor 指向的 item 仍在过滤结果中，保持该 item。
- 如果当前 item 不再可见，移动到第一个可用项。
- resize 只触发重绘，不提交、不取消、不清空错误。

## Backend 行为

### fake

- 支持固定尺寸配置，例如 `NewWithOptions(events, WithSize(width, height int))`。
- `Size()` 返回当前 fake session 尺寸。
- `ReadEvent()` 读到 `EventResize` 后更新当前尺寸。
- 继续记录所有 `View` 快照，供过滤和分页断言使用。

### plain

- 保持当前行输入行为。
- `Size()` 继续返回 `0, 0`。
- 不主动发出 `EventResize`。
- `Select` 接收 item key，空行使用默认 key。
- `MultiSelect` 接收逗号分隔 item key，空行使用默认 key 列表。
- 过滤配置不会改变 plain backend 的输入模型。

### readline

- 保持 raw terminal 输入。
- `ReadEvent(ctx)` 同时响应键盘输入、ctx 取消和 resize 通知。
- 会话内维护事件 channel 和关闭信号，避免每次读取都泄漏 goroutine。
- 跨平台优先使用尺寸变化检测；Unix 可用 `SIGWINCH` 更及时地唤醒 resize 检测。
- `Close()` 负责恢复终端状态、停止 resize 监听、恢复光标显示。

## 内部结构

新增小型 helper 文件，避免 `select.go` 和 `multiselect.go` 继续膨胀：

- `interact/ui/filter.go`：filter query 编辑、匹配和过滤结果。
- `interact/ui/list_view.go`：根据 cursor、过滤结果、尺寸和 page size 计算可见窗口。
- `interact/ui/terminal_size.go`：resize event 尺寸解析、默认页大小和可用高度计算。

组件状态机仍由各组件自己维护：

- `Select` 负责单选提交、禁用项校验、默认值。
- `MultiSelect` 负责 selected map、`MinSelected` 和结果组装。
- helper 不包含业务提交规则。

## 错误处理

沿用现有 exported error：

- `ErrAborted`
- `ErrNoTTY`
- `ErrInvalidState`

新增 UI 状态文案不导出为 error：

- `no matched option`
- `selected option is disabled`
- `select at least N option(s)`

resize 不返回错误。只有 backend 读取失败、渲染失败或 context 取消才返回 error。

## 测试策略

组件测试优先使用 fake backend：

- `Select` 启用过滤后输入 `be`，只显示 Beta，Enter 返回 `b`。
- `Select` 过滤无结果，Enter 显示 `no matched option`，Backspace 后恢复选择。
- `MultiSelect` 过滤后勾选项，清空过滤后已选项仍存在。
- 禁用项在过滤结果中仍不可提交。
- resize event 改变 page size，但不改变当前选择和 filter query。
- 小高度窗口不 panic，至少显示一个可选项和状态行。
- `Filterable=false` 的既有测试保持通过。

backend 测试：

- `fake.NewWithOptions(nil, fake.WithSize(...))` 返回指定尺寸。
- fake 收到 `EventResize` 后更新尺寸。
- readline 的 resize 检测用内部可测试函数覆盖，不依赖真实终端信号。
- `go test ./...` 必须通过。

## 文档更新

更新以下文档：

- `docs/zh-CN/interact-ui.md`
- `docs/interact-ui.md`

补充内容：

- `Select` / `MultiSelect` 的 `Filterable`、`PageSize` 示例。
- resize 行为说明。
- backend 行为差异：plain 不实时过滤，readline 支持逐键过滤和 resize，fake 用于测试事件流。

## 兼容性

本设计保持现有调用方式兼容：

- 现有 `ui.NewSelect(...).Run(...)` 无需改动。
- 现有 fake 事件构造无需改动。
- 现有 plain backend 行为不变。
- 新字段均为零值关闭增强行为。

## 实施顺序

1. 扩展 `backend.Event` 和 `backend.View`，并增强 fake backend 尺寸能力。
2. 增加过滤、分页和尺寸 helper。
3. 为 `Select` 增加过滤、分页和 resize 处理。
4. 为 `MultiSelect` 增加过滤、分页和 resize 处理。
5. 增强 readline backend 的 resize 事件路径和光标隐藏处理。
6. 更新中英文文档。
7. 运行 `go test ./...`。
