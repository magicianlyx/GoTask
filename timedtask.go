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

type AddCallback func(info *TaskInfo, err error)
type CancelCallback func(key string, err error)
type ExecuteCallback func(info *TaskInfo, res map[string]interface{}, err error)

type TimedTask struct {
	l               sync.RWMutex
	tMap            *TaskMap
	tasks           chan *TaskInfo
	refreshSign     chan struct{}
	routineCount    int
	addCallback     []AddCallback
	cancelCallback  []CancelCallback
	executeCallback []ExecuteCallback
}

func NewTimedTask(routineCount int) (*TimedTask) {
	tt := &TimedTask{
		sync.RWMutex{},
		NewTaskMap(),
		make(chan *TaskInfo),
		make(chan struct{}),
		routineCount,
		[]AddCallback{},
		[]CancelCallback{},
		[]ExecuteCallback{},
	}
	tt.goExecutor()
	tt.goTimedIssue()
	return tt
}

func (tt *TimedTask) AddAddCallback(cb AddCallback) {
	tt.addCallback = append(tt.addCallback, cb)
}
func (tt *TimedTask) AddCancelCallback(cb CancelCallback) {
	tt.cancelCallback = append(tt.cancelCallback, cb)
}
func (tt *TimedTask) AddExecuteCallback(cb ExecuteCallback) {
	tt.executeCallback = append(tt.executeCallback, cb)
}

func (tt *TimedTask) invokeAddCallback(info *TaskInfo, err error) {
	go func() {
		for _, cb := range tt.addCallback {
			cb(info, err)
		}
	}()
}

func (tt *TimedTask) invokeCancelCallback(key string, err error) {
	go func() {
		for _, cb := range tt.cancelCallback {
			cb(key, err)
		}
	}()
}

func (tt *TimedTask) invokeExecuteCallback(info *TaskInfo, res map[string]interface{}, err error) {
	go func() {
		for _, cb := range tt.executeCallback {
			cb(info, res, err)
		}
	}()
}

func (tt *TimedTask) add(key string, task TaskObj, spec int) error {
	if tt.tMap.IsExist(key) {
		return ErrTaskIsExist
	}
	NewTaskInfo(key, task, spec)
	tt.reSelectAfterUpdate()
	return nil
}

func (tt *TimedTask) Add(key string, task TaskObj, spec int) {
	err := tt.add(key, task, spec)
	nti := tt.tMap.Get(key)
	// tt.invokeAddCallback(nti, err)
	_ = err
	_ = nti
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
	err := tt.cancel(key)
	tt.invokeCancelCallback(key, err)
}

func (tt *TimedTask) goExecutor() {
	for i := 0; i < tt.routineCount; i++ {
		go func(rid int) {
			for {
				ti := <-tt.tasks
				res, err := ti.Task()
				// tt.invokeExecuteCallback(ti, res, err)
				_ = err
				_ = res
			}
		}(i)
	}
}

func (tt *TimedTask) goTimedIssue() {
	go func() {
		for {
			task := tt.tMap.SelectNextExec()
			if task == nil {
				// 任务列表中没有任务 等待刷新信号来到后 重新选择任务
				<-tt.refreshSign
				continue
			}
			spec := task.LastTime.Add(time.Duration(task.Spec) * time.Second).Sub(time.Now())
			if spec.Nanoseconds() < 0 {
				spec = time.Nanosecond
			}
			var ticker = time.NewTicker(spec)
			select {
			case <-ticker.C:
				if task != nil {
					// 先更新任务信息再执行任务 减少时间误差
					tt.updateMapAfterExec(task)
					tt.tasks <- task
				}
				break
			case <-tt.refreshSign:
				break
			}
		}
	}()
}

func (tt *TimedTask) updateMapAfterExec(task *TaskInfo) {
	task.Count += 1
	task.LastTime = task.LastTime.Add(time.Duration(task.Spec) * time.Second)
	tt.tMap.Set(task.Key, task)
}

// 触发更新定时最早一个被执行的定时任务
func (tt *TimedTask) reSelectAfterUpdate() {
	go func() {
		tt.refreshSign <- struct{}{}
	}()
}

// 获取定时任务列表信息
func (tt *TimedTask) GetTimedTaskInfo() map[string]*TaskInfo {
	return tt.tMap.GetAll()
}
