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

type banCallback func(*BanCbArgs)
type unBanCallback func(*UnBanCbArgs)

type TimedTask struct {
	l               sync.RWMutex
	tMap            *taskMap       // 定时任务字典
	bMap            *Set           // 被禁止添加执行的key
	tasks           chan *TaskInfo // 即将被执行的任务通道
	refreshSign     chan struct{}
	routineCount    int
	addCallback     *funcMap
	cancelCallback  *funcMap
	executeCallback *funcMap
	banCallback     *funcMap
	unBanCallback   *funcMap
	singleValue     int64
}

func NewTimedTask(routineCount int) (*TimedTask) {
	tt := &TimedTask{
		sync.RWMutex{},
		newtaskMap(),
		NewSet(),
		make(chan *TaskInfo),
		make(chan struct{}),
		routineCount,
		newFuncMap(),
		newFuncMap(),
		newFuncMap(),
		newFuncMap(),
		newFuncMap(),
		0,
	}
	tt.goExecutor()
	tt.goTimedIssue()
	return tt
}

func (tt *TimedTask) AddAddCallback(cb func(*AddCbArgs)) {
	tt.addCallback.add(cb)
}

func (tt *TimedTask) DelAddCallback(cb func(*AddCbArgs)) {
	tt.addCallback.del(cb)
}

func (tt *TimedTask) AddCancelCallback(cb func(*CancelCbArgs)) {
	tt.cancelCallback.add(cb)
}

func (tt *TimedTask) DelCancelCallback(cb func(*CancelCbArgs)) {
	tt.cancelCallback.del(cb)
}

func (tt *TimedTask) AddExecuteCallback(cb func(*ExecuteCbArgs)) {
	tt.executeCallback.add(cb)
}
func (tt *TimedTask) DelExecuteCallback(cb func(*ExecuteCbArgs)) {
	tt.executeCallback.del(cb)
}
func (tt *TimedTask) AddBanCallback(cb func(*BanCbArgs)) {
	tt.banCallback.add(cb)
}
func (tt *TimedTask) DelBanCallback(cb func(*BanCbArgs)) {
	tt.banCallback.del(cb)
}
func (tt *TimedTask) AddUnBanCallback(cb func(*UnBanCbArgs)) {
	tt.unBanCallback.add(cb)
}

func (tt *TimedTask) DelUnBanCallback(cb func(*UnBanCbArgs)) {
	tt.unBanCallback.del(cb)
}

func (tt *TimedTask) invokeAddCallback(info *TaskInfo, err error) {
	go func() {
		addCallbacks := []addCallback{}
		tt.addCallback.getAll(&addCallbacks)
		for _, cb := range addCallbacks {
			cb(&AddCbArgs{info, err})
		}
	}()
}

func (tt *TimedTask) invokeCancelCallback(key string, err error) {
	go func() {
		cancelCallbacks := []cancelCallback{}
		tt.cancelCallback.getAll(&cancelCallbacks)
		for _, cb := range cancelCallbacks {
			cb(&CancelCbArgs{key, err})
		}
	}()
}

func (tt *TimedTask) invokeExecuteCallback(info *TaskInfo, res map[string]interface{}, err error, rid int) {
	go func() {
		executeCallbacks := []executeCallback{}
		tt.executeCallback.getAll(&executeCallbacks)
		for _, cb := range executeCallbacks {
			cb(&ExecuteCbArgs{info, res, err, rid})
		}
	}()
}

func (tt *TimedTask) invokeBanCallback(key string, err error) {
	go func() {
		banCallbacks := []banCallback{}
		tt.banCallback.getAll(&banCallbacks)
		for _, cb := range banCallbacks {
			cb(&BanCbArgs{key, err})
		}
	}()
}

func (tt *TimedTask) invokeUnBanCallback(key string, err error) {
	go func() {
		unBanCallbacks := []unBanCallback{}
		tt.unBanCallback.getAll(&unBanCallbacks)
		for _, cb := range unBanCallbacks {
			cb(&UnBanCbArgs{key, err})
		}
	}()
}

