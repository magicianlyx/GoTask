package GoTaskv1

// 指定执行次数定时任务
type MultiTask struct {
	*TimedTask
}

func NewMultiTask(routineCount int) *MultiTask {
	return &MultiTask{NewTimedTask(routineCount)}
}

func (mt *MultiTask) Add(key string, task TaskObj, spec int, count int) {
	var f executeCallback
	f = func(args *ExecuteCbArgs) {
		if args.Key == key && args.Count >= count {
			mt.cancelWithCb(key, false)
			// delete this callback function after cancel
			mt.DelExecuteCallback(f)
		}
	}

	mt.AddExecuteCallback(f)
	mt.TimedTask.Add(key, task, spec)
}

// 指定只执行一次定时任务
type SingleTask struct {
	*MultiTask
}

func NewSingleTask(routineCount int) *SingleTask {
	return &SingleTask{NewMultiTask(routineCount)}
}

func (st *SingleTask) Add(key string, task TaskObj, spec int) {
	st.MultiTask.Add(key, task, spec, 1)
}
