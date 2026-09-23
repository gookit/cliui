# cliui 缺陷修复计划

- 日期: 2026-09-23
- 基线: `main` @ `eecb2e5`（工作区干净，`go vet ./...` 无输出，`go test ./...` 全绿）
- 来源: 对 `gookit2/cliui` 全量静态审查（`docs/`、`show/`、`interact/`、`progress/`、`cutypes/`），所有条目均已对照源码逐条核实
- 约定: 单元测试使用 `github.com/gookit/goutil/x/assert`（GR106）；单文件 < 1000 行（GR104）

## 背景

`cliui` 是从 `gookit/gcli/v3` 抽出的终端 UI 模块，分 `show` / `interact` / `progress` 三块。现有测试通过率高，但覆盖集中在 happy path，因此下面这些缺陷全部藏在未覆盖分支里。

## 验收标准

1. 阶段一列出的每个缺陷都有对应的失败复现（或明确边界）用例。
2. `go build ./... && go vet ./... && go test ./...` 全绿。
3. 不改变既有导出 API 的签名与默认行为（阶段一、二均遵守）。

---

## 阶段一：崩溃与正确性（本次执行）

### F1 `progress` 随机主题 off-by-one
- 位置: `progress/helper.go:33`、`helper.go:76`、`progress/quickstart.go:130`
- 根因: `rand.IntN(len(x)-1)` 取值范围 `[0,len-1)`，最后一个元素永远取不到。
- 修复: 改为 `rand.IntN(len(x))`。
- 测试: 多次调用 `RandomCharTheme/RandomCharsTheme/RandomBarStyle`，断言能命中最后一个元素（或遍历覆盖集合）。

### F2 `progress` RoundTrip 负数重复 panic
- 位置: `progress/widgets.go:214` `strings.Repeat(" ", boxWidth-position-charNum)`
- 根因: `charNum > boxWidth` 时长度为负，`strings.Repeat` panic；经 `SpinnerRoundTrip` 时在动画 goroutine 内 panic，进程崩溃。
- 修复: 在 builder 内对空格数做下界裁剪（`padding < 0 → 0`），或在构造时把 `charNum` 夹到 `<= boxWidth`。采用前者，保持输出可控。
- 测试: `RoundTripWidget('=', 20, 12)` 调用不 panic 且返回定宽字符串。

### F3 `progress` Reset 先于 Start 使状态错乱
- 位置: `progress/progress.go:416` `reset()` 内 `p.started = true`
- 根因: 未启动的 bar 调用 `Reset()` 后 `started=true`，随后 `Start()` 命中 `panic("already started")`（`progress.go:333`）。
- 修复: `reset()` 不再强制 `started=true`，保留调用前的启动状态（`Reset`/`ResetWith` 已自行记录 `wasStarted`）。已确认 `TestProgressReset`/`TestProgressResetManaged` 语义不变。
- 测试: `New(); Reset(); Start()` 不 panic；`Reset` 后 `Started()` 反映真实状态。

### F4 `progress` addWidget/setWidget 空 map panic
- 位置: `progress/progress.go:308`、`progress/progress.go:314`
- 根因: `addWidgets` 会初始化 `Widgets`，`addWidget`/`setWidget` 不会，零值 `&Progress{}` 上调用即 panic。
- 修复: 两处补 `if p.Widgets == nil { p.Widgets = make(map[string]WidgetFunc) }`。
- 测试: 零值 `&Progress{}` 上 `AddWidget`/`SetWidget` 不 panic。

### F5 `progress` Destroy 受管时写错 writer
- 位置: `progress/progress.go:562`
- 根因: 受管 bar 的 `render` 写 `p.out()`，绕开 `MultiProgress`，也不更新其行数统计，破坏下一次块重绘。
- 修复: 受管（`p.manager != nil`）时直接返回——受管输出由 manager 统一负责；独立 bar 行为不变。
- 测试: 受管 bar 调 `Destroy()` 不产出游离输出、不改动 manager 块。

