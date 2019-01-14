package GoTaskv1

import (
	"sync"
	"time"
)

type TaskMap struct {
	tMap         sync.Map
	tasks        chan *TaskInfo
	routineCount int
}

func NewTaskMap() *TaskMap {
	return &TaskMap{
		tMap: sync.Map{},
	}
}

func (tm *TaskMap) Add(key string, task *TaskInfo) {
	if !tm.IsExist(key) {
		tm.tMap.Store(key, task)
	}
}

func (tm *TaskMap) Set(key string, task *TaskInfo) {
	if tm.IsExist(key) {
		tm.tMap.Store(key, task)
	}
}

func (tm *TaskMap) AddOrSet(key string, task *TaskInfo) {
	tm.tMap.Store(key, task)
}

func (tm *TaskMap) Delete(key string) {
	tm.tMap.Delete(key)
}

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
func (tm *TaskMap) SelectNextExec() *TaskInfo {
	var minv *TaskInfo
	tm.tMap.Range(func(key, value interface{}) bool {
		v, ok := value.(*TaskInfo)
		if !ok {
			return true
		}
		if minv == nil || minv.LastTime.Add(time.Duration(minv.Spec) * time.Second).UnixNano() > v.LastTime.Add(time.Duration(v.Spec) * time.Second).UnixNano() {
			minv = v
		}
		return true
	})
	if minv == nil {
		return nil
	}
	return minv.Clone()
}

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
		m[k] = v
		return true
	})
	return m
}
