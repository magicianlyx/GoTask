package task

import (
	"testing"
	"fmt"
	"time"
)

func TestA(t *testing.T) {
	rt := InitTaskRoutine()
	rt.AddExecuteCallBack(PrintLog)
	rt.Add("A", func() (map[string]interface{}, error) {
		PrintMsg("A")
		return nil, nil
	}, 3)
	time.Sleep(10 * time.Second)
	//rt.Cancel("A")
	rt.Restart("A")
	time.Sleep(time.Hour)
}

func PrintMsg(msg string) {
	fmt.Printf("%s\r\n", msg)
}

func PrintLog(p ExecuteArgs) {
	fmt.Println(p.Key, "   ", p.LastExecuteTime, "  ", p.AddTime, "   ", p.Time)
}
