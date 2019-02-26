package GoTaskv1

import (
	"testing"
	"fmt"
)

func TestFuncMap(t *testing.T) {
	fm := newFuncMap()
	
	banCb := []banCallback{}
	
	f := func(args *BanCbArgs) {
		fmt.Println("A")
	}
	fm.add(f)
	f1 := func(args *BanCbArgs) {
		fmt.Println("B")
	}
	fm.add(f1)
	
	n := fm.getAll(&banCb)
	for _, cb := range banCb {
		cb(nil)
	}
	fmt.Println(n)
	
}
