// Package progress provide terminal progress bar display.
// Such as: `Txt`, `Bar`, `Loading`, `RoundTrip`, `DynamicText` ...
package progress

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/gookit/cliui/cutypes"
	"github.com/gookit/color"
)

// use for match like "{@bar}" "{@percent:3s}" "{@name:.20s}"
var widgetMatch = regexp.MustCompile(`{@([\w_]+)(?::([^}]+))?}`)

var widgetAliases = map[string]string{
	"eta": "remaining",
}

// WidgetFunc handler func for progress widget
type WidgetFunc func(p *Progress) string

// Progresser progress interface
type Progresser interface {
	Start(maxSteps ...int64)
	Advance(steps ...int64)
	AdvanceTo(step int64)
	Finish(msg ...string)
	Bound() any
}

// Progress definition
// Refer:
//
//	https://github.com/inhere/php-console/blob/master/src/utils/ProgressBar.php
type Progress struct {
	// Out output writer. default is cutypes.Output
	Out io.Writer
	// Format string the bar format
	Format string
	// Newline render progress on newline
	Newline bool
	// MaxSteps maximal steps.
	MaxSteps int64
	// StepWidth the width for display "{@current}". default 0 means auto.
	// eg: 342 计数场景，StepWidth = 3. 用于计算填充空格宽度
	StepWidth uint8
	// Overwrite prev output. default is True
	Overwrite bool
	// RedrawFreq redraw freq. default is 1
	RedrawFreq uint
	// Widgets for build the progress bar
	Widgets map[string]WidgetFunc
	// Messages named messages for build progress bar
	// Example:
	// 	{"msg": "downloading ..."}
	// 	"{@percent}% {@msg}" -> "83% downloading ..."
	Messages map[string]string
	// current step value
	step int64
	// bound user custom data.
	bound any
	// mark start status
	started bool
	// completed percent. eg: "83.8"
	percent float32
	// mark is first running
	firstRun bool
	// time consumption record
	startedAt  time.Time
	finishedAt time.Time
	// managed by a multi progress renderer
	manager *MultiProgress
	index   int
	hidden  bool
	removed bool
}

/*************************************************************
 * quick use
 *************************************************************/

// New Progress instance
func New(maxSteps ...int64) *Progress {
	var max int64
	if len(maxSteps) > 0 {
		max = normalizeMaxSteps(maxSteps[0])
	}

	return &Progress{
		Out:       cutypes.Output,
		Format:    DefFormat,
		MaxSteps:  max,
		Overwrite: true,
		// init widgets
		Widgets: make(map[string]WidgetFunc),
		// add a default message
		Messages: map[string]string{"message": ""},
	}
}

// NewWithConfig create new Progress
func NewWithConfig(fn func(p *Progress), maxSteps ...int64) *Progress {
	return New(maxSteps...).Config(fn)
}

/*************************************************************
 * Some quick config methods
 *************************************************************/

// RenderFormat set rendered format option
func RenderFormat(f string) func(p *Progress) {
	return func(p *Progress) {
		p.Format = f
	}
}

// MaxSteps setting max steps
func MaxSteps(maxStep int64) func(p *Progress) {
	return func(p *Progress) {
		p.MaxSteps = normalizeMaxSteps(maxStep)
	}
}

/*************************************************************
 * Config progress
 *************************************************************/

// Config the progress instance
func (p *Progress) Config(fn func(p *Progress)) *Progress {
	fn(p)
	return p
}

// WithOptions add more option at once for the progress instance
func (p *Progress) WithOptions(fns ...func(p *Progress)) *Progress {
	if len(fns) > 0 {
		for _, fn := range fns {
			fn(p)
		}
	}
	return p
}

// WithMaxSteps setting max steps
func (p *Progress) WithMaxSteps(maxSteps ...int64) *Progress {
	if len(maxSteps) > 0 {
		p.MaxSteps = normalizeMaxSteps(maxSteps[0])
	}
	return p
}

