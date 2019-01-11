package GoTaskv1

import (
	"sync"
	"errors"
	"time"
)

var (
	ErrTaskIsExist    = errors.New("task is exist")
	ErrTaskIsNotExist = errors.New("task is not exist")
)

type TimedTask struct {
	l            sync.RWMutex
	tMap         *TaskMap
	tasks        chan *TaskInfo
	routineCount int
}

func NewTimedTask() (*TimedTask) {
	tt := &TimedTask{
		tMap: NewTaskMap(),
	}
	return tt
}

func (tt *TimedTask) add(key string, task TaskObj, spec int) error {
	if tt.tMap.IsExist(key) {
		return ErrTaskIsExist
	}
	tt.tMap.Add(key, NewTaskInfo(key, task, spec))
	tt.reSelectAfterUpdate()
	return nil
}

func (tt *TimedTask) Add(key string, task TaskObj, spec int) {
	tt.l.Lock()
	tt.add(key, task, spec)
	tt.l.Unlock()
}

func (tt *TimedTask) cancel(key string) error {
	if !tt.tMap.IsExist(key) {
		return ErrTaskIsNotExist
	}
	tt.tMap.Delete(key)
	tt.reSelectAfterUpdate()
	return nil
}

func (tt *TimedTask) Cancel(key string) {
	tt.l.Lock()
	tt.cancel(key)
	tt.l.Unlock()
}

func (tt *TimedTask) goExecutor(routineCount int) {
	for i := 0; i < routineCount; i++ {
		go func(rid int) {
			ti := <-tt.tasks
			res, err := ti.Task()
			_ = res
			_ = err
			tt.l.Lock()
			tt.updateMapAfterExec(ti)
			tt.reSelectAfterUpdate()
			tt.l.Unlock()
		}(i)
	}
}

func (tt *TimedTask)goTimedTask(){
	go func() {
	
	}()
}

func (tt *TimedTask) updateMapAfterExec(task *TaskInfo) {
	task.Count += 1
	task.LastTime = time.Now()
	tt.tMap.Set(task.Key, task)
}

// 将下一个最早被执行的任务推入定时任务
func (tt *TimedTask) reSelectAfterUpdate() {
}
