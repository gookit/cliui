# progress 多进度条设计稿

## 背景

当前 `progress` 子包已经提供了单条进度显示能力，包括：

- 文本进度
- 条形进度条
- loading 风格进度
- spinner

现有实现适合“一次只显示一个进度实例”的场景，但不适合多个任务并行执行、需要同时展示多条进度的场景，例如：

- 并行下载多个文件
- 同时执行多个子任务并展示各自进度
- 主任务下的多个阶段并发推进

因此需要在保持当前单进度条 API 兼容的前提下，为 `progress` 引入一套多进度条统一渲染机制。

## 目标

第一期目标：

- 支持由一个统一管理器同时渲染多个 `Progress`
- 保持现有单条 `Progress` 的使用方式和行为兼容
- 避免多个 `Progress` 实例直接并发写终端导致的覆盖和闪烁
- 为后续扩展到 spinner 或混合条目类型预留结构

非目标：

- 本期不将 `SpinnerFactory` 纳入同一个统一调度器
- 本期不引入完整终端布局系统
- 本期不支持任意嵌套区域、分页、鼠标交互等高级 TUI 能力
- 本期不改变现有 `Progress` 的格式系统和 widget 机制

## 现状问题

当前 `Progress` 的渲染模型是典型的“单行覆盖式输出”：

- 每个 `Progress` 都默认假设自己独占当前终端输出位置
- 刷新逻辑依赖 `\r` 和清行控制序列
- `SpinnerFactory` 也采用同样的单行重绘方式

这会导致以下问题：

### 1. 多实例并发输出会互相覆盖

如果多个 goroutine 中的 `Progress` 同时调用 `Display()` 或 `Advance()`，它们会竞争同一个 stdout，最终表现为：

- 行内容交错
- 光标位置混乱
- 视觉闪烁
- 最终输出不可读

### 2. 当前 `Progress` 没有区域概念

单个 `Progress` 只能重绘“当前行”，没有能力：

- 记录自己位于第几行
- 回到多行区域顶部进行整体刷新
- 控制多行块的生命周期

### 3. 渲染职责分散在各实例内部

当前每个 `Progress` 实例既负责：

- 维护状态
- 计算展示内容
- 直接输出到终端

这使得多实例场景下很难统一调度输出。

## 设计原则

### 1. 统一渲染，单点写终端

多进度条模式下，只允许一个管理器负责终端写入。各个 `Progress` 实例只维护状态并产出当前行文本，不直接各写各的。

### 2. 兼容优先

现有 `Progress` 的单条模式保持不变。用户如果没有显式使用多进度条管理器，不应感知行为变化。

### 3. 最小改动引入多条模式

第一期仅围绕 `Progress` 增加“被统一托管渲染”的能力，不扩展为通用终端 UI 框架。

### 4. 结构上预留未来扩展

虽然第一期不支持 `SpinnerFactory`，但多条渲染机制应当为未来纳入其他可渲染条目类型保留空间。

## 设计范围

第一期范围限定为：

- 多个 `Progress` 实例
- 由一个 `MultiProgress` 管理器统一创建或统一接管
- 统一刷新多行区域
- 支持进度更新、结束、整体收尾

明确不包含：

- `SpinnerFactory`
- `Progress` 与 `SpinnerFactory` 混合显示
- 动态插入复杂区域分组
- 多终端/多 writer 并发同步

## 核心方案

第一期推荐引入一个新的统一管理器：`MultiProgress`。

职责如下：

- 管理多个 `Progress` 条目
- 统一处理终端输出
- 负责多行区域的首次初始化、重绘、结束收尾
- 串行化所有刷新动作

核心思路：

1. `Progress` 继续维护原有状态和格式逻辑
2. 将“渲染字符串”与“写终端”两个职责拆开
3. 单条模式下，`Progress` 仍然可以直接写终端
4. 多条模式下，`Progress` 不直接写终端，而是由 `MultiProgress` 统一取各条当前文本并重绘整块区域

## 推荐 API

### MultiProgress

建议新增如下类型：

```go
type MultiProgress struct {
	Overwrite bool
	AutoRefresh bool
	RefreshInterval time.Duration
}
```

建议提供以下构造和方法：

```go
func NewMulti() *MultiProgress

func (mp *MultiProgress) Add(p *Progress) *Progress
func (mp *MultiProgress) New(maxSteps ...int) *Progress

func (mp *MultiProgress) Start()
func (mp *MultiProgress) Refresh()
func (mp *MultiProgress) Finish()
```

说明：

- `Add(p *Progress)` 用于接管已存在的 `Progress`
- `New(maxSteps ...int)` 用于直接创建归属于当前管理器的 `Progress`
- `Start()` 初始化多行区域
- `Refresh()` 主动重绘全部条目
- `Finish()` 结束整个多进度块