// SetMaxSteps updates the progress max steps.
func (p *Progress) SetMaxSteps(maxSteps int64) *Progress {
	if p.manager != nil {
		p.manager.updateBar(updateConfig, p, func() bool {
			p.MaxSteps = normalizeMaxSteps(maxSteps)
			return true
		})
		return p
	}

	p.MaxSteps = normalizeMaxSteps(maxSteps)
	return p
}

// SetFormat updates the progress render format.
func (p *Progress) SetFormat(format string) *Progress {
	if p.manager != nil {
		p.manager.updateBar(updateConfig, p, func() bool {
			p.Format = format
			return true
		})
		return p
	}

	p.Format = format
	return p
}

// Binding user custom data to instance
func (p *Progress) Binding(data any) *Progress {
	p.bound = data
	return p
}

// Bound get bound sub struct instance
func (p *Progress) Bound() any {
	return p.bound
}

// AddMessage to progress instance
func (p *Progress) AddMessage(name, message string) {
	if p.manager != nil {
		p.manager.updateBar(updateConfig, p, func() bool {
			p.addMessage(name, message)
			return true
		})
		return
	}

	p.addMessage(name, message)
}

// SetMessage sets a named message and returns the progress instance.
func (p *Progress) SetMessage(name, message string) *Progress {
	if p.manager != nil {
		p.manager.updateBar(updateConfig, p, func() bool {
			p.addMessage(name, message)
			return true
		})
		return p
	}

	p.addMessage(name, message)
	return p
}

// AddMessages to progress instance
func (p *Progress) AddMessages(msgMap map[string]string) {
	if p.manager != nil {
		p.manager.updateBar(updateConfig, p, func() bool {
			p.addMessages(msgMap)
			return true
		})
		return
	}

	p.addMessages(msgMap)
}

// SetMessages sets multiple named messages and returns the progress instance.
func (p *Progress) SetMessages(msgMap map[string]string) *Progress {
	if p.manager != nil {
		p.manager.updateBar(updateConfig, p, func() bool {
			p.addMessages(msgMap)
			return true
		})
		return p
	}

	p.addMessages(msgMap)
	return p
}

func (p *Progress) addMessage(name, message string) {
	if p.Messages == nil {
		p.Messages = make(map[string]string)
	}
	p.Messages[name] = message
}

func (p *Progress) addMessages(msgMap map[string]string) {
	if p.Messages == nil {
		p.Messages = make(map[string]string)
	}

	for name, message := range msgMap {
		p.Messages[name] = message
	}
}

// AddWidget to progress instance
func (p *Progress) AddWidget(name string, handler WidgetFunc) *Progress {
	if p.manager != nil {
		p.manager.updateBar(updateConfig, p, func() bool {
			p.addWidget(name, handler)
			return true
		})
		return p
	}

	p.addWidget(name, handler)
	return p
}

// SetWidget to progress instance
func (p *Progress) SetWidget(name string, handler WidgetFunc) *Progress {
	if p.manager != nil {
		p.manager.updateBar(updateConfig, p, func() bool {
			p.setWidget(name, handler)
			return true
		})
		return p
	}

	p.setWidget(name, handler)
	return p
}

// AddWidgets to progress instance
func (p *Progress) AddWidgets(widgets map[string]WidgetFunc) {
	if p.manager != nil {
		p.manager.updateBar(updateConfig, p, func() bool {
			p.addWidgets(widgets)
			return true
		})
		return
	}

	p.addWidgets(widgets)
}

func (p *Progress) addWidget(name string, handler WidgetFunc) {
	if _, ok := p.Widgets[name]; !ok {
		p.Widgets[name] = handler
	}
}

func (p *Progress) setWidget(name string, handler WidgetFunc) {
	p.Widgets[name] = handler
}