### F6 `progress` Spinner Start/Stop/Restart 竞态
- 位置: `progress/spinner.go`（`Start`/`Stop`/`Restart`/`Active`）
- 根因: 复用同一个带缓冲 `stopCh`；`Restart=Stop+Start` 时残留 token 可能被新 goroutine 抢走导致新动画立即退出、旧 goroutine 泄漏；`active` 未加锁；`prepare` panic 后 `active` 卡在 true。
- 修复: 每次 `Start` 新建 `stopCh`，`Stop` 改为 `close(stopCh)`；`active` 读写统一走 `s.lock`；先校验/`prepare` 成功再置 `active=true`。
- 测试: `Start→Stop→Start` 后 `Active()==true` 且能再次 `Stop`；`Start` 两次不产生第二个 goroutine。

### F7 `show/emoji` ToUnicode 与 Render 正则
- 位置: `show/emoji/emoji.go:101`（`ToUnicode`）、`emoji.go:9`（`codeMatch`）
- 根因: `emoji[0]` 取的是首字节（`"💖"`→`"f0"`），空串 panic；`(:\w+:)` 匹配不到含 `+`/`-` 的名字（map 中有 4 个，如 `:+1:`、`:-1:`、`:e-mail:`）。
- 修复: 用 `utf8.DecodeRuneInString` 取首 rune，空串返回 `""`；正则改为 `(:[\w+-]+:)`。
- 测试: `ToUnicode("💖")=="1f496"`、`ToUnicode("")==""`；`Render(":+1:")` 命中映射。

### F8 `show/table` 表头按字节截断
- 位置: `show/table/table.go:416` `strutil.Resize`（表体用 `Utf8Resize`，`table.go:500`）
- 根因: `Resize` 以 `len(s)` 衡量并在溢出时按字节切片，`colWidths` 却是显示宽度，多字节表头被截成非法 UTF-8。
- 修复: 表头改用 `strutil.Utf8Resize`，与表体一致。
- 测试: 中文表头渲染后无非法字节、列宽与表体一致。

### F9 `show/table` 重复渲染破坏表格
- 位置: `show/table/table.go:258`（`reset`）、`296`（`PrependHead("#")`）、`346`（前插行号单元格）、`599`(`Row.Init` 早退)
- 根因: `prepare()` 每次 `Format()` 都前插 `#` 表头与行号单元格，且 `reset()` 不还原 `row.init`/表头单元；`Render()` 又直接再调 `Format()`，导致每次渲染多出一列。
- 修复: `reset()` 中追加“撤销上次注入”，并同时重置 `row.init` 与表头单元 `init/width/height`；`reset()` 重置 `headHeight`。使 `Format()` 幂等。
- 测试: 同一 `Table` 连续 `Render()`/`String()`/`Render()` 输出稳定、不新增列。

### F10 `show/title` 左右对齐宽度多 1
- 位置: `show/title/title.go:162`（`renderLeft`）、`230`（`renderRight`）
- 根因: `PaddingLR` 分支总宽 = `width+1`，默认标题比边框宽一列；`renderRight` 在 `width-2 < titleLen < width` 时 `make([]rune, 负数)` panic。
- 修复: 右侧填充量改为 `width - titleLen - 3`；两个分支的溢出保护统一为 `titleLen >= width-2` 时返回未填充标题。
- 测试: 三种对齐的内容行显示宽度都等于 `Width`（复用 `titleBorderWidth` 思路）。

### F11 `interact/ui` Input/Confirm 被空格/Tab/resize 误提交
- 位置: `interact/ui/input.go`（尾部分支 143-156）、`interact/ui/confirm.go`（71-75）
- 根因: `readline` 把空格/Tab 发成无 `Text` 的按键事件（`readline.go:176,194`），resize 事件也是空文本；两组件无对应分支，落入“空文本即提交默认值”的兜底。
- 修复: `Input` 增加 `KeySpace`（插入空格）、`KeyTab`（忽略）、`EventResize`（跳过）分支；尾部兜底由“提交”改为“忽略”。`Confirm` 增加 `EventResize` 跳过，空文本兜底改为忽略（提交仅由 `KeyEnter` 触发）。
- 测试: fake 事件流中夹入 `KeySpace`/`KeyTab`/`EventResize`，断言 `Input` 正常收字并可输入空格、`Confirm` 不被误提交。

