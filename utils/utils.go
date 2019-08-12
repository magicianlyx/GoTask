package utils

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
)

// 测试使用
func ToJson(v interface{}) string {
	bs, err := jsoniter.MarshalIndent(v, "", "   ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(bs)
}
