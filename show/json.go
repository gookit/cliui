package show

import "github.com/gookit/cliui/show/showcom"

// PrettyJSON struct
type PrettyJSON struct {
	showcom.Base
}

// NewPrettyJSON create an instance.
func NewPrettyJSON() *PrettyJSON {
	return &PrettyJSON{}
}