func (p *Progress) addWidgets(widgets map[string]WidgetFunc) {
	if p.Widgets == nil {
		p.Widgets = make(map[string]WidgetFunc)
	}

	for name, handler := range widgets {
		p.addWidget(name, handler)
	}
}

/*************************************************************
 * running
 *************************************************************/

// Start the progress bar
func (p *Progress) Start(maxSteps ...int64) {
	if p.started {
		panic("Progress bar already started")
	}

	if p.manager != nil {
		p.manager.startProgress(p, maxSteps...)
		return
	}

	// init
	p.init(maxSteps...)

	// render
	p.Display()
}

func (p *Progress) init(maxSteps ...int64) {
	p.step = 0
	p.percent = 0.0
	p.started = true
	p.firstRun = true
	p.startedAt = time.Now()

	if p.RedrawFreq == 0 {
		p.RedrawFreq = 1
	}

	if len(maxSteps) > 0 {
		p.MaxSteps = normalizeMaxSteps(maxSteps[0])
	} else {
		p.MaxSteps = normalizeMaxSteps(p.MaxSteps)
	}

	// load default widgets
	p.addWidgets(builtinWidgets)
}

// Reset resets the progress state so the instance can be reused.
func (p *Progress) Reset(maxSteps ...int64) {
	if p.manager != nil {
		p.manager.updateBar(updateKeyState, p, func() bool {
			p.reset(maxSteps...)
			return true
		})
		return
	}

	wasStarted := p.started
	p.reset(maxSteps...)
	if wasStarted {
		p.Display()
	}
}

// ResetWith resets the progress and applies fn while holding the manager lock
// when the progress is managed.
func (p *Progress) ResetWith(fn func(p *Progress)) {
	if p.manager != nil {
		p.manager.updateBar(updateKeyState, p, func() bool {
			p.reset()
			if fn != nil {
				fn(p)
				p.MaxSteps = normalizeMaxSteps(p.MaxSteps)
			}
			return true
		})
		return
	}

	wasStarted := p.started
	p.reset()
	if fn != nil {
		fn(p)
		p.MaxSteps = normalizeMaxSteps(p.MaxSteps)
	}
	if wasStarted {
		p.Display()
	}
}

func (p *Progress) reset(maxSteps ...int64) {
	p.step = 0
	p.percent = 0.0
	p.started = true
	p.firstRun = true
	p.startedAt = time.Now()
	p.finishedAt = time.Time{}

	if p.RedrawFreq == 0 {
		p.RedrawFreq = 1
	}

	if len(maxSteps) > 0 {
		p.MaxSteps = normalizeMaxSteps(maxSteps[0])
	} else {
		p.MaxSteps = normalizeMaxSteps(p.MaxSteps)
	}

	p.addWidgets(builtinWidgets)
}

// Advance specified step size. default is 1
func (p *Progress) Advance(steps ...int64) {
	var step int64 = 1
	if len(steps) > 0 && steps[0] > 0 {
		step = steps[0]
	}

	if p.manager != nil {
		p.manager.updateBar(updateProgress, p, func() bool {
			p.checkStart()
			return p.applyStep(p.step + step)
		})
		return
	}

	p.checkStart()
	p.AdvanceTo(p.step + step)
}

// AdvanceTo specified number of steps
func (p *Progress) AdvanceTo(step int64) {
	if p.manager != nil {
		p.manager.updateBar(updateProgress, p, func() bool {
			p.checkStart()
			return p.applyStep(step)
		})
		return
	}

	p.checkStart()
	if p.applyStep(step) {
		p.display()
	}
}

func (p *Progress) applyStep(step int64) bool {
	// check arg
	if step < 0 {
		step = 0
	}
	if p.MaxSteps > 0 && step > p.MaxSteps {
		p.MaxSteps = step
	}

	freq := uint64(p.RedrawFreq)
	prevPeriod := uint64(p.step) / freq
	currPeriod := uint64(step) / freq

	p.step = step
	if p.MaxSteps > 0 {
		p.percent = float32(p.step) / float32(p.MaxSteps)
	}

	return prevPeriod != currPeriod || p.MaxSteps == step
}

