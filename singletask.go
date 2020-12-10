package GoTask

import (
	"gitee.com/magicianlyx/GoTask/task"
	"time"
)

// 指定执行次数定时任务
type MultiTask struct {
	*TimedTask
}

func NewMultiTask(routineCount int) *MultiTask {
	return &MultiTask{NewTimedTask(routineCount)}
}

func (mt *MultiTask) Add(key string, t task.TaskObj, spec int, count int) {
	mt.TimedTask.Add(key, t, task.NewSpecTimeSchedule(time.Duration(spec)*time.Second, count))
}

// 指定只执行一次定时任务
type SingleTask struct {
	*MultiTask
}

func NewSingleTask(routineCount int) *SingleTask {
	return &SingleTask{NewMultiTask(routineCount)}
}

func (st *SingleTask) Add(key string, task task.TaskObj, spec int) {
	st.MultiTask.Add(key, task, spec, 1)
}
