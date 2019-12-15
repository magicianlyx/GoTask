package pool

import (
	"fmt"
	"io"
	"os"
)

var o io.Writer

func SetLog(w io.Writer) {
	o = w
}

func printf(format string, vals ...interface{}) {
	if o == nil {
		o = os.Stdout
	}
	_, _ = o.Write([]byte(fmt.Sprintf(format, vals...) + "\r\n"))
}
