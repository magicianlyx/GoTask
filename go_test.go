package GoTaskv1

import (
	"testing"
	"fmt"
	"time"
	"github.com/json-iterator/go"
)

// 测试使用
func ToJson(v interface{}) string {
	bs, err := jsoniter.MarshalIndent(v, "", "   ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(bs)
}

func TestClone(*testing.T) {
	t := &TaskInfo{
		Key: "123",
	}
	tc := t.Clone()
	fmt.Printf("%v\r\n", ToJson(tc))
}

func TestTimedTask(t *testing.T) {
	tt := NewTimedTask(10)
	tt.Add("B", func() (map[string]interface{}, error) {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "BBBBB")
		return nil, nil
	}, 3)
	
	tt.Add("A", func() (map[string]interface{}, error) {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "AAAAA")
		return nil, nil
	}, 2)
	
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "[Init]")
	
	time.Sleep(time.Second * 6)
	tt.Cancel("A")
	fmt.Println("取消A")
	fmt.Println(tt.GetTimedTaskInfo())
	time.Sleep(time.Second * 6)
	tt.Cancel("B")
	fmt.Println("取消B")
	time.Sleep(time.Second * 6)
	tt.Add("A", func() (map[string]interface{}, error) {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "AAAAA")
		return nil, nil
	}, 2)
	fmt.Println("新增A")
	time.Sleep(time.Hour)
	
}

func TestRecursionCall(t *testing.T) {
	tt := NewTimedTask(10)
	
	tt.Add("B", func() (map[string]interface{}, error) {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "BBBBB")
		return nil, nil
	}, 3)
	
	tt.AddExecuteCallback(func(info *TaskInfo, res map[string]interface{}, err error) {
		if info.Key == "B" && info.Count >= 3 {
			tt.Cancel("B")
		}
	})
	
	tt.AddCancelCallback(func(key string, err error) {
		fmt.Println("cancel timed task: ", key)
	})
	
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "[Init]")
	time.Sleep(time.Hour)
}

func TestReAdd(t *testing.T) {
	tt := NewTimedTask(10)
	
	tt.AddAddCallback(func(info *TaskInfo, err error) {
		if err != nil {
			fmt.Println(fmt.Sprintf("add task: %s ,error msg: %s", info.Key, err))
		} else {
			fmt.Println(fmt.Sprintf("add task: %s", info.Key))
		}
	})
	tt.Add("B", func() (map[string]interface{}, error) {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "BBBBB")
		return nil, nil
	}, 3)
	tt.Add("B", func() (map[string]interface{}, error) {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "BBBBB")
		return nil, nil
	}, 3)
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "[Init]")
	time.Sleep(time.Hour)
}
