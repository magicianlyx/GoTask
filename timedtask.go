package GoTask

import (
	"gitee.com/magicianlyx/GoTask/pool"
	"gitee.com/magicianlyx/GoTask/structure"
	"gitee.com/magicianlyx/GoTask/task"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrTaskIsExist    = errors.New("task is exist")
	ErrTaskIsNotExist = errors.New("task is not exist")
	ErrTaskIsBan      = errors.New("task is ban")
	ErrTaskIsUnBan    = errors.New("task is already unban")
)

type addCallback func(*task.AddCbArgs)
type cancelCallback func(*task.CancelCbArgs)
type executeCallback func(*task.ExecuteCbArgs)
type banCallback func(*task.BanCbArgs)
type unBanCallback func(*task.UnBanCbArgs)

type TimedTask struct {
	l                    sync.RWMutex
	tMap                 *task.TaskMap       // 定时任务字典
	bMap                 *structure.Set      // 被禁止添加执行的key
	tasks                chan *task.TaskInfo // 即将被执行的任务通道
	refreshSign          chan struct{}       // 刷新信号通知通道
	singleValue          int64               // 保证同一时刻单刷新信号
	shutdownExecutorSign chan struct{}       // 任务执行线程 停止信号通知通道
	shutdownIssueSign    chan struct{}       // 任务发射线程 停止信号通知通道
	routineCount         int
	addCallback          *CbFuncMap
	cancelCallback       *CbFuncMap
	executeCallback      *CbFuncMap
	banCallback          *CbFuncMap
	unBanCallback        *CbFuncMap
	wg                   *sync.WaitGroup
}

func NewTimedTask(maxRoutineCount int) *TimedTask {
	tt := &TimedTask{
		sync.RWMutex{},
		task.NewTaskMap(),
		structure.NewSet(),
		make(chan *task.TaskInfo),
		make(chan struct{}),
		0,
		make(chan struct{}),
		make(chan struct{}),
		maxRoutineCount,
		NewCbFuncMap(),
		NewCbFuncMap(),
		NewCbFuncMap(),
		NewCbFuncMap(),
		NewCbFuncMap(),
		&sync.WaitGroup{},
	}
	tt.goExecutor()
	// tt.goExecutorV2(maxRoutineCount)
	tt.goTimedIssue()
	return tt
}

func (tt *TimedTask) Stop() {
	tt.shutdownIssueSign <- struct{}{}
	for i := 0; i < int(tt.routineCount); i++ {
		tt.shutdownExecutorSign <- struct{}{}
	}
	tt.wg.Wait()
	close(tt.tasks)
	close(tt.refreshSign)
	return
}

func (tt *TimedTask) AddAddCallback(cb func(*task.AddCbArgs)) {
	tt.addCallback.Add(cb)
}

func (tt *TimedTask) DelAddCallback(cb func(*task.AddCbArgs)) {
	tt.addCallback.Del(cb)
}

func (tt *TimedTask) AddCancelCallback(cb func(*task.CancelCbArgs)) {
	tt.cancelCallback.Add(cb)
}

func (tt *TimedTask) DelCancelCallback(cb func(*task.CancelCbArgs)) {
	tt.cancelCallback.Del(cb)
}

func (tt *TimedTask) AddExecuteCallback(cb func(*task.ExecuteCbArgs)) {
	tt.executeCallback.Add(cb)
}

func (tt *TimedTask) DelExecuteCallback(cb func(*task.ExecuteCbArgs)) {
	tt.executeCallback.Del(cb)
}

func (tt *TimedTask) AddBanCallback(cb func(*task.BanCbArgs)) {
	tt.banCallback.Add(cb)
}

func (tt *TimedTask) DelBanCallback(cb func(*task.BanCbArgs)) {
	tt.banCallback.Del(cb)
}

func (tt *TimedTask) AddUnBanCallback(cb func(*task.UnBanCbArgs)) {
	tt.unBanCallback.Add(cb)
}