// Finish the progress output.
// if provide finish message, will delete progress bar then print the message.
func (p *Progress) Finish(message ...string) {
	if p.manager != nil {
		p.manager.updateBar(updateFinalState, p, func() bool {
			p.checkStart()
			p.finishManaged(message...)
			return true
		})
		return
	}

	p.checkStart()
	p.finishedAt = time.Now()

	if p.MaxSteps == 0 {
		p.MaxSteps = p.step
	}

	// prevent double 100% output
	if p.step == p.MaxSteps && !p.Overwrite {
		return
	}

	p.AdvanceTo(p.MaxSteps)

	if len(message) > 0 {
		p.render(message[0])
	}

	fmt.Fprintln(p.out()) // new line
}

// Done marks the progress as completed.
func (p *Progress) Done(message ...string) {
	p.finishStatus("done", true, message...)
}

// Fail marks the progress as failed without advancing it to completion.
func (p *Progress) Fail(message ...string) {
	p.finishStatus("failed", false, message...)
}

// Skip marks the progress as skipped without advancing it to completion.
func (p *Progress) Skip(message ...string) {
	p.finishStatus("skipped", false, message...)
}

// Display outputs the current progress string.
func (p *Progress) Display() {
	if p.manager != nil {
		p.manager.updateBar(updateSilent, p, func() bool {
			return true
		})
		return
	}

	p.display()
}

func (p *Progress) display() {
	if p.Format == "" {
		p.Format = DefFormat
	}

	p.render(p.buildLine())
}

// Destroy removes the progress bar from the current line.
//
// This is useful if you wish to write some output while a progress bar is running.
// Call display() to show the progress bar again.
func (p *Progress) Destroy() {
	if p.Overwrite {
		p.render("")
	}
}

// Write advances the progress by the number of bytes in bs.
func (p *Progress) Write(bs []byte) (int, error) {
	n := len(bs)
	if n > 0 {
		p.Advance(int64(n))
	}
	return n, nil
}

// WrapReader wraps r and advances the progress by each successful read count.
func (p *Progress) WrapReader(r io.Reader) io.Reader {
	return progressReader{reader: r, progress: p}
}

// WrapWriter wraps w and advances the progress by each successful write count.
func (p *Progress) WrapWriter(w io.Writer) io.Writer {
	return progressWriter{writer: w, progress: p}
}

type progressReader struct {
	reader   io.Reader
	progress *Progress
}

func (r progressReader) Read(bs []byte) (int, error) {
	n, err := r.reader.Read(bs)
	if n > 0 {
		r.progress.Advance(int64(n))
	}
	return n, err
}

type progressWriter struct {
	writer   io.Writer
	progress *Progress
}

func (w progressWriter) Write(bs []byte) (int, error) {
	n, err := w.writer.Write(bs)
	if n > 0 {
		w.progress.Advance(int64(n))
	}
	return n, err
}

/*************************************************************
 * helper methods
 *************************************************************/

// render progress bar to terminal
func (p *Progress) render(text string) {
	if p.Overwrite {
		// first run. create new line
		if p.firstRun && p.Newline {
			fmt.Fprintln(p.out())
			p.firstRun = false

			// delete prev rendered line.
		} else {
			// \x0D - Move the cursor to the beginning of the line
			// \x1B[2K - Erase(Delete) the line
			fmt.Fprint(p.out(), "\x0D\x1B[2K")
		}

		color.Fprint(p.out(), text)
	} else if p.step > 0 {
		color.Fprintln(p.out(), text)
	}
}

func (p *Progress) out() io.Writer {
	if p.Out != nil {
		return p.Out
	}
	return cutypes.Output
}

func (p *Progress) currentStepWidth() int {
	if p.StepWidth > 0 {
		return int(p.StepWidth)
	}
	if p.MaxSteps > 0 {
		return len(fmt.Sprint(p.MaxSteps))
	}
	return 3
}

