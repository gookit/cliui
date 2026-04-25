# interact 抽象交互层设计稿

## 背景

当前 `interact` 子包的交互模型以“打印提示 + 读取一行输入”为主，典型实现包括：

- `ReadInput` / `ReadLine`
- `Ask` / `Question`
- `Confirm`
- `SelectOne` / `MultiSelect`

这套实现简单直接，适合普通问答式终端交互，但不适合以下场景：

- 基于方向键的选项切换
- 多选框实时勾选
- 输入中的即时校验与重绘
- 后续接入 `readline` 一类的行编辑或 raw terminal 能力

因此需要在不破坏现有 `interact` API 的前提下，引入一套新的交互抽象层，为未来的 `readline`/TUI 化组件提供稳定边界。

## 目标

第一期目标：

- 新增一套独立于现有 `interact` API 的交互抽象层
- 第一批覆盖四类组件：`Input`、`Confirm`、`Select`、`MultiSelect`
- 先定义稳定接口与状态模型，不绑定具体底层库
- 为未来接入 `readline`、plain 终端回退实现、fake backend 测试实现预留扩展点

非目标：

- 本期不直接替换现有 `interact.SelectOne`、`interact.Ask`、`interact.Confirm`
- 本期不绑定具体实现库，例如 `github.com/chzyer/readline`
- 本期不设计完整通用 widget tree 或完整 TUI 框架
- 本期不解决所有终端 UI 场景，例如复杂布局、分页表格、鼠标事件

## 现状问题

当前代码存在以下结构性限制：

### 1. 交互流程是阻塞式的一次性读取

现有 `ReadLine` 与 `Prompt` 读取的是整行输入，组件拿不到逐键事件，因此无法支持：

- 上下方向键移动
- 空格切换勾选
- Esc / Ctrl+C 统一取消
- 输入中实时过滤

### 2. 组件逻辑与 IO 紧耦合

现有 `Select` / `Question` 会直接渲染内容并读取输入，导致：

- 状态难以复用
- 渲染与状态更新无法解耦
- 无法在不同 backend 间复用组件逻辑

### 3. 错误处理偏命令式

现有实现中多处通过 `os.Exit` 直接退出，这不适合作为一套可组合的底层交互库。新抽象层应统一返回错误值，而不是结束进程。

### 4. 依赖全局 `Input` / `Output`

现有 `interact` 使用包级别全局输入输出流。该做法对简单交互足够，但对多 backend、测试替身、会话级资源管理并不理想。

## 设计原则

### 1. 新旧并存，先隔离再整合

第一期采用“新增子包”的方式建设新抽象层，现有 `interact` 代码保持兼容，不做侵入式改造。

### 2. 模型与渲染分离

组件只维护自己的状态与行为规则，不直接依赖具体终端渲染实现。

### 3. backend 可替换

抽象层只依赖统一的会话接口，后续可接：

- `plain` backend
- `readline` backend
- `fake` backend

### 4. 错误返回优先

取消、退出、终端不可用、校验失败等情况都返回 `error`，而不是直接 `os.Exit`。

### 5. 先覆盖高频组件

第一期仅覆盖：

- `Input`
- `Confirm`
- `Select`
- `MultiSelect`

不引入超前的通用组件系统。

## 推荐包结构

推荐先增加如下目录结构：

```text
interact/
  ui/
    types.go
    errors.go
    input.go
    confirm.go
    select.go
    multiselect.go
  backend/
    backend.go
    event.go
    view.go
```

说明：

- `interact/ui` 负责面向使用方的组件 API 与状态模型
- `interact/backend` 负责抽象底层会话、事件、渲染能力
- `backend/plain`、`backend/readline` 等具体实现不纳入本期范围，但结构上预留位置

如果后续希望减少包层级，也可以把 `backend` 收敛到 `interact/ui/backend`，但不影响本设计核心。

## 核心抽象

### Backend

`Backend` 表示一个终端交互能力提供者，负责创建会话。

```go
type Backend interface {
	NewSession(in io.Reader, out io.Writer) (Session, error)
}
```

### Session

`Session` 表示一次完整的终端交互会话，负责：

