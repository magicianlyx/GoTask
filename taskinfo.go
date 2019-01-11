package GoTaskv1

import (
	"time"
	"reflect"
)

type TaskObj func() (map[string]interface{}, error)

type TaskInfo struct {
	Key      string    // 任务标志key
	Task     TaskObj   // 任务
	LastTime time.Time // 最后一次执行任务的时间
	AddTime  time.Time // 任务添加的时间
	Count    int       // 任务执行次数
	Spec     int       // 任务执行时间间隔
}

// 生成副本
func (t *TaskInfo) Clone() *TaskInfo {
	rt := &TaskInfo{}
	tv := reflect.ValueOf(t).Elem()
	rtv := reflect.ValueOf(rt).Elem()
	for i := 0; i < tv.NumField(); i++ {
		rtv.Field(i).Set(tv.Field(i))
	}
	return rt
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
