package alert_test

import (
	"bytes"
	"testing"

	"github.com/gookit/cliui/cutypes"
	"github.com/gookit/cliui/show/alert"
	"github.com/gookit/goutil/testutil/assert"
)

func TestErrorPrintsBanner(t *testing.T) {
	is := assert.New(t)

	buf := new(bytes.Buffer)
	cutypes.SetOutput(buf)
	defer cutypes.ResetOutput()

	code := alert.Error("failed: %s", "network")
	out := buf.String()

	is.Eq(1, code)
	is.Contains(out, "ERROR: failed: network")
	is.Contains(out, "╭")
	is.Contains(out, "╰")
}

func TestSuccessPrintsBanner(t *testing.T) {
	is := assert.New(t)

	buf := new(bytes.Buffer)
	cutypes.SetOutput(buf)
	defer cutypes.ResetOutput()

	code := alert.Success("created %s", "user")
	out := buf.String()

	is.Eq(0, code)
	is.Contains(out, "SUCCESS: created user")
	is.Contains(out, "╭")
	is.Contains(out, "╰")
}

func TestInfoAndWarningPrintBanner(t *testing.T) {
	is := assert.New(t)

	buf := new(bytes.Buffer)
	cutypes.SetOutput(buf)
	defer cutypes.ResetOutput()

	is.Eq(0, alert.Info("ready"))
	is.Eq(0, alert.Warning("check config"))

	out := buf.String()
	is.Contains(out, "INFO: ready")
	is.Contains(out, "WARNING: check config")
	is.Contains(out, "╭")
	is.Contains(out, "╰")
}
