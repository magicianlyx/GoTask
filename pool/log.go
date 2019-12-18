package pool

import (
	"fmt"
	"io"
)

var o io.Writer

func SetLog(w io.Writer) {
	o = w
}

func printf(format string, vals ...interface{}) {
	if o == nil {
		return
		// o = os.Stdout
	}
	_, _ = o.Write([]byte(fmt.Sprintf(format, vals...) + "\r\n"))
}
