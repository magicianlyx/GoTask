package pool

import (
	"fmt"
	"github.com/magicianlyx/GoTask/utils"
	"testing"
)

func TestNewDefaultOptions(t *testing.T) {
	o := &Options{}
	o.fillDefaultOptions()
	fmt.Printf("%v\r\n", utils.ToJson(o))
}

func TestOptions_Clone(t *testing.T) {
	o := &Options{}
	o.fillDefaultOptions()
	fmt.Printf("%v\r\n", utils.ToJson(o))
	oc := o.Clone()
	oc.TaskChannelSize = 10
	fmt.Printf("%v\r\n", utils.ToJson(o))
	fmt.Printf("%v\r\n", utils.ToJson(oc))
}