func (tt *TimedTask) DelUnBanCallback(cb func(*task.UnBanCbArgs)) {
	tt.unBanCallback.Del(cb)
}

func (tt *TimedTask) invokeAddCallback(info *task.TaskInfo, err error) {
	go func() {
		addCallbacks := make([]addCallback, 0)
		tt.addCallback.GetAll(&addCallbacks)
		for _, cb := range addCallbacks {
			cb(&task.AddCbArgs{info, err})
		}
	}()
}

func (tt *TimedTask) invokeCancelCallback(key string, err error) {
	go func() {
		cancelCallbacks := make([]cancelCallback, 0)
		tt.cancelCallback.GetAll(&cancelCallbacks)
		for _, cb := range cancelCallbacks {
			cb(&task.CancelCbArgs{key, err})
		}
	}()
}

func (tt *TimedTask) invokeExecuteCallback(info *task.TaskInfo, res map[string]interface{}, err error, gid pool.GoroutineUID) {
	go func() {
		executeCallbacks := make([]executeCallback, 0)
		tt.executeCallback.GetAll(&executeCallbacks)
		for _, cb := range executeCallbacks {
			cb(&task.ExecuteCbArgs{info, res, err, gid})
		}
	}()
}

func (tt *TimedTask) invokeBanCallback(key string, err error) {
	go func() {
		banCallbacks := make([]banCallback, 0)
		tt.banCallback.GetAll(&banCallbacks)
		for _, cb := range banCallbacks {
			cb(&task.BanCbArgs{key, err})
		}
	}()
}

func (tt *TimedTask) invokeUnBanCallback(key string, err error) {
	go func() {
		unBanCallbacks := make([]unBanCallback, 0)
		tt.unBanCallback.GetAll(&unBanCallbacks)
		for _, cb := range unBanCallbacks {
			cb(&task.UnBanCbArgs{key, err})
		}
	}()
}

func (tt *TimedTask) add(key string, obj task.TaskObj, sche task.ISchedule) error {
	if tt.tMap.IsExist(key) {
		return ErrTaskIsExist
	}
	if tt.isBan(key) {
		return ErrTaskIsBan
	}
	tt.tMap.Add(key, task.NewTaskInfo(key, obj, sche))
	tt.reSelectAfterUpdate()
	return nil
}

func (tt *TimedTask) addWithCb(key string, obj task.TaskObj, sche task.ISchedule, cb bool) {
	tt.l.Lock()
	err := tt.add(key, obj, sche)
	tt.l.Unlock()
	if cb {
		tt.invokeAddCallback(task.NewTaskInfo(key, obj, sche), err)
	}
}

func (tt *TimedTask) Add(key string, obj task.TaskObj, sche task.ISchedule) {
	tt.addWithCb(key, obj, sche, true)
}

func (tt *TimedTask) set(key string, obj task.TaskObj, sche task.ISchedule) error {
	if tt.isBan(key) {
		return ErrTaskIsBan
	}
	tt.tMap.AddOrSet(key, task.NewTaskInfo(key, obj, sche))
	tt.reSelectAfterUpdate()
	return nil
}

func (tt *TimedTask) setWithCb(key string, obj task.TaskObj, sche task.ISchedule, cb bool) {
	tt.l.Lock()
	err := tt.set(key, obj, sche)
	tt.l.Unlock()
	if cb {
		tt.invokeAddCallback(task.NewTaskInfo(key, obj, sche), err)
	}
}

func (tt *TimedTask) Set(key string, obj task.TaskObj, sche task.ISchedule) {
	tt.setWithCb(key, obj, sche, true)
}