func (p *Progress) checkStart() {
	if !p.started {
		panic("Progress bar has not yet been started.")
	}
}

func (p *Progress) finishManaged(message ...string) {
	p.finishedAt = time.Now()

	if p.MaxSteps == 0 {
		p.MaxSteps = p.step
	}

	p.step = p.MaxSteps
	if p.MaxSteps > 0 {
		p.percent = 1
	}

	if len(message) > 0 {
		p.addMessage("message", message[0])
	}
}

func (p *Progress) finishStatus(status string, complete bool, message ...string) {
	if p.manager != nil {
		p.manager.updateBar(updateFinalState, p, func() bool {
			p.checkStart()
			p.applyStatus(status, complete, message...)
			return true
		})
		return
	}

	p.checkStart()
	p.applyStatus(status, complete, message...)
	p.display()
	fmt.Fprintln(p.out())
}

func (p *Progress) applyStatus(status string, complete bool, message ...string) {
	p.finishedAt = time.Now()

	if complete {
		if p.MaxSteps == 0 {
			p.MaxSteps = p.step
		}
		p.step = p.MaxSteps
		if p.MaxSteps > 0 {
			p.percent = 1
		}
	} else if p.MaxSteps > 0 {
		p.percent = float32(p.step) / float32(p.MaxSteps)
	}

	text := status
	if len(message) > 0 {
		text = message[0]
	}

	p.addMessage("message", text)
	p.addMessage("status", status)
}

// build widgets form Format string.
func (p *Progress) buildLine() string {
	return widgetMatch.ReplaceAllStringFunc(p.Format, func(s string) string {
		var text string
		// {@current} -> current
		// {@percent:3s} -> percent:3s
		name := strings.Trim(s, "{@}")
		fmtArg := ""

		// percent:3s
		if pos := strings.IndexRune(name, ':'); pos > 0 {
			fmtArg = name[pos+1:]
			name = name[0:pos]
		}

		if handler, ok := p.Widgets[name]; ok {
			text = handler(p)
		} else if msg, ok := p.Messages[name]; ok {
			text = msg
		} else if handler := p.Handler(name); handler != nil {
			text = handler(p)
		} else {
			return s
		}

		// like {@percent:3s} "7%" -> "  7%"
		if fmtArg != "" {
			text = fmt.Sprintf("%"+fmtArg, text)
		}
		// fmt.Println("info:", arg, name, ", text:", text)
		return text
	})
}

// Line returns the current rendered progress line.
func (p *Progress) Line() string {
	p.checkStart()
	if p.Format == "" {
		p.Format = DefFormat
	}

	return p.buildLine()
}

// Handler get widget handler by widget name
func (p *Progress) Handler(name string) WidgetFunc {
	if handler, ok := p.Widgets[name]; ok {
		return handler
	}
	if widgetName, ok := widgetAliases[name]; ok {
		return p.Widgets[widgetName]
	}

	return nil
}

/*************************************************************
 * getter methods
 *************************************************************/

// Percent gets the current percent
func (p *Progress) Percent() float32 {
	return p.percent
}

// Started reports whether the progress has been started.
func (p *Progress) Started() bool {
	return p.started
}

// Finished reports whether the progress has been finished.
func (p *Progress) Finished() bool {
	return !p.finishedAt.IsZero()
}

// Step gets the current step position.
func (p *Progress) Step() int64 {
	return p.step
}

// Max gets the max steps.
func (p *Progress) Max() int64 {
	return p.MaxSteps
}

// Progress alias of the Step()
func (p *Progress) Progress() int64 {
	return p.step
}

// StartedAt time get
func (p *Progress) StartedAt() time.Time {
	return p.startedAt
}

// FinishedAt time get
func (p *Progress) FinishedAt() time.Time {
	return p.finishedAt
}
