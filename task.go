package task

import (
	"sync"
	"time"
	"errors"
	"fmt"
)

type ProcessArgs struct {
	Key  string    // 任务标识
	Time time.Time // 操作时间
}

type ExecuteArgs struct {
	Key             string                 // 任务标识
	Spec            int                    // 定时时长
	AddTime         time.Time              // 任务添加时间
	LastExecuteTime time.Time              // 任务最后执行时间
	Time            int                    // 方法被执行次数
	Error           error                  // 方法执行错误信息
	Res             map[string]interface{} // 方法执行结果
}

type objtask struct {
	Key  string                                 // 任务标识key
	Task func() (map[string]interface{}, error) // 任务方法
	Spec int                                    // 定时时长
}

type TaskInfo struct {
	Key  string                                 // 任务标识
	Spec int                                    // 定时时长
	Task func() (map[string]interface{}, error) // 任务方法
	AddTime         time.Time                   // 任务添加时间
	LastExecuteTime time.Time                   // 任务最后执行时间
	Time            int                         // 方法被执行次数
}

type BanInfo struct {
	Key     string    // 任务标识
	BanTime time.Time // 任务禁止时间
}

type TableInfo struct {
	TaskInfo map[string]TaskInfo
	BanInfo  map[string]BanInfo
}

// 定时任务处理表
type ReadyHandleTable struct {
	TableInfo
	m               map[string]chan struct{} // 用于停止任务的 任务key存在时，m的key也存在
	ready           chan objtask             // 任务方法
	l               sync.RWMutex             // 线程锁
	cancelCallBack  []func(ProcessArgs)      // 取消任务回调
	executeCallBack []func(ExecuteArgs)      // 执行任务回调
	addCallBack     []func(ProcessArgs)      // 添加任务回调
	banCallBack     []func(ProcessArgs)      // 禁封任务回调
	unBanCallBack   []func(ProcessArgs)      // 解封任务回调
}

func InitTaskRoutine() *ReadyHandleTable {
	t := &ReadyHandleTable{
		TableInfo: TableInfo{
			TaskInfo: make(map[string]TaskInfo),
			BanInfo:  make(map[string]BanInfo),
		},
		m:               make(map[string]chan struct{}),
		ready:           make(chan objtask),
		cancelCallBack:  []func(ProcessArgs){},
		executeCallBack: []func(ExecuteArgs){},
		addCallBack:     []func(ProcessArgs){},
		banCallBack:     []func(ProcessArgs){},
		unBanCallBack:   []func(ProcessArgs){},
	}
	go t.runClear()
	return t
}

// 阻塞执行清除任务，等待准备信号
func (t *ReadyHandleTable) runClear() {
	for {
		select {
		// 从任务执行管道中获取一个任务方法
		// 管道为空时会阻塞
		case c := <-t.ready:
			t.l.Lock()
			// 执行方法
			mp, err := c.Task()
			// 执行方法成功
			// 更新任务状态表
			t.updateTableInfoExecute(c.Key)
			// 执行回调
			for _, cb := range t.executeCallBack {
				ti := t.TaskInfo[c.Key]
				cb(ExecuteArgs{Key: ti.Key, Spec: ti.Spec, AddTime: ti.AddTime, LastExecuteTime: ti.LastExecuteTime, Error: err, Res: mp, Time: ti.Time})
			}
			t.l.Unlock()
			break
		}
	}
}

func (t *ReadyHandleTable) get(key string) (chan struct{}, bool) {
	v, ok := t.m[key]
	return v, ok
}

// 任务已被禁封或任务已经存在时返回error
func (t *ReadyHandleTable) add(key string, task func() (map[string]interface{}, error), second int) (error) {
	if t.isBan(key) {
		// 任务已被禁封 无法添加
		return errors.New(fmt.Sprintf("task prohibited, add task fail, key = %s\r\n", key))
	}
	_, ok := t.get(key)
	if ok {
		// 任务已经存在 无法再次添加
		return errors.New(fmt.Sprintf("task is exist, can not add again, key = %s\r\n", key))
	}
	done := make(chan struct{})
	t.m[key] = done
	ticker := time.NewTicker(time.Second * time.Duration(second))
	go func() {
		for {
			select {
			case <-ticker.C:
				// 向任务执行线程发送将要执行的任务
				t.ready <- objtask{
					Key:  key,
					Task: task,
					Spec: second,
				}
				break
			case <-done:
				// 接收到任务停止信号 取消任务
				ticker.Stop()
			}
		}
	}()
	// 添加任务成功
	// 更新状态表
	t.updateTableInfoAdd(key, second, task)
	// 触发回调
	for _, cb := range t.addCallBack {
		cb(ProcessArgs{key, time.Now()})
	}
	return nil
}