func (tt *TimedTask) cancel(key string) error {
	if !tt.tMap.IsExist(key) {
		return ErrTaskIsNotExist
	}
	tt.tMap.Delete(key)
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

func (tt *TimedTask) ban(key string) error {
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

// 主动执行一次指定key任务 不影响既有定时任务 执行的记录将会添加到任务总结信息中
func (tt *TimedTask) Execute(key string) {
	ti := tt.tMap.Get(key)
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

func (tt *TimedTask) isBan(key string) bool {
	return tt.bMap.IsExist(key)
}

func (tt *TimedTask) IsBan(key string) bool {
	tt.l.RLock()
	b := tt.isBan(key)
	tt.l.RUnlock()
	return b
}

func (tt *TimedTask) isExist(key string) bool {
	return tt.tMap.IsExist(key)
}

func (tt *TimedTask) IsExist(key string) bool {
	tt.l.RLock()
	b := tt.isExist(key)
	tt.l.RUnlock()
	return b
}

func (tt *TimedTask) goExecutor() {
	for i := 0; i < int(tt.routineCount); i++ {
		go func(rid int) {
			tt.wg.Add(1)
			defer tt.wg.Done()
			for {
				var ti *task.TaskInfo
				select {
				case ti = <-tt.tasks:
					break
				case <-tt.shutdownExecutorSign:
					return
				}
				if tt.tMap.Get(ti.Key) != nil {
					// 执行任务
					res, err := ti.Task()
					ti.LastResult = &task.TaskResult{res, err}

					// 如果没有下一次的执行计划 那么将会清除任务
					if !ti.HasNextExecute() {
						tt.tMap.Delete(ti.Key)
					}

					// 执行回调
					tt.invokeExecuteCallback(ti, res, err, pool.GoroutineUID(rid))

				}
			}
		}(i)
	}
}

func (tt *TimedTask) goExecutorV2(maxRoutineCount int) {
	options := &pool.Options{
		GoroutineLimit: maxRoutineCount,
	}

	grd := pool.NewGoroutinePool(options)

	go func() {
		tt.wg.Add(1)
		defer tt.wg.Done()
		for {
			var ti *task.TaskInfo
			select {
			case ti = <-tt.tasks:
				break
			case <-tt.shutdownExecutorSign:
				grd.Stop()
				return
			}
			// 构成一个任务
			task := func(gid pool.GoroutineUID) {
				if tt.tMap.Get(ti.Key) != nil {

					// 执行任务
					res, err := ti.Task()
					ti.LastResult = &task.TaskResult{res, err}

					// 如果没有下一次的执行计划 那么将会清除任务
					if !ti.HasNextExecute() {
						tt.tMap.Delete(ti.Key)
					}

					// 执行回调
					tt.invokeExecuteCallback(ti, res, err, gid)
				}
			}

			// 向动态线程池派发一个任务
			grd.Put(task)
		}
	}()
}

func (tt *TimedTask) goTimedIssue() {
	go func() {
		tt.wg.Add(1)
		defer tt.wg.Done()
		for {
			task, spec, ok := tt.tMap.SelectNextExec()
			if !ok {
				// 任务列表中没有任务 等待刷新信号来到后 重新选择任务
				select {
				case <-tt.refreshSign:
					continue
				case <-tt.shutdownIssueSign:
					return
				}
			}

			var ticker = time.NewTicker(spec)
			select {
			case <-ticker.C:
				ticker.Stop()
				// 先更新任务信息再执行任务 防止调度出问题
				tt.updateMapBeforeExec(task)
				tt.tasks <- task
				break
			case <-tt.refreshSign:
				ticker.Stop()
				break
			case <-tt.shutdownIssueSign:
				ticker.Stop()
				return
			}
		}
	}()
}

func (tt *TimedTask) updateMapBeforeExec(task *task.TaskInfo) {
	// 更新任务信息
	task.Update()
	// 写回到字典 更新
	tt.tMap.Set(task.Key, task)
}

// 触发更新定时最早一个被执行的定时任务
func (tt *TimedTask) reSelectAfterUpdate() {
	if atomic.LoadInt64(&tt.singleValue) > 0 {
		return
	}
	atomic.AddInt64(&tt.singleValue, 1)
	defer atomic.AddInt64(&tt.singleValue, -1)
	tt.refreshSign <- struct{}{}
}

// 获取定时任务列表信息
func (tt *TimedTask) GetTimedTaskInfo() map[string]*task.TaskInfo {
	return tt.tMap.GetAll()
}
