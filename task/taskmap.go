package task

import (
	"sync"
	"time"
)

// 任务字典 线程安全
type TaskMap struct {
	tMap sync.Map
}

func NewTaskMap() *TaskMap {
	return &TaskMap{
		tMap: sync.Map{},
	}
}

// 添加
func (tm *TaskMap) Add(key string, task *TaskInfo) {
	if !tm.IsExist(key) {
		tm.tMap.Store(key, task)
	}
}

// 存在时才修改
func (tm *TaskMap) Set(key string, task *TaskInfo) {
	if tm.IsExist(key) {
		tm.tMap.Store(key, task)
	}
}

// 添加或修改
func (tm *TaskMap) AddOrSet(key string, task *TaskInfo) {
	tm.tMap.Store(key, task)
}

// 删除
func (tm *TaskMap) Delete(key string) {
	tm.tMap.Delete(key)
}

// 获取 返回副本
func (tm *TaskMap) Get(key string) *TaskInfo {
	if v, ok := tm.tMap.Load(key); ok {
		if v1, ok1 := v.(*TaskInfo); ok1 {
			return v1.Clone()
		}
	}
	return nil
}

// 键是否存在
func (tm *TaskMap) IsExist(key string) bool {
	_, ok := tm.tMap.Load(key)
	return ok
}

// 选择下一个最早执行的任务
func (tm *TaskMap) SelectNextExec() (*TaskInfo, time.Duration, bool) {
	var minv *TaskInfo
	tm.tMap.Range(func(key, value interface{}) bool {
		v, ok := value.(*TaskInfo)
		if !ok {
			return true
		}
		if !v.HasNext {
			return true
		}
		if minv == nil {
			minv = v
			return true
		}
		mnt := minv.NextScheduleTime()
		vnt := v.NextScheduleTime()

		if mnt.UnixNano() > vnt.UnixNano() {
			minv = v
		}
		return true
	})
	if minv == nil {
		return nil, 0, false
	}
	ns := minv.NextScheduleTime()
	spec := ns.Sub(time.Now())
	if spec <= 0 {
		spec = time.Nanosecond
	}
	return minv, spec, true
}

// 获取所有返回副本
func (tm *TaskMap) GetAll() map[string]*TaskInfo {
	m := map[string]*TaskInfo{}
	tm.tMap.Range(func(key, value interface{}) bool {
		k, ok := key.(string)
		if !ok {
			return true
		}
		v, ok := value.(*TaskInfo)
		if !ok {
			return true
		}
		m[k] = v.Clone()
		return true
	})
	return m
}