### Progress 与 MultiProgress 的关系

每个 `Progress` 在多条模式下，应增加一个“托管关系”概念，例如：

```go
type Progress struct {
	// existing fields ...
	manager *MultiProgress
	index   int
}
```

含义：

- `manager == nil`：单条模式，沿用当前行为
- `manager != nil`：托管模式，不直接写终端，由 manager 统一刷新

## Progress 需要的最小改动

为了支撑多条模式，`Progress` 需要做如下最小改动。

### 1. 暴露当前渲染文本

目前 `Progress` 内部已经有 `buildLine()` 逻辑。第一期建议增加一个公开或受控公开的方法，例如：

```go
func (p *Progress) Line() string
```

作用：

- 返回当前进度条对应的一行完整文本
- 供 `MultiProgress` 重绘时使用

这一步非常关键。多条模式的本质不是让每条自己写，而是让 manager 收集所有行文本后统一输出。

### 2. 引入托管模式下的刷新分流

当前 `AdvanceTo()` 在达到刷新条件时会直接调用 `Display()`。多条模式下需要改成：

- 若未托管，则保持当前行为
- 若已托管，则通知 manager 进行刷新或标记 dirty

可选方式：

```go
func (p *Progress) requestRefresh()
```

内部逻辑：

- `manager == nil` 时走 `Display()`
- `manager != nil` 时走 `manager.Refresh()` 或向 manager 发刷新信号

### 3. Finish 行为要区分单条与多条

当前单条 `Finish()` 会：

- 补齐到最大进度
- 可能输出消息
- 最后打印换行

多条模式下不应由每个 `Progress` 自己额外换行，否则会破坏多行区域。建议改成：

- 单条模式：保持现有逻辑
- 多条模式：只更新状态，不自行打印额外换行
- 最终由 `MultiProgress.Finish()` 决定如何结束整个区域

## MultiProgress 渲染模型

### 区域模型

`MultiProgress` 接管一块连续终端区域，每个 `Progress` 对应其中一行。

示意：

```text
task A  [======>-----]  52%
task B  [========>---]  73%
task C  [===>--------]  26%
```

一次刷新流程如下：

1. 首次启动时，输出 N 行占位区域
2. 后续刷新时，将光标移动回区域顶部
3. 逐行清除并重绘全部条目
4. 刷新完成后，光标停留在区域底部或最后一行

### 为什么要整块重绘

推荐整块重绘，而不是只刷新某一行，原因如下：

- 简化终端控制逻辑
- 避免跟踪每一行的宽度与残留字符
- 更容易保证跨平台行为一致

第一期优先采用“整块重绘全部条目”的方案，必要时以后再优化为脏行刷新。

### 输出控制

`MultiProgress` 需要内部串行化刷新，防止多个 `Progress` 并发更新时同时写终端。

建议增加互斥锁：

```go
type MultiProgress struct {
	mu sync.Mutex
	// ...
}
```

所有 `Start()`、`Refresh()`、`Finish()`、条目更新触发的刷新都经过同一个锁。

## 刷新策略

第一期建议支持两种策略，但默认只实现其中一种也可以。

### 方案一：同步立即刷新

每次某个 `Progress` 更新时，直接触发 `MultiProgress.Refresh()`。

优点：

- 实现简单
- 行为直观

缺点：

- 高频更新时刷新次数多

### 方案二：dirty 标记 + 定时刷新

每个 `Progress` 更新时只标记 dirty，由 manager 的刷新 loop 周期性重绘。

优点：

- 刷新频率稳定
- 更适合大量并发任务

缺点：

- 结构更复杂

第一期推荐：

- 先实现同步立即刷新
- `AutoRefresh` 和 `RefreshInterval` 作为未来扩展预留

原因是第一期重点是建立可靠结构，不是极致性能。

## 生命周期设计

### Start

`MultiProgress.Start()` 的职责：

- 锁定并初始化内部状态
- 记录条目数量
- 输出多行占位区域
- 执行首次整体渲染

### 运行中

当某个条目调用：

- `Advance`
- `AdvanceTo`
- `AddMessage`
- `SetWidget`

只要最终展示内容可能变化，都应触发 manager 刷新。

### Finish

`MultiProgress.Finish()` 的职责：

- 执行最后一次整体渲染
- 在区域末尾换行，结束多条显示块
- 标记 manager 已结束，防止继续刷新

如果未来需要支持“单条完成后折叠成日志行”之类能力，可以另行扩展，但不纳入本期。

## 并发与线程安全

这是第一期设计中必须明确的部分。

### 1. 多条更新可能来自不同 goroutine

`Progress` 在多任务场景下通常会被不同 goroutine 更新，因此需要保证：