// 任务本身不存在时返回error
func (t *ReadyHandleTable) cancel(key string) (error) {
	v, ok := t.get(key)
	if !ok {
		// 任务不存在 无法取消
		return errors.New(fmt.Sprintf("cancel-task is not running, key = %s\r\n", key))
	} else {
		delete(t.m, key)
		close(v)// 取消成功
		// 更新状态表
		t.updateTableInfoCancel(key)
		// 取消成功触发回调方法
		for _, cb := range t.cancelCallBack {
			cb(ProcessArgs{key, time.Now()})
		}
		return nil
	}
}

// 任务已经被禁封时返回error
func (t *ReadyHandleTable) ban(key string) (error) {
	if !t.isBan(key) {
		t.BanInfo[key] = BanInfo{key, time.Now()}
		// 禁封成功
		// 触发禁封回调
		for _, cb := range t.banCallBack {
			cb(ProcessArgs{Key: key, Time: time.Now()})
		}
		return nil
	} else {
		return errors.New(fmt.Sprintf("task prohibited, can not ban again, key = %s\r\n", key))
	}
}

func (t *ReadyHandleTable) isBan(key string) bool {
	_, ok := t.BanInfo[key]
	return ok
}

// 任务本身没被禁封时 返回错误
func (t *ReadyHandleTable) unBan(key string) (error) {
	if t.isBan(key) {
		delete(t.BanInfo, key)
		for _, cb := range t.unBanCallBack {
			cb(ProcessArgs{Key: key, Time: time.Now()})
		}
		return nil
	} else {
		return errors.New(fmt.Sprintf("unban task, key = %s\r\n", key))
	}
}

// 添加任务时更新任务信息表的状态
func (t *ReadyHandleTable) updateTableInfoAdd(key string, second int, task func() (map[string]interface{}, error)) {
	t.TaskInfo[key] = TaskInfo{Key: key, Spec: second, Task: task, AddTime: time.Now()}
}

// 执行任务时更新任务信息表的状态
func (t *ReadyHandleTable) updateTableInfoExecute(key string) {
	ti, ok := t.TaskInfo[key]
	if ok {
		ti.Time += 1
		ti.LastExecuteTime = time.Now()
		t.TaskInfo[key] = ti
	}
}

// 取消任务时更新任务信息表的状态
func (t *ReadyHandleTable) updateTableInfoCancel(key string) {
	delete(t.TaskInfo, key)
}

// 添加任务取消回调
func (t *ReadyHandleTable) AddCancelCallBack(cb func(ProcessArgs)) {
	t.cancelCallBack = append(t.cancelCallBack, cb)
}

// 添加执行任务回调
func (t *ReadyHandleTable) AddExecuteCallBack(cb func(ExecuteArgs)) {
	t.executeCallBack = append(t.executeCallBack, cb)
}

// 添加添加任务回调
func (t *ReadyHandleTable) AddAddCallBack(cb func(ProcessArgs)) {
	t.addCallBack = append(t.addCallBack, cb)
}

// 添加禁封任务回调
func (t *ReadyHandleTable) AddBanCallBack(cb func(ProcessArgs)) {
	t.banCallBack = append(t.banCallBack, cb)
}

// 添加解封任务回调
func (t *ReadyHandleTable) AddUnBanCallBack(cb func(ProcessArgs)) {
	t.unBanCallBack = append(t.unBanCallBack, cb)
}

// 查看任务是否存在
func (t *ReadyHandleTable) Get(key string) (chan struct{}, bool) {
	t.l.RLock()
	defer t.l.RUnlock()
	return t.get(key)
}

// 取消任务
func (t *ReadyHandleTable) Cancel(key string) {
	t.l.RLock()
	defer t.l.RUnlock()
	t.cancel(key)
}

// 添加任务
func (t *ReadyHandleTable) Add(key string, task func() (map[string]interface{}, error), second int) {
	t.l.Lock()
	defer t.l.Unlock()
	t.add(key, task, second)
}

// 停止并禁用任务
func (t *ReadyHandleTable) StopAndBan(key string) {
	t.l.Lock()
	defer t.l.Unlock()
	t.cancel(key)
	t.ban(key)
}

// 重启服务
// 取消任务和启动任务各触发一次
func (t *ReadyHandleTable) Restart(key string) {
	t.l.Lock()
	defer t.l.Unlock()
	task := t.TaskInfo[key].Task
	spec := t.TaskInfo[key].Spec
	t.cancel(key)
	t.add(key, task, spec)
}

// 解除禁用任务+不执行
func (t *ReadyHandleTable) UnBan(key string) {
	t.l.Lock()
	defer t.l.Unlock()
	t.unBan(key)
}

// 任务是否被禁用
func (t *ReadyHandleTable) IsBan(key string) bool {
	t.l.Lock()
	defer t.l.Unlock()
	return t.isBan(key)
}

// 获取所有正在运行任务的信息
func (t *ReadyHandleTable) GetTaskInfo() (map[string]TaskInfo) {
	return t.TaskInfo
}

// 获取被禁任务的信息
func (t *ReadyHandleTable) GetBanInfo() (map[string]BanInfo) {
	return t.BanInfo
}