- 初始化和释放终端状态
- 读取事件
- 刷新视图
- 提供终端尺寸等运行时信息

```go
type Session interface {
	Render(view View) error
	ReadEvent(ctx context.Context) (Event, error)
	Size() (width, height int)
	Close() error
}
```

### Event

`Event` 统一描述用户输入事件。第一期只需要覆盖键盘事件和中断事件，不做鼠标或复杂输入法处理。

```go
type EventType int

const (
	EventUnknown EventType = iota
	EventKey
	EventInterrupt
	EventResize
)

type Key string

const (
	KeyEnter     Key = "enter"
	KeyUp        Key = "up"
	KeyDown      Key = "down"
	KeyLeft      Key = "left"
	KeyRight     Key = "right"
	KeySpace     Key = "space"
	KeyEsc       Key = "esc"
	KeyBackspace Key = "backspace"
	KeyCtrlC     Key = "ctrl+c"
)

type Event struct {
	Type EventType
	Key  Key
	Text string
}
```

### View

`View` 表示一次可渲染的终端输出快照。第一期不设计复杂布局树，最小化为“多行文本 + 光标位置”即可。

```go
type View struct {
	Lines        []string
	CursorRow    int
	CursorColumn int
}
```

该定义足以支持：

- 输入框
- 单选列表
- 多选列表
- 错误提示
- 当前项高亮

## 组件模型

### 1. Input

职责：

- 接收一段文本输入
- 支持默认值
- 支持校验函数
- 支持取消

建议接口：

```go
type Input struct {
	Prompt   string
	Default  string
	Validate func(string) error
}

func (c *Input) Run(ctx context.Context, be Backend) (string, error)
```

状态要点：

- 当前输入内容
- 光标位置
- 校验错误消息
- 是否已提交

### 2. Confirm

职责：

- 提供 yes/no 二选一
- 支持默认值
- 支持方向键切换或快捷键确认

建议接口：

```go
type Confirm struct {
	Prompt  string
	Default bool
}

func (c *Confirm) Run(ctx context.Context, be Backend) (bool, error)
```

### 3. Select

职责：

- 从候选项中选择一个值
- 支持默认选中项
- 支持方向键移动
- 支持回车确认
- 可选支持简单过滤输入

建议接口：

```go
type Item struct {
	Key     string
	Label   string
	Value   any
	Disabled bool
}

type Select struct {
	Prompt       string
	Items        []Item
	DefaultKey   string
}

func (c *Select) Run(ctx context.Context, be Backend) (*Result, error)
```

返回值建议复用统一结果结构：

```go
type Result struct {
	Key    string
	Keys   []string
	Value  any
	Values []any
}
```

### 4. MultiSelect

职责：

- 从候选项中选择多个值
- 支持默认选中项
- 支持空格勾选，回车确认
- 支持最少选择数量限制

建议接口：

```go
type MultiSelect struct {
	Prompt       string
	Items        []Item
	DefaultKeys  []string
	MinSelected  int
}

func (c *MultiSelect) Run(ctx context.Context, be Backend) (*Result, error)
```

## 组件行为约定

第一期建议统一以下交互规则：

### 通用规则

- `Ctrl+C` / `Esc`：取消当前交互，返回 `ErrAborted`
- `Enter`：提交
- `Resize`：触发重绘
- `ctx.Done()`：立即退出并返回上下文错误

### Input

- 普通字符：插入
- `Backspace`：删除前一个字符
- `Enter`：执行校验，通过则返回

### Confirm

- `Left` / `Right` 或 `y` / `n`：切换选项
- `Enter`：确认当前值

### Select

- `Up` / `Down`：移动高亮项
- `Enter`：选择当前项
- 如果启用过滤，普通字符写入 filter，列表实时收缩

### MultiSelect

- `Up` / `Down`：移动高亮项
- `Space`：切换当前项选中状态
- `Enter`：若满足 `MinSelected` 则提交，否则显示错误

## 错误模型

统一定义抽象层错误：

```go
var (
	ErrAborted      = errors.New("interact: aborted")
	ErrNoTTY        = errors.New("interact: tty required")
	ErrInvalidState = errors.New("interact: invalid state")
)
```

