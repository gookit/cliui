package interact

import (
	"syscall"
)

func syscallStdin() int {
	// on Windows, must convert 'syscall.Stdin' to int
	return int(syscall.Stdin)
}
