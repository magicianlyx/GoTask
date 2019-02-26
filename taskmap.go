package GoTaskv1

import (
	"sync"
	"time"
)

// 任务字典 线程安全
type taskMap struct {
	tMap         sync.Map
	tasks        chan *TaskInfo
}

func newtaskMap() *taskMap {
	return &taskMap{
		tMap: sync.Map{},
	}
}

func (tm *taskMap) add(key string, task *TaskInfo) {
	if !tm.isExist(key) {
		tm.tMap.Store(key, task)
	}
}

func (tm *taskMap) set(key string, task *TaskInfo) {
	if tm.isExist(key) {
		tm.tMap.Store(key, task)
	}
}

func (tm *taskMap) addOrSet(key string, task *TaskInfo) {
	tm.tMap.Store(key, task)
}

func (tm *taskMap) delete(key string) {
	tm.tMap.Delete(key)
}

func (tm *taskMap) get(key string) *TaskInfo {
	if v, ok := tm.tMap.Load(key); ok {
		if v1, ok1 := v.(*TaskInfo); ok1 {
			return v1.clone()
		}
	}
	return nil
}

// 键是否存在
func (tm *taskMap) isExist(key string) bool {
	_, ok := tm.tMap.Load(key)
	return ok
}

// 选择下一个最早执行的任务
func (tm *taskMap) selectNextExec() *TaskInfo {
	var minv *TaskInfo
	tm.tMap.Range(func(key, value interface{}) bool {
		v, ok := value.(*TaskInfo)
		if !ok {
			return true
		}
		if minv == nil {
			minv = v
			return true
		}
		
		var vlt time.Time
		if v.LastTime.IsZero() {
			vlt = v.AddTime
		} else {
			vlt = v.LastTime
		}
		
		var mlt time.Time
		if minv.LastTime.IsZero() {
			mlt = minv.AddTime
		} else {
			mlt = minv.LastTime
		}
		
		if mlt.Add(time.Duration(minv.Spec) * time.Second).UnixNano() > vlt.Add(time.Duration(v.Spec) * time.Second).UnixNano() {
			minv = v
		}
		return true
	})
	if minv == nil {
		return nil
	}
	return minv.clone()
}

func (tm *taskMap) getAll() map[string]*TaskInfo {
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
		m[k] = v.clone()
		return true
	})
	return m
}
