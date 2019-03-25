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
	
	tt.Add("B", func() (map[string]interface{}, error) {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "BBBBB")
		return nil, nil
	}, 3)
	
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
	
	// 让C任务执行10次后取消
	tt.AddExecuteCallback(func(args *ExecuteCbArgs) {
		if args.Key == "C" && args.Count == 10 {
			tt.Cancel("C")
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
	
	// 添加任务A
	tt.Add("A", func() (map[string]interface{}, error) {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "Execute Task `A`")
		return nil, nil
	}, 2)
	
	// 重复添加任务A
	tt.Add("A", func() (map[string]interface{}, error) {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "Execute Task `A`")
		time.Sleep(time.Second)
		return nil, nil
	}, 2)
	
	// 添加任务B
	tt.Add("B", func() (map[string]interface{}, error) {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "Execute Task `B`")
		time.Sleep(time.Second)
		return nil, nil
	}, 1)
	
	// 添加任务C
	tt.Add("C", func() (map[string]interface{}, error) {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "Execute Task `C`")
		time.Sleep(time.Second)
		return nil, nil
	}, 3)
	
	// 打印时间
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "[Init]")
	
	// 6秒后取消A
	time.Sleep(time.Second * 4)
	tt.Cancel("A")
	
	// 6秒后取消B
	time.Sleep(time.Second * 4)
	tt.Cancel("B")
	
	fmt.Printf("%s\r\n", ToJson(tt.GetAllGoroutineStatus()))
	time.Sleep(time.Hour)
	
}

func TestBan(t *testing.T) {
	tt := NewTimedTask(10)
	
	tt.AddExecuteCallback(func(args *ExecuteCbArgs) {
		
		if args.Error != nil {
			fmt.Println(fmt.Sprintf("exec task: %s ,error msg: %s", args.Key, args.Error))
		} else {
			fmt.Println(fmt.Sprintf("exec task: %s  ,  %s", args.Key, args.LastTime.Format("2006-01-02 15:04:05")))
		}
	})
	
	// 添加添加任务回调
	tt.AddAddCallback(func(args *AddCbArgs) {
		if args.Error != nil {
			fmt.Println(fmt.Sprintf("add task: %s ,error msg: %s", args.Key, args.Error))
		} else {
			fmt.Println(fmt.Sprintf("add task: %s", args.Key))
		}
	})
	
	tt.AddBanCallback(func(args *BanCbArgs) {
		if args.Error != nil {
			fmt.Println(fmt.Sprintf("add ban task: %s ,error msg: %s", args.Key, args.Error))
		} else {
			fmt.Println(fmt.Sprintf("add ban task: %s", args.Key))
		}
	})
	
	key := "AAA"
	tt.Ban(key)
	
	tt.Add(key, func() (map[string]interface{}, error) {
		return nil, nil
	}, 2)
	
	tt.UnBan(key)
	
	tt.Add(key, func() (map[string]interface{}, error) {
		return nil, nil
	}, 2)
	
	tt.tMap.tMap.Range(func(key, value interface{}) bool {
		fmt.Println(key)
		return true
	})
	
	time.Sleep(10 * time.Second)
	tt.Ban(key)
	
	
	time.Sleep(time.Hour)
}

func TestAddCallBack(t *testing.T) {
	
	tt := NewTimedTask(10)
	tt.AddAddCallback(func(args *AddCbArgs) {
		fmt.Println(fmt.Sprintf("add ban task: %s", args.Key))
	})
	
	tt.Add("A", func() (map[string]interface{}, error) {
		return nil, nil
	}, 1)
	
	tt.Add("B", func() (map[string]interface{}, error) {
		return nil, nil
	}, 1)
	
	time.Sleep(time.Hour)
	
}
