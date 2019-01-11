package GoTaskv1

import (
	"testing"
	"fmt"
	"github.com/json-iterator/go"
)

func TestClone(*testing.T) {
	t := &TaskInfo{
		Key: "123",
	}
	tc := t.Clone()
	fmt.Printf("%v\r\n", ToJson(tc))
}

// 测试使用
func ToJson(v interface{}) string {
	bs, err := jsoniter.MarshalIndent(v, "", "   ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(bs)
}
