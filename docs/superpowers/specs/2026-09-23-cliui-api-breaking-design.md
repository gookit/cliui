# cliui 破坏性 API 调整设计

- 日期: 2026-09-23
- 基线: `main`（阶段一/二缺陷修复已合并）
- 范围: 仅破坏性 API 调整（不含非破坏性修复与语义调整）

## 背景

旧 `interact` 层在库内部直接 `os.Exit`/`panic`，无法嵌入、测试或取消；`show.JSON`/`TabWriter` 出错直接 panic；`title.OptionFunc` 与 banner/lists/table 签名不一致；另有若干死代码。本次统一处理这些破坏性项。

## 决策

1. **OptionFunc 统一为 `func(*Options)`**：只改 `title`（把 `widthSet` 移入 `Options`），banner/lists/table 保持不变。四个包签名一致，改动面最小。

2. **`show.JSON` / `show.TabWriter` 返回 error**：`JSON` → `(int, error)`（保留 `OK`/`ERR` 语义），`TabWriter` → `(*tabwriter.Writer, error)`。`PrettyJSON.Format` 不再 panic，错误写入 `pj.Err`。

## 变更后的公开表面

### interact（legacy 层：从 os.Exit/panic 改为返回 error）

| 旧签名 | 新签名 |
|---|---|
| `Question.Run() *Value` | `Question.Run() (*Value, error)` |
| `Select.Run() *SelectResult` | `Select.Run() (*SelectResult, error)` |
| `AnswerIsYes(defVal ...bool) bool` | `AnswerIsYes(defVal ...bool) (bool, error)` |
| `Confirm(msg string, defVal ...bool) bool` | `Confirm(msg string, defVal ...bool) (bool, error)` |
| `Unconfirmed(msg string, defVal ...bool) bool` | `Unconfirmed(msg string, defVal ...bool) (bool, error)` |
| `Ask(...) string` | `Ask(...) (string, error)` |
| `Query(...) string` | `Query(...) (string, error)` |
| `Choice(...) string` | `Choice(...) (string, error)` |
| `SingleSelect(...) string` | `SingleSelect(...) (string, error)` |
| `SelectOne(...) string` | `SelectOne(...) (string, error)` |
| `SelectOneKey(...) string` | `SelectOneKey(...) (string, error)` |
| `Checkbox(...) []string` | `Checkbox(...) ([]string, error)` |
| `MultiSelect(...) []string` | `MultiSelect(...) ([]string, error)` |

新增错误值：

- `interact.ErrQuit`：用户在选择中主动退出（旧行为是 `os.Exit(0)`）。
- `interact.ErrMaxAttempts`：超过 `Question.MaxTimes` 允许的错误次数（旧行为是 `os.Exit`）。

删除（unexported，无公开影响）：`exitWithErr`、`exitWithMsg`。

### show

| 旧签名 | 新签名 |
|---|---|
| `JSON(v any, prefixAndIndent ...string) int` | `JSON(v any, prefixAndIndent ...string) (int, error)` |
| `TabWriter(rows []string) *tabwriter.Writer` | `TabWriter(rows []string) (*tabwriter.Writer, error)` |

- `PrettyJSON.Format()` 签名不变（仍是 `func()`），但失败时不再 panic，而是写入 `Base.Err`。
- `AnyData` 签名不变，内部忽略 `JSON` 的错误（无 panic 风险）。

### title

| 旧签名 | 新签名 |
|---|---|
| `type OptionFunc func(t *Title)` | `type OptionFunc func(o *Options)` |

- 所有 `WithXxx() OptionFunc` 改为配置 `*Options`；`Title.WithOptionFns([]OptionFunc)` 改为应用 `&t.Options`。
- `Title` 的私有字段 `widthSet` 移入 `Options.widthSet`（未导出，无公开影响）。

### 删除的死代码（导出符号，属破坏性变更）

- `show.Writer.Print()`、`show.Writer.Flush()`
- `progress.Progresser` 接口
- `progress.(*MultiProgress).update`（未导出）
- 文件 `show/showcom/box.go`（空文件）

## 迁移指南

```go
// interact: 现在必须处理 error
ok, err := interact.Confirm("Continue?", true)
if err != nil { return err }

name, err := interact.Ask("Your name?", "guest", nil)
if err != nil { return err }

result, err := interact.NewSelect("Env", opts).Run()
if err != nil {
	if errors.Is(err, interact.ErrQuit) { /* 用户退出 */ }
	return err
}

// show
code, err := show.JSON(data)
w, err := show.TabWriter(rows)

// title: option func 参数变为 *title.Options
title.New("Deploy", func(o *title.Options) { o.Width = 40 })
```

