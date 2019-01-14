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
	tc := t.clone()
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
	
	tt.AddExecuteCallback(func(args *ExecuteCbArgs) {
		if args.Key == "B" && args.Count >= 3 {
			tt.Cancel("B")
		}
	})
	
	tt.AddCancelCallback(func(args *CancelCbArgs) {
		if args.Error == nil {
			fmt.Println("cancel timed task: ", args.Key)
		} else {
			fmt.Println(fmt.Sprintf("cancel timed task:%s ,error msg: %s", args.Key, args.Error))
		}
	})
	
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "[Init]")
	time.Sleep(time.Hour)
}

func TestReAdd(t *testing.T) {
	tt := NewTimedTask(10)
	
	// 添加添加任务回调
	tt.AddAddCallback(func(args *AddCbArgs) {
		if args.Error != nil {
			fmt.Println(fmt.Sprintf("add task: %s ,error msg: %s", args.Key, args.Error))
		} else {
			fmt.Println(fmt.Sprintf("add task: %s", args.Key))
		}
	})
	
	// 添加取消任务回调
	tt.AddCancelCallback(func(args *CancelCbArgs) {
		if args.Error == nil {
			fmt.Println("cancel timed task: ", args.Key)
		} else {
			fmt.Println(fmt.Sprintf("cancel timed task:%s ,error msg: %s", args.Key, args.Error))
		}
	})
	
	// 添加任务B
	tt.Add("B", func() (map[string]interface{}, error) {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "Execute Task `B`")
		return nil, nil
	}, 1)
	
	// 添加任务A
	tt.Add("A", func() (map[string]interface{}, error) {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "Execute Task `A`")
		return nil, nil
	}, 2)
	
	// 重复添加任务A
	tt.Add("A", func() (map[string]interface{}, error) {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "Execute Task `A`")
		return nil, nil
	}, 2)
	
	// 打印时间
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "[Init]")
	
	// 6秒后取消A
	time.Sleep(time.Second * 6)
	tt.Cancel("A")
	
	// 6秒后取消B
	time.Sleep(time.Second * 10)
	tt.Cancel("B")
	
}
