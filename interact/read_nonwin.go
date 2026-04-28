//go:build !windows
// +build !windows

package interact

import (
	"syscall"
)

func syscallStdin() int {
	return syscall.Stdin
}