// func (tt *TimedTask) isExist(key string) (bool) {
// 	return tt.tMap.isExist(key)
// }

func (tt *TimedTask) add(key string, task TaskObj, spec int) error {
	if tt.tMap.isExist(key) {
		return ErrTaskIsExist
	}
	if tt.isBan(key) {
		return ErrTaskIsBan
	}
	tt.tMap.add(key, newTaskInfo(key, task, spec))
	tt.reSelectAfterUpdate()
	return nil
}

func (tt *TimedTask) addWithCb(key string, task TaskObj, spec int, cb bool) {
	tt.l.Lock()
	err := tt.add(key, task, spec)
	tt.l.Unlock()
	if cb {
		tt.invokeAddCallback(newTaskInfo(key, task, spec), err)
	}
}

func (tt *TimedTask) Add(key string, task TaskObj, spec int) {
	tt.addWithCb(key, task, spec, true)
}

func (tt *TimedTask) cancel(key string) error {
	if !tt.tMap.isExist(key) {
		return ErrTaskIsNotExist
	}
	tt.tMap.delete(key)
	tt.reSelectAfterUpdate()
	return nil
}

func (tt *TimedTask) cancelWithCb(key string, cb bool) {
	tt.l.Lock()
	err := tt.cancel(key)
	tt.l.Unlock()
	if cb {
		tt.invokeCancelCallback(key, err)
	}
}

func (tt *TimedTask) Cancel(key string) {
	tt.cancelWithCb(key, true)
}

func (tt *TimedTask) ban(key string) (error) {
	if tt.isBan(key) {
		return ErrTaskIsBan
	} else {
		tt.cancel(key)
		tt.bMap.Add(key)
		return nil
	}
}

func (tt *TimedTask) banWithCb(key string, cb bool) {
	tt.l.Lock()
	err := tt.ban(key)
	tt.l.Unlock()
	if cb {
		tt.invokeBanCallback(key, err)
	}
}

// 主动执行一次指定key任务
func (tt *TimedTask) Execute(key string) {
	ti := tt.tMap.get(key)
	if ti != nil {
		tt.tasks <- ti
	}
}

func (tt *TimedTask) Ban(key string) {
	tt.banWithCb(key, true)
}

func (tt *TimedTask) unBan(key string) error {
	if !tt.isBan(key) {
		return ErrTaskIsUnBan
	} else {
		tt.bMap.Delete(key)
	}
	return nil
}

func (tt *TimedTask) unBanWithCb(key string, cb bool) {
	tt.l.Lock()
	err := tt.unBan(key)
	tt.l.Unlock()
	if cb {
		tt.invokeUnBanCallback(key, err)
	}
}

func (tt *TimedTask) UnBan(key string) {
	tt.unBanWithCb(key, true)
}

func (tt *TimedTask) isBan(key string) (bool) {
	return tt.bMap.IsExist(key)
}

func (tt *TimedTask) IsBan(key string) (bool) {
	tt.l.RLock()
	b := tt.isBan(key)
	tt.l.RUnlock()
	return b
}

func (tt *TimedTask) goExecutor() {
	for i := 0; i < tt.routineCount; i++ {
		go func(rid int) {
			for {
				ti := <-tt.tasks
				if tt.tMap.get(ti.Key) != nil {
					res, err := ti.task()
					ti.LastResult = &TaskResult{res, err}
					tt.invokeExecuteCallback(ti, res, err, rid)
				}
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
			if spec.Nanoseconds() <= 0 {
				spec = time.Nanosecond
			}
			var ticker = time.NewTicker(spec)
			select {
			case <-ticker.C:
				// 先更新任务信息再执行任务 减少时间误差
				tt.updateMapAfterExec(task)
				tt.tasks <- task
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
