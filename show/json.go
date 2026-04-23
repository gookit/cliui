package show

import "github.com/gookit/cliui/show/showcom"

// PrettyJSON struct
type PrettyJSON struct {
	showcom.Base
}

// NewPrettyJSON instance
func NewPrettyJSON() *PrettyJSON {
	return &PrettyJSON{}
}
