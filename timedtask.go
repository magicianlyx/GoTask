package GoTaskv1

import (
	"errors"
	"time"
	"sync/atomic"
	"sync"
)

var (
	ErrTaskIsExist    = errors.New("task is exist")
	ErrTaskIsNotExist = errors.New("task is not exist")
	ErrTaskIsBan      = errors.New("task is ban")
	ErrTaskIsUnBan    = errors.New("task is already unban")
)

type addCallback func(*AddCbArgs)
type cancelCallback func(*CancelCbArgs)
type executeCallback func(*ExecuteCbArgs)

type TimedTask struct {
	l               sync.Mutex
	tMap            *taskMap       // 定时任务字典
	bMap            *Set           // 被禁止添加执行的key
	tasks           chan *TaskInfo // 即将被执行的任务通道
	refreshSign     chan struct{}
	routineCount    int
	addCallback     []addCallback
	cancelCallback  []cancelCallback
	executeCallback []executeCallback
	singleValue     int64
}

func NewTimedTask(routineCount int) (*TimedTask) {
	tt := &TimedTask{
		sync.Mutex{},
		newtaskMap(),
		NewSet(),
		make(chan *TaskInfo),
		make(chan struct{}),
		routineCount,
		[]addCallback{},
		[]cancelCallback{},
		[]executeCallback{},
		0,
	}
	tt.goExecutor()
	tt.goTimedIssue()
	return tt
}

func (tt *TimedTask) AddAddCallback(cb func(*AddCbArgs)) {
	tt.addCallback = append(tt.addCallback, cb)
}
func (tt *TimedTask) AddCancelCallback(cb func(*CancelCbArgs)) {
	tt.cancelCallback = append(tt.cancelCallback, cb)
}
func (tt *TimedTask) AddExecuteCallback(cb func(*ExecuteCbArgs)) {
	tt.executeCallback = append(tt.executeCallback, cb)
}
func (tt *TimedTask) invokeAddCallback(info *TaskInfo, err error) {
	go func() {
		for _, cb := range tt.addCallback {
			cb(&AddCbArgs{info, err})
		}
	}()
}

func (tt *TimedTask) invokeCancelCallback(key string, err error) {
	go func() {
		for _, cb := range tt.cancelCallback {
			cb(&CancelCbArgs{key, err})
		}
	}()
}

func (tt *TimedTask) invokeExecuteCallback(info *TaskInfo, res map[string]interface{}, err error, rid int) {
	go func() {
		for _, cb := range tt.executeCallback {
			cb(&ExecuteCbArgs{info, res, err, rid})
		}
	}()
}

func (tt *TimedTask) isExist(key string) (bool) {
	return tt.tMap.isExist(key)
}

func (tt *TimedTask) add(key string, task TaskObj, spec int) error {
	if tt.tMap.isExist(key) {
		return ErrTaskIsExist
	}
	tt.tMap.add(key, newTaskInfo(key, task, spec))
	tt.reSelectAfterUpdate()
	return nil
}

func (tt *TimedTask) Add(key string, task TaskObj, spec int) {
	err := tt.add(key, task, spec)
	nti := newTaskInfo(key, task, spec)
	tt.invokeAddCallback(nti, err)
}

func (tt *TimedTask) cancel(key string) error {
	if !tt.tMap.isExist(key) {
		return ErrTaskIsNotExist
	}
	tt.tMap.delete(key)
	tt.reSelectAfterUpdate()
	return nil
}

func (tt *TimedTask) Cancel(key string) {
	err := tt.cancel(key)
	tt.invokeCancelCallback(key, err)
}

func (tt *TimedTask) ban(key string, error) {
	if tt.isBan(key) {
	
	}
}

func (tt *TimedTask) Ban(key string) {
	if tt.isBan(key) {
		return
	} else {
		tt.l.Lock()
		defer tt.l.Unlock()
		tt.cancel(key)
		tt.bMap.Add(key)
	}
	
}

func (tt *TimedTask) UnBan(key string) {
	if
}

func (tt *TimedTask) isBan(key string) (bool) {
	return tt.bMap.IsExist(key)
}

func (tt *TimedTask) goExecutor() {
	for i := 0; i < tt.routineCount; i++ {
		go func(rid int) {
			for {
				ti := <-tt.tasks
				res, err := ti.Task()
				tt.invokeExecuteCallback(ti, res, err, rid)
			}
		}(i)
	}
}

func (tt *TimedTask) goTimedIssue() {
	go func() {
		for {
			task := tt.tMap.selectNextExec()
			if task == nil {
				// 任务列表中没有任务 等待刷新信号来到后 重新选择任务
				<-tt.refreshSign
				continue
			}
			var spec time.Duration
			if task.LastTime.IsZero() {
				spec = task.AddTime.Add(time.Duration(task.Spec) * time.Second).Sub(time.Now())
			} else {
				spec = task.LastTime.Add(time.Duration(task.Spec) * time.Second).Sub(time.Now())
			}
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
	if task.LastTime.IsZero() {
		task.LastTime = task.AddTime.Add(time.Duration(task.Spec) * time.Second)
	} else {
		task.LastTime = task.LastTime.Add(time.Duration(task.Spec) * time.Second)
	}
	tt.tMap.set(task.Key, task)
}

// 触发更新定时最早一个被执行的定时任务
func (tt *TimedTask) reSelectAfterUpdate() {
	if atomic.LoadInt64(&tt.singleValue) > 0 {
		return
	}
	atomic.AddInt64(&tt.singleValue, 1)
	defer atomic.AddInt64(&tt.singleValue, -1)
	go func() {
		tt.refreshSign <- struct{}{}
	}()
}

// 获取定时任务列表信息
func (tt *TimedTask) GetTimedTaskInfo() map[string]*TaskInfo {
	return tt.tMap.getAll()
}
