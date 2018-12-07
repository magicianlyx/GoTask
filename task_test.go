package task

import (
	"testing"
	"fmt"
	"time"
)

func TestA(t *testing.T) {
	rt := InitTaskRoutine()
	rt.Add("A", func() (map[string]interface{}, error) {
		PrintMsg("A")
		return nil,nil
	}, 3)
	// rt.Add("B", func() {
	// 	PrintMsg("B")
	// }, 3)
	rt.AddExecuteCallBack(PrintLog)
	time.Sleep(time.Hour)
}

func PrintMsg(msg string) {
	fmt.Printf("%s\r\n", msg)
}

func PrintLog(p ExecuteArgs) {
	fmt.Println(p.Key, "   ", p.LastExecuteTime, "  ", p.AddTime)
}
