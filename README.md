# gotaskv1

go定时任务工具库


### Api
```
// 创建一个定时任务对象
// @params: routineCount    执行任务的线程池大小
// @retuen: *TimedTask      创建的定时任务对象
func NewTimedTask(routineCount int) (*TimedTask)


// 添加定时任务
func (tt *TimedTask) Add(key string, task TaskObj, spec int)
// @params: key              任务键
// @params: task             任务方法
// @params: spec             任务定时时长


// 取消定时任务
func (tt *TimedTask) Cancel(key string)
// @params: key              任务键


// 获取定时任务列表信息
func (tt *TimedTask)GetTimedTaskInfo()map[string]*TaskInfo
// @retuen: map[string]*TaskInfo      定时任务列表

```

### Struct
```
// 执行任务回调函数参数结构体
type ExecuteCbArgs struct {
	Key      string                         任务键key
	Task     TaskObj                        任务方法
	LastTime time.Time                      最后一次执行任务的时间（未执行过时为time.Time{}）
	AddTime  time.Time                      任务添加的时间
	Count    int                            任务执行次数
	Spec     int                            任务执行时间间隔
	Res     map[string]interface{}          任务执行结果
	Error   error                           任务执行错误信息
}


// 添加任务回调函数参数结构体
type AddCbArgs struct {
	Key      string                         任务键key
	Task     TaskObj                        任务方法
	LastTime time.Time                      最后一次执行任务的时间（未执行过时为time.Time{}）
	AddTime  time.Time                      任务添加的时间
	Count    int                            任务执行次数
	Spec     int                            任务执行时间间隔
	Error error                             添加任务操作错误（如果任务键已经存在时会出错）
}


// 取消任务回调函数参数结构体
type CancelCbArgs struct {
	key   string                            任务键key
	Error error                             取消任务操作错误（如果任务键本身不存在时会出错）
}


// 定时任务信息
type TaskInfo struct {
	Key      string                         任务键key
	Task     TaskObj                        任务方法
	LastTime time.Time                      最后一次执行任务的时间（未执行过时为time.Time{}）
	AddTime  time.Time                      任务添加的时间
	Count    int                            任务执行次数
	Spec     int                            任务执行时间间隔
}


// 定时任务方法
type TaskObj func() (map[string]interface{}, error)


// 添加任务回调函数
type addCallback func(*AddCbArgs)


// 取消任务回调函数
type cancelCallback func(*CancelCbArgs)


// 执行任务回调函数
type executeCallback func(*ExecuteCbArgs)
```

### Demo
```
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
		return nil, nil
	}, 2)

	// 添加任务B
	tt.Add("B", func() (map[string]interface{}, error) {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "Execute Task `B`")
		return nil, nil
	}, 1)

	// 添加任务C
	tt.Add("C", func() (map[string]interface{}, error) {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "Execute Task `C`")
		return nil, nil
	}, 3)

	// 打印时间
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "[Init]")

	// 6秒后取消A
	time.Sleep(time.Second * 6)
	tt.Cancel("A")

	// 6秒后取消B
	time.Sleep(time.Second * 10)
	tt.Cancel("B")

	time.Sleep(time.Hour)
```