### F12 `interact` Select 污染调用方 map
- 位置: `interact/select.go:115`（`s.valMap = optsData`）、`160`（注入 `"q"`）
- 根因: `map[string]string` 分支直接别名调用方 map，随后写入 `"q"`，静默篡改调用方数据。
- 修复: 复制到内部新 map，再注入 `"q"`。
- 测试: 传入 map 调用后，断言原 map 未新增 `"q"` 键。

---

## 阶段二：轻量性能（本次执行）

### F13 `progress` memory widget 每帧 ReadMemStats
- 位置: `progress/widgets.go:48`
- 根因: `runtime.ReadMemStats` 会 STW，`Full` 格式每次重绘都调用。
- 修复: 按固定 TTL（如 1s）缓存读取结果，复用上次值。
- 测试: 连续多次渲染不改变输出健壮性（功能不变），必要时以帧间一致性断言。

### F14 `show/emoji` 每次调用重编译正则
- 位置: `show/emoji/emoji.go`（`FromUnicode`、`Decode` 内各自 `regexp.MustCompile`）
- 修复: 提升为包级 `var`；`Decode` 复用 `FromUnicode` 逻辑。
- 测试: 现有行为不变（补充 `Decode`/`Encode` 往返用例）。

### F15 补齐关键回归测试
- 为 F1–F14 各补最小失败用例；填充当前为空的 `show/emoji/emoji_test.go`。

---

## 阶段三：待决策（本次不改，需单独确认）

以下改动会改变导出 API 或默认行为，超出“修复”范围，先列不改：

1. legacy `interact` 的 `os.Exit`/`panic`（`base.go`、`read.go`、`select.go` 多处）→ 改为返回 error：破坏现有调用方契约。
2. `show.JSON`/`TabWriter` 出错 `panic` → 返回 error：改公开行为。
3. `OptionFunc` 签名统一（`title` 与 banner/list/table 不一致）：破坏兼容。
4. 删除死代码：`show/writer.go` 空 `Print`/恒 nil `Flush`、`showcom/box.go` 空文件、`multi.update`、`Progresser`：删导出符号属破坏性变更。
5. 输出流语义统一（`Progress` 构造期快照 `cutypes.Output` vs `MultiProgress` 运行期解析）。
6. `table.OverflowWrap` 实际未换行；`table` 边框角字符受 `BorderLeft/Right` 控制（`BorderDefault` 不含 L/R）。
7. `interact/backend/plain`、`prompt.go` 取消时 goroutine 泄漏；`readline` 控制字节（Ctrl-D）未识别。

## 执行结果（2026-09-23）

阶段一 F1–F12、阶段二 F13–F15 已完成，全部为逻辑修复，未改动导出 API 签名与默认行为。

- 改动文件: `progress/{helper,quickstart,widgets,progress,spinner}.go`、`show/emoji/emoji.go`、`show/table/table.go`、`show/title/title.go`、`interact/{select.go,ui/input.go,ui/confirm.go}`
- 新增回归测试: `progress/bugfix_test.go`、`show/emoji/emoji_test.go`(原为空)、`show/table/bugfix_test.go`、`show/title/bugfix_test.go`、`interact/bugfix_test.go`、`interact/ui/bugfix_test.go`
- 验证: `go build ./...`、`go vet ./...` 干净；`go test ./...` 全绿；`go test -count=3 ./progress/... ./interact/... ./show/...` 稳定通过
- 未执行: `go test -race`（本机无 cgo/C 编译器，`-race requires cgo`），相关并发改动（F6）改为逻辑审查 + 重复运行验证

阶段三条目仍未改动，待确认后再逐条处理。

## 风险与回退

- 所有改动限于逻辑修复，均有失败复现用例；每个功能点完成后 `go test ./...`，可按文件/阶段 `git checkout` 回退。
- F9（表格幂等）与 F6（spinner）改动面相对大，若回归优先回退这两处单独提交。
