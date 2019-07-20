package GoTaskv1

import (
	"testing"
	"fmt"
	"time"
)

func TestMultiTask(t *testing.T) {
	mt := NewMultiTask(10)
	mt.AddExecuteCallback(func(args *ExecuteCbArgs) {
		fmt.Println("执行了一次", args.Key, time.Now().Format("2006-01-02 15:04:05"))
		if args.Key == "A" && args.Count == 5 {
			mt.Add("B", func() (map[string]interface{}, error) {
				fmt.Println("B")
				return nil, nil
			}, 2, 5)
		}
	})
	mt.Add("A", func() (map[string]interface{}, error) {
		fmt.Println("A")
		return nil, nil
	}, 2, 5)
	
	time.Sleep(time.Hour)
}

func TestSingleTask(t *testing.T) {
	mt := NewSingleTask(10)
	mt.Add("key", func() (map[string]interface{}, error) {
		fmt.Println("F")
		return nil, nil
	}, 2)
	time.Sleep(time.Hour)
}

func TestNewTimedTask(t *testing.T) {
	for i := 0; i < 100000; i++ {
		go func() {
			a := NewTimedTask(3)
			a.Add("A", func() (map[string]interface{}, error) {
				fmt.Println("执行任务A")
				return nil, nil
			}, 1)
			time.Sleep(3 * time.Second)
			a.Stop()
			fmt.Println("成功停止")
		}()
	}
	
	time.Sleep(time.Hour)
}
