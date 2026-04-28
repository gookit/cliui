package interact

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gookit/cliui/cutypes"
	"github.com/gookit/cliui/internal"
	"github.com/gookit/color"
	"github.com/gookit/goutil/arrutil"
	"github.com/gookit/goutil/cliutil"
	"github.com/gookit/goutil/envutil"
	"golang.org/x/term"
)

// ReadInput read user input form Stdin
func ReadInput(question string) (string, error) {
	if len(question) > 0 {
		color.Fprint(cutypes.Output, question)
	}

	scanner := bufio.NewScanner(cutypes.Input)
	if !scanner.Scan() { // reading
		return "", scanner.Err()
	}

	return strings.TrimSpace(scanner.Text()), nil
}

// ReadLine read one line from user input.
//
// Usage:
//
//	in := ReadLine("")
//	ans, _ := ReadLine("your name?")
func ReadLine(question string) (string, error) {
	return internal.ReadLineWithOutput(question, cutypes.Output)
}

// ReadFirst read first char
func ReadFirst(question string) (string, error) {
	answer, err := ReadLine(question)
	if len(answer) == 0 {
		return "", err
	}

	return string(answer[0]), err
}

// AnswerIsYes check user inputted answer is right
//
// Usage:
//
//	fmt.Print("are you OK?")
//	ok := AnswerIsYes()
//	ok := AnswerIsYes(true)
func AnswerIsYes(defVal ...bool) bool {
	mark := " [yes|no]: "
	if len(defVal) > 0 {
		var defShow string
		if defVal[0] {
			defShow = "yes"
		} else {
			defShow = "no"
		}

		mark = fmt.Sprintf(" [yes|no](default <cyan>%s</>): ", defShow)
	}

	// _, err := fmt.Scanln(&answer)
	// _, err := fmt.Scan(&answer)
	fChar, err := ReadFirst(mark)
	if err != nil {
		panic(err)
	}

	if len(fChar) > 0 {
		fChar := strings.ToLower(fChar)
		if fChar == "y" {
			return true
		}
		if fChar == "n" {
			return false
		}
	} else if len(defVal) > 0 { // has default value
		return defVal[0]
	}

	_, _ = fmt.Fprint(cutypes.Output, "Please try again")
	return AnswerIsYes()
}

// ReadPassword from terminal
func ReadPassword(question ...string) string {
	fmt.Fprint(cutypes.Output, arrutil.FirstOr(question, "Enter Password: "))

	bs, err := term.ReadPassword(syscallStdin())
	if err != nil {
		return ""
	}

	fmt.Fprintln(cutypes.Output) // new line
	return string(bs)
}

// GetHiddenInput interactively prompts for input without echoing to the terminal.
//
// Usage:
//
//	// askPassword
//	pwd := GetHiddenInput("Enter Password:")
func GetHiddenInput(message string, trimmed bool) (string, error) {
	var err error
	var input string
	var hasResult bool

	// like *nix, git-bash ...
	if envutil.HasShellEnv("sh") {
		// COMMAND: sh -c 'read -p "Enter Password:" -s user_input && echo $user_input'
		cmd := fmt.Sprintf(`'read -p "%s" -s user_input && echo $user_input'`, message)
		input, err = cliutil.ShellExec(cmd)
		if err != nil {
			return "", err
		}

		fmt.Fprintln(cutypes.Output) // new line
		hasResult = true
	} else if envutil.IsWin() { // at windows cmd.exe
		// create a temp VB script file
		vbFile, err := os.CreateTemp("", "cliui-pwd")
		if err != nil {
			return "", err
		}
		defer func() {
			// delete file
			vbFile.Close()
			_ = os.Remove(vbFile.Name())
		}()

		script := fmt.Sprintf(`wscript.echo(InputBox("%s", "", "password here"))`, message)
		_, _ = vbFile.WriteString(script)
		hasResult = true

		// exec VB script
		// COMMAND: cscript //nologo vbFile.Name()
		input, err = cliutil.ExecCmd("cscript", []string{"//nologo", vbFile.Name()})
		if err != nil {
			return "", err
		}
	}

	if hasResult {
		if trimmed {
			return strings.TrimSpace(input), nil
		}
		return input, nil
	}
	return "", errors.New("current env is not support the method")
}
