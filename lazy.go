package lazyflag

import (
	"os"
)

var Default = NewLoader()

func init() {
	Default.Parse(os.Args[1:])
}
