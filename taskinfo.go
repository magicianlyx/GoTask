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
	Key        string     `json:"key"`         // 任务标志key
	task       TaskObj    `json:"-"`           // 任务方法
	LastTime   time.Time  `json:"last_time"`   // 最后一次执行任务的时间（未执行过时为time.Time{}）
	AddTime    time.Time  `json:"add_time"`    // 任务添加的时间
	Count      int        `json:"count"`       // 任务执行次数
	Spec       int        `json:"spec"`        // 任务执行时间间隔
	LastResult *TaskResult `json:"last_result"` // 任务最后一次执行的结果
}

// 生成副本
func (t *TaskInfo) clone() *TaskInfo {
	rt := &TaskInfo{}
	rt.Key = t.Key
	rt.task = t.task
	rt.LastTime = t.LastTime
	rt.AddTime = t.AddTime
	rt.Count = t.Count
	rt.Spec = t.Spec
	return rt
}

// 创建一个任务信息对象
func newTaskInfo(key string, task TaskObj, spec int) *TaskInfo {
	now := time.Now()
	return &TaskInfo{
		Key:      key,
		task:     task,
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