- manager 刷新过程串行
- 不会出现部分条目刷新、部分条目中途被写入导致的混乱

### 2. 状态读取需要一致性

`MultiProgress.Refresh()` 在读取每个条目的 `Line()` 时，需要尽量看到一致状态。

第一期可接受的简化方案：

- 由 manager 锁住整个刷新过程
- 要求 `Progress` 的更新路径在托管模式下也走 manager 控制的刷新路径

如果后续发现并发粒度需要更细，可再引入每条进度自己的锁，但第一期不建议一开始就做复杂化。

## 与现有 API 的兼容策略

第一期必须保证以下兼容性：

### 1. 单条 `Progress` 用法不变

以下代码行为保持不变：

```go
p := progress.Bar(100)
p.Start()
for i := 0; i < 100; i++ {
	p.Advance()
}
p.Finish()
```

### 2. 现有格式和 widgets 不变

`Format`、`AddWidget()`、`Messages`、`BarWidget()` 等能力继续工作，多条模式只是改变“谁来输出”，不改变“输出什么”。

### 3. 不改 `SpinnerFactory`

`SpinnerFactory` 保持现状。本期不将其塞入 `MultiProgress`，避免把两种不同生命周期模型在第一版强行合并。

## 未来扩展方向

虽然本期不做，但设计上应允许未来扩展为更通用的条目管理器。

可预留的演进方向：

- 让 `MultiProgress` 接受 `Progress` 和 `Spinner` 的统一渲染条目接口
- 支持条目级别的名称或前缀
- 支持动态插入或移除条目
- 支持完成条目折叠
- 支持定时刷新和脏标记策略

如果未来需要走得更远，可以考虑抽象出类似：

```go
type Renderable interface {
	Line() string
	Done() bool
}
```

但本期不建议提前抽象，避免为未来猜测的需求过度设计。

## 备选方案对比

### 方案 A：新增 `MultiProgress`，统一接管渲染

优点：

- 兼容性最好
- 对现有单条模式侵入最小
- 终端输出控制清晰
- 最适合第一期落地

缺点：

- 需要给 `Progress` 增加托管模式分支

### 方案 B：让每个 `Progress` 自己记录行号并分别刷新

优点：

- 听起来更直接

缺点：

- 每个实例都要操作光标
- 并发冲突复杂
- 容易出现行定位错误和残留字符问题

结论：不推荐。

### 方案 C：直接实现通用终端布局框架

优点：

- 理论扩展性最好

缺点：

- 明显超出当前需求
- 实现和维护成本过高

结论：本期不考虑。

## 测试设计

第一期建议覆盖以下测试场景：

- 单条 `Progress` 行为与当前版本保持一致
- 多个 `Progress` 接入同一个 `MultiProgress` 后能稳定输出多行
- 某一个条目更新时，整体区域重绘正确
- 多个 goroutine 并发更新不同条目时，不出现 panic 和明显错乱
- `MultiProgress.Finish()` 后不再继续刷新
- `Progress` 在托管模式下调用 `Finish()` 不额外破坏区域换行

对于终端控制序列相关测试，建议分两层：

- 纯状态测试：验证条目数量、刷新触发、生命周期状态
- writer 捕获测试：用 buffer 捕获输出，验证控制序列和重绘结果包含预期模式

## 实施步骤

推荐按以下顺序实现：

### 阶段 1：拆分字符串构建与直接输出

- 为 `Progress` 暴露 `Line()` 或等价方法
- 保持单条模式行为不变

### 阶段 2：引入 `MultiProgress`

- 支持注册多个 `Progress`
- 支持首次渲染和整体刷新
- 支持结束整个区域

### 阶段 3：让托管模式接管刷新

- `Progress` 检测是否被 manager 托管
- 托管模式下不再直接写终端
- 状态变更时交给 manager 刷新

### 阶段 4：补并发与兼容测试

- 覆盖多 goroutine 更新
- 验证旧 API 无回归

## 最终结论

`progress` 的多进度条支持，第一期最合理的方案是新增一个 `MultiProgress` 作为统一渲染管理器，仅接管多个 `Progress` 的并行显示，不将 `SpinnerFactory` 一并纳入。

最终推荐结论如下：

- 保持现有单条 `Progress` API 完全兼容
- 新增 `MultiProgress` 统一管理多条 `Progress`
- 将“构建显示内容”和“写终端”两个职责拆开
- 多条模式下仅允许 manager 负责终端输出
- 第一版使用整块重绘策略，不做脏行优化
- 暂不纳入 `SpinnerFactory`

这套方案实现风险低、兼容性好，能够在当前 `progress` 结构基础上稳妥引入多进度条能力，并为后续更复杂的渲染需求保留扩展空间。