## 后续补充（同批完成）

- 清理遗留死符号：删除 `interact.RunFace`、`interact.ComOptions`、`interact.ItemFace`（`cparam.InputParam.ValidFn` 是字段，保留）。
- 非破坏性项一并处理：
  - readline：`Ctrl-D` 触发中断（EOF），其他控制字节不再变成字面文本。
  - table：top/bottom 边框角字符不再受 `BorderLeft/BorderRight` 限制（新增 `edge` 参数）；`OverflowWrap` 现在会按列宽真正换行。
  - progress：`Progress`/`SpinnerFactory` 的 `Out` 默认留空，改为渲染时解析 `cutypes.Output`，与 `MultiProgress` 语义一致。
  - interact：`plain` backend 与 `Prompt` 改为每个输入流复用一个读取 goroutine，取消后不再重复读取或丢失已到达的行。

## 后续补充（二）

- `table`：反射处理 `[]map[string]string` 等非 `map[string]any` 行时，`MapIndex` 的 key 由 `*Cell` 改为按表头名转换到 map 的 key 类型，避免 panic。
- `show` 缓存失效：`table`、`lists`、`banner`、`title` 的 mutator（Add/Set/With*/Config*）现在会调用 `ResetFormat()`，`String()` 不再返回变更前的旧结果；`List`/`Lists`/`PrettyJSON` 的 `Format` 改为幂等（先 `InitBuffer`），`Lists.Format` 让每个子列表使用独立 buffer。
- `lists`：key 列填充改用显示宽度（`strutil.Utf8Padding` + `TextWidth`），中文 key 与英文 key 的 value 起始列对齐。
- `progress`：`MultiProgress` 在可见块变短时先清空旧块再重绘，`Refresh()` 后不再残留多余行。

## 后续补充（三）

- `progress`：`IsTerminal` 除 `*os.File` 外也识别实现了 `Fd() uintptr` 的包装 writer，`UseAutoRenderMode` 不再误判为非终端。
- `show/alert`：`Render`/`Println` 复用同一个 banner 构造（`newBanner`），去掉重复的选项列表。
- `readline`：`Render` 不再吞掉写错误，改为返回首个错误。
- `interact`：`Question.Run` 使用默认值时也会走校验函数；`StepsRun` 新增 `RunContext(ctx)`，在步骤间检查取消（`Run` 委托给它）。
- `interact`：`internal.ReadLineWithOutput` 改为每个输入流复用一个 `bufio.Reader`，避免一次性管道/文件里的后续行被丢弃。
- `interact`：`GetHiddenInput` 的 Windows 临时脚本改为 `.vbs` 后缀（`cscript` 依赖扩展名），并对 VB 双引号与 sh 单引号做转义、检查写入错误。

## 后续补充（四）

- `progress`：`Start` 对受管 bar 不再无锁读取 `p.started`，改由 `manager.startProgress` 在锁内校验；`Finish` 增加“已结束则直接返回”保护，避免重复换行；`GetCharTheme(0)`/`GetCharsTheme(0)` 现在返回索引 0 的主题（此前会随机）。
- `show/table`：`SortColumn` 在两列都是数字时按数值排序（此前 `"10" < "9"` 的字典序）。

## 仍未处理（有意保留）

- `progress`：`Progress.manager` 仍以“Add 后使用”为前提无锁读取；`started` 在只读访问器（`Started()` 等）中未加锁——需引入 `Progress` 级锁并明确与 `MultiProgress` 锁的顺序，属较大改动。
- `progress`：`Finish(msg)` 按文档语义“删除进度行并打印原始消息”，与 `Done/Fail/Skip` 的格式化输出不同，属既定设计，未改。
- `show/table`：单元格值含 color 标签时宽度测量未剔除标签（需带标签的 resize 辅助函数，属边界场景）。
- `interact`：`readline` 的 `ctx` 仍只在阻塞前检查（需平台相关的 read deadline）；终端恢复无信号保护（库内注册全局信号处理有副作用）；`GetHiddenInput` 的 Windows `InputBox` 仍明文显示（需换用 PowerShell `-AsSecureString` 等机制）。
