package GoTaskv1

import (
	"fmt"
	"github.com/json-iterator/go"
	"testing"
)

// 测试使用
func ToJson(v interface{}) string {
	bs, err := jsoniter.MarshalIndent(v, "", "   ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(bs)
}

func TestA(t *testing.T) {
}
