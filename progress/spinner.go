package progress

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gookit/cliui/cutypes"
	"github.com/gookit/color"
)

// BuilderFunc build char string
type BuilderFunc func() string

// SpinnerFactory definition. ref https://github.com/briandowns/spinner
type SpinnerFactory struct {
	// Out output writer. default is cutypes.Output
	Out io.Writer
	// Speed is the running speed
	Speed time.Duration
	// Format setting display format
	Format string
	// Builder build custom spinner text
	Builder BuilderFunc
	// locker
	lock *sync.RWMutex
	// mark spinner status
	active bool
	// control the spinner running.
	stopCh chan struct{}
}

// Spinner instance
func Spinner(speed time.Duration) *SpinnerFactory {
	return &SpinnerFactory{
		Out:    cutypes.Output,
		Speed:  speed,
		Format: "%s",
		// color: color.Normal.Sprint,
		lock: &sync.RWMutex{},
		// writer:   os.Stdout,
		stopCh: make(chan struct{}, 1),
	}
}

// RoundTripOptions for create Round-Trip Spinner
type RoundTripOptions struct {
	Char  rune
	Speed time.Duration
	CharN int // char number
	BoxW  int // box width
}

func defaultRoundTripOptions() *RoundTripOptions {
	return &RoundTripOptions{
		Char:  '=',
		Speed: 100 * time.Millisecond,
		CharN: 4,
		BoxW:  12,
	}
}

// SpinnerRoundTrip quickly create a Round-Trip Spinner instance.
func SpinnerRoundTrip(optFns ...func(opts *RoundTripOptions)) *SpinnerFactory {
	opts := defaultRoundTripOptions()
	if len(optFns) > 0 {
		for _, optFn := range optFns {
			optFn(opts)
		}
	}

	return Spinner(opts.Speed).WithBuilder(roundTripTextBuilder(opts.Char, opts.CharN, opts.BoxW))
}

// RoundTripLoading create. alias of RoundTripSpinner
func RoundTripLoading(char rune, speed time.Duration, charNumAndBoxWidth ...int) *SpinnerFactory {
	return RoundTripSpinner(char, speed, charNumAndBoxWidth...)
}

// RoundTripSpinner instance create. eg: [ =   ] - 字符在盒子里来回移动
func RoundTripSpinner(char rune, speed time.Duration, charNumAndBoxWidth ...int) *SpinnerFactory {
	charNum := 4
	boxWidth := 12
	if ln := len(charNumAndBoxWidth); ln > 0 {
		charNum = charNumAndBoxWidth[0]
		if ln > 1 {
			boxWidth = charNumAndBoxWidth[1]
		}
	}

	return Spinner(speed).WithBuilder(roundTripTextBuilder(char, charNum, boxWidth))
}

// LoadingSpinner instance create. 多个字符不停的变化，显示出旋转的效果
func LoadingSpinner(chars []rune, speed time.Duration) *SpinnerFactory {
	return Spinner(speed).WithBuilder(loadingCharBuilder(chars))
}

/*************************************************************
 * spinner running
 *************************************************************/

// WithBuilder set spinner text builder
func (s *SpinnerFactory) WithBuilder(builder BuilderFunc) *SpinnerFactory {
	s.Builder = builder
	return s
}

func (s *SpinnerFactory) prepare(format []string) {
	if s.Builder == nil {
		panic("spinner: field SpinnerFactory.Builder must be setting")
	}

	if len(format) > 0 {
		s.Format = format[0]
	}

	if s.Format != "" && !strings.Contains(s.Format, "%s") {
		s.Format = "%s " + s.Format
	}

	if s.Speed == 0 {
		s.Speed = 100 * time.Millisecond
	}
}

// Start run spinner
func (s *SpinnerFactory) Start(format ...string) {
	if s.active {
		return
	}

	s.active = true
	s.prepare(format)

	go func() {
		for {
			select {
			case <-s.stopCh:
				return
			default:
				s.lock.Lock()

				// \x0D - Move the cursor to the beginning of the line
				// \x1B[2K - Erase(Delete) the line
				fmt.Fprint(s.out(), "\x0D\x1B[2K")
				color.Fprintf(s.out(), s.Format, s.Builder())
				s.lock.Unlock()

				time.Sleep(s.Speed)
			}
		}
	}()
}

// Stop run spinner
func (s *SpinnerFactory) Stop(finalMsg ...string) {
	if !s.active {
		return
	}

	s.lock.Lock()
	s.active = false
	fmt.Fprint(s.out(), "\x0D\x1B[2K")

	if len(finalMsg) > 0 {
		fmt.Fprintln(s.out(), finalMsg[0])
	}

	s.stopCh <- struct{}{}
	s.lock.Unlock()
}

// Restart will stop and start the spinner
func (s *SpinnerFactory) Restart() {
	s.Stop()
	s.Start()
}

// Active status
func (s *SpinnerFactory) Active() bool {
	return s.active
}

func (s *SpinnerFactory) out() io.Writer {
	if s.Out != nil {
		return s.Out
	}
	return cutypes.Output
}
