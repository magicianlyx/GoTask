package task

import (
	"time"
)

type TaskObj func() (map[string]interface{}, error)

type TaskResult struct {
	Result map[string]interface{}
	Err    error
}

func (tr *TaskResult) Clone() *TaskResult {
	if tr == nil {
		return nil
	}
	return &TaskResult{
		Result: tr.Result,
		Err:    tr.Err,
	}
}

type TaskInfo struct {
	Key        string      // 任务标志key
	Task       TaskObj     // 任务方法
	LastTime   time.Time   // 最后一次执行任务的时间（未执行过时为time.Time{}）
	AddTime    time.Time   // 任务添加的时间
	NextTime   time.Time   // 下次执行时间
	Count      int         // 任务执行次数
	Sche       ISchedule   // 任务计划
	HasNext    bool        // 是否还有下一次执行
	LastResult *TaskResult // 任务最后一次执行的结果
	timer      TimerObj    // 计时器
}

// // 开启计时器
// func (t *TaskInfo) TimerStart() {
// 	t.timer = Timer()
// }
//
// // 结算计时器
// func (t *TaskInfo) TimerSettle() (spec time.Duration, start, end time.Time) {
// 	if t.timer != nil {
// 		return t.timer()
// 	} else {
// 		return 0, time.Time{}, time.Time{}
// 	}
// }

// 生成副本
func (t *TaskInfo) Clone() *TaskInfo {
	if t == nil {
		return nil
	}
	rt := &TaskInfo{}
	rt.Key = t.Key
	rt.Task = t.Task
	rt.LastTime = t.LastTime
	rt.AddTime = t.AddTime
	rt.NextTime = t.NextTime
	rt.Count = t.Count
	rt.Sche = t.Sche
	rt.HasNext = t.HasNext
	rt.LastResult = t.LastResult.Clone()
	return rt
}

// 任务添加时间
func (t *TaskInfo) GetAddTaskTime() time.Time {
	return t.AddTime
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
func (t *TaskInfo) NextScheduleTime() time.Time {
	return t.NextTime
}

// 执行后调用 调整任务信息
// 返回是否还有下一次执行
func (t *TaskInfo) Update() {
	t.Count += 1
	t.LastTime = time.Now()
	t.NextTime, t.HasNext = t.Sche.Expression(t)
}

// 是否还有下一次执行
func (t *TaskInfo) HasNextExecute() bool {
	return t.HasNext
}

// 创建一个任务信息对象
func NewTaskInfo(key string, task TaskObj, sche ISchedule) *TaskInfo {
	now := time.Now()
	ti := &TaskInfo{
		Key:      key,
		Task:     task,
		LastTime: time.Time{},
		AddTime:  now,
		Count:    0,
		Sche:     sche,
	}
	ti.NextTime, ti.HasNext = sche.Expression(ti)
	return ti
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
