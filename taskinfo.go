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
	Key        string      `json:"key"`         // 任务标志key
	task       TaskObj     `json:"-"`           // 任务方法
	lastTime   time.Time   `json:"last_time"`   // 最后一次执行任务的时间（未执行过时为time.Time{}）
	addTime    time.Time   `json:"add_time"`    // 任务添加的时间
	count      int         `json:"count"`       // 任务执行次数
	spec       int         `json:"spec"`        // 任务执行时间间隔
	LastResult *TaskResult `json:"last_result"` // 任务最后一次执行的结果
}

// 生成副本
func (t *TaskInfo) clone() *TaskInfo {
	rt := &TaskInfo{}
	rt.Key = t.Key
	rt.task = t.task
	rt.lastTime = t.lastTime
	rt.addTime = t.addTime
	rt.count = t.count
	rt.spec = t.spec
	return rt
}

// 任务添加时间
func (t *TaskInfo) AddTaskTime() time.Time {
	return t.addTime
}

// 第一次执行执行时间
func (t *TaskInfo) FirstExecuteTime() (time.Time, bool) {
	if t.lastTime.IsZero() {
		return time.Time{}, false
	} else {
		return t.addTime.Add(time.Duration(t.spec) * time.Second), true
	}
}

// 最后一次执行时间
func (t *TaskInfo) LastExecuteTime() (time.Time, bool) {
	if t.lastTime.IsZero() {
		return time.Time{}, false
	} else {
		return t.lastTime, true
	}
}

// 下次执行时间
func (t *TaskInfo) NextScheduleTime() time.Time {
	lastTime := t.lastTime
	if lastTime.IsZero() {
		lastTime = t.addTime
	}
	return lastTime.Add(time.Duration(t.spec) * time.Second)
}


func (t *TaskInfo) UpdateAfterExecute() {
	t.count += 1
	if t.lastTime.IsZero() {
		t.lastTime = t.addTime.Add(time.Duration(t.spec) * time.Second)
	} else {
		t.lastTime = t.lastTime.Add(time.Duration(t.spec) * time.Second)
	}
}

// 创建一个任务信息对象
func NewTaskInfo(key string, task TaskObj, spec int) *TaskInfo {
	now := time.Now()
	return &TaskInfo{
		Key:      key,
		task:     task,
		lastTime: time.Time{},
		addTime:  now,
		count:    0,
		spec:     spec,
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