规则：

- 用户主动取消返回 `ErrAborted`
- 后端要求 TTY 但当前环境不满足时返回 `ErrNoTTY`
- 组件配置错误或内部状态异常返回普通错误或 `ErrInvalidState`

不允许在新抽象层内部直接 `os.Exit`。

## 与现有 `interact` 的关系

第一期明确采用“并存策略”：

- 现有 `interact.ReadLine`、`Ask`、`Confirm`、`SelectOne`、`MultiSelect` 保持原样
- 新交互层以新包提供，不替换旧实现
- 待新方案稳定后，再评估是否提供桥接层

不建议本期直接改造成：

- 让旧 `Select.Run()` 自动切换新 backend
- 让旧 `Ask()` 在内部依赖新组件

原因：

- 当前旧实现大量基于简单 IO 与 `os.Exit`
- 强行桥接会把兼容逻辑、交互状态机和终端驱动耦合在一起
- 第一版会明显增大风险和维护成本

## plain / readline 后续实现建议

虽然本期只设计抽象层，但需要给后续实现留出清晰路径。

### plain backend

目标：

- 非 TTY 环境仍可工作
- 行为尽量回退到当前 `interact` 的风格

特点：

- 不支持逐键事件时，可降级为整行输入
- `Select` / `MultiSelect` 可退回编号输入模式

### readline backend

目标：

- 提供事件驱动的逐键读取
- 支持方向键、空格、Esc、Ctrl+C、回删、光标移动等基础能力

特点：

- 负责 raw mode 和资源清理
- 负责把底层键盘输入映射到统一 `Event`
- 不把底层库类型泄露到 `ui` 抽象层

## 测试设计

本期设计应保证后续实现可以进行可重复测试。

建议测试策略：

- `ui` 层通过 fake session/feed event 的方式测试状态机
- backend 层单独测试事件转换与资源释放
- 组件行为测试覆盖：
  - 默认值生效
  - 校验失败后的错误展示
  - 中断返回 `ErrAborted`
  - `MultiSelect` 最少选择数限制
  - `Select` 光标移动边界

测试重点不应依赖真实终端，而应优先验证：

- 事件流是否驱动正确状态转移
- 视图快照是否符合预期
- 返回值和错误是否稳定

## 迁移与演进计划

建议分三步推进：

### 阶段 1：抽象层落地

- 建立 `ui` 与 `backend` 接口
- 固化错误模型、事件模型、视图模型
- 不绑定具体库

### 阶段 2：plain backend 与 fake backend

- 先实现一个简单可运行的 plain backend
- 配套 fake backend，验证接口是否合理

### 阶段 3：readline backend

- 在抽象层稳定后再接具体 readline 实现
- 增加输入过滤、快捷键、帮助提示等增强能力

## 备选方案对比

### 方案 A：新增独立子包，先做抽象层

优点：

- 风险低
- 兼容性最好
- 结构清晰
- 便于后续接多种 backend

缺点：

- 短期内会存在两套交互 API

### 方案 B：直接重构现有 `interact` API 的内部实现

优点：

- 最终用户感知最小

缺点：

- 与现有简单 IO 模式差异过大
- 第一阶段实现成本高
- 容易在兼容层上积累技术债

### 方案 C：直接引入完整 TUI/widget 框架

优点：

- 能力上限高

缺点：

- 明显超出当前项目需求
- 学习与维护成本过高

结论：选择方案 A。

## 最终结论

`interact` 的 readline 化演进，第一期应当优先建设“独立子包 + 抽象层先行”的结构，而不是直接改写现有 API。

最终推荐结论如下：

- 新增 `interact/ui` 与 `interact/backend` 两层抽象
- 第一批只覆盖 `Input`、`Confirm`、`Select`、`MultiSelect`
- 统一使用 `Session + Event + View` 作为 backend 边界
- 新抽象层只返回 `result/error`，不直接退出进程
- 本期不绑定具体 readline 库
- 后续按 `plain`、`fake`、`readline` 三类实现逐步推进

这套设计可以在不破坏现有 `interact` 的前提下，为后续更强的终端交互能力提供稳定基础。
