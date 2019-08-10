package GoTaskv1

import (
	"time"
)

type TaskObj func() (map[string]interface{}, error)

type TaskResult struct {
	Result map[string]interface{}
	Err    error
}

type TaskInfo struct {
	Key        string      // 任务标志key
	Task       TaskObj     // 任务方法
	LastTime   time.Time   // 最后一次执行任务的时间（未执行过时为time.Time{}）
	AddTime    time.Time   // 任务添加的时间
	Count      int         // 任务执行次数
	Spec       int         // 任务执行时间间隔
	LastResult *TaskResult // 任务最后一次执行的结果
}

// 生成副本
func (t *TaskInfo) clone() *TaskInfo {
	rt := &TaskInfo{}
	rt.Key = t.Key
	rt.Task = t.Task
	rt.LastTime = t.LastTime
	rt.AddTime = t.AddTime
	rt.Count = t.Count
	rt.Spec = t.Spec
	return rt
}

// 任务添加时间
func (t *TaskInfo) GetAddTaskTime() time.Time {
	return t.AddTime
}

// 第一次执行执行时间
func (t *TaskInfo) GetFirstExecuteTime() (time.Time, bool) {
	if t.LastTime.IsZero() {
		return time.Time{}, false
	} else {
		return t.AddTime.Add(time.Duration(t.Spec) * time.Second), true
	}
}

// 最后一次执行时间
func (t *TaskInfo) GetLastExecuteTime() (time.Time, bool) {
	if t.LastTime.IsZero() {
		return time.Time{}, false
	} else {
		return t.LastTime, true
	}
}

// 下次执行时间
func (t *TaskInfo) NextScheduleTime() (time.Time, bool) {
	lastTime := t.LastTime
	if lastTime.IsZero() {
		lastTime = t.AddTime
	}
	return lastTime.Add(time.Duration(t.Spec) * time.Second), true
}

// 执行或调用 调整任务信息
func (t *TaskInfo) UpdateAfterExecute() {
	t.Count += 1
	if t.LastTime.IsZero() {
		t.LastTime = t.AddTime.Add(time.Duration(t.Spec) * time.Second)
	} else {
		t.LastTime = t.LastTime.Add(time.Duration(t.Spec) * time.Second)
	}
}

// 创建一个任务信息对象
func NewTaskInfo(key string, task TaskObj, spec int) *TaskInfo {
	now := time.Now()
	return &TaskInfo{
		Key:      key,
		Task:     task,
		LastTime: time.Time{},
		AddTime:  now,
		Count:    0,
		Spec:     spec,
	}
}

type ExecuteCbArgs struct {
	*TaskInfo
	Res       map[string]interface{}
	Error     error
	RoutineId int
}

type AddCbArgs struct {
	*TaskInfo
	Error error
}

type CancelCbArgs struct {
	Key   string
	Error error
}

type BanCbArgs struct {
	Key   string
	Error error
}
type UnBanCbArgs BanCbArgs
