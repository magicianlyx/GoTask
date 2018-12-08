package gotask

import (
	"sync"
	"time"
	"errors"
	"fmt"
	"github.com/mohae/deepcopy"
)

/*
	因为不是同一个线程处理的问题 所以有可能在cancel回调之后任务还会执行一段时间
	知道cancel信号到达执行任务的线程并且处理完毕 任务才会停止执行
 */

const (
	unitTime = time.Millisecond * 1
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
	RoutineId       int                    // 执行该方法的线程id
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

type CallBack struct {
}

// 定时任务处理表
type ReadyHandleTable struct {
	m               map[string]chan struct{} // 用于停止任务的 任务key存在时，m的key也存在
	ready           chan objtask             // 任务方法
	l               sync.RWMutex             // 线程锁
	cancelCallBack  []func(ProcessArgs)      // 取消任务回调
	executeCallBack []func(ExecuteArgs)      // 执行任务回调
	addCallBack     []func(ProcessArgs)      // 添加任务回调
	banCallBack     []func(ProcessArgs)      // 禁封任务回调
	unBanCallBack   []func(ProcessArgs)      // 解封任务回调
	taskInfo        map[string]TaskInfo      // 任务信息
	banInfo         map[string]BanInfo       // 禁封键信息
	routineCount    int                      // 线程数
}

func InitTaskRoutine(routineCount int) *ReadyHandleTable {
	if routineCount < 1 {
		routineCount = 10
	}
	t := &ReadyHandleTable{
		taskInfo:        make(map[string]TaskInfo),
		banInfo:         make(map[string]BanInfo),
		m:               make(map[string]chan struct{}),
		ready:           make(chan objtask),
		cancelCallBack:  []func(ProcessArgs){},
		executeCallBack: []func(ExecuteArgs){},
		addCallBack:     []func(ProcessArgs){},
		banCallBack:     []func(ProcessArgs){},
		unBanCallBack:   []func(ProcessArgs){},
		routineCount:    routineCount,
	}
	t.runClear()
	return t
}

// 阻塞执行清除任务，等待准备信号
func (t *ReadyHandleTable) runClear() {
	for i := 0; i < t.routineCount; i++ {
		go func(i int) {
			for {
				select {
				// 从任务执行管道中获取一个任务方法
				// 管道为空时会阻塞
				case c := <-t.ready:
					func() {
						t.l.Lock()
						defer t.l.Unlock()
						if _, ok := t.get(c.Key); !ok {
							return
						}
						// 执行方法
						mp, err := c.Task()
						// 执行方法成功
						// 执行任务时更新任务信息表的状态
						ti, ok := t.taskInfo[c.Key]
						if ok {
							ti.Time += 1
							ti.LastExecuteTime = time.Now()
							t.taskInfo[c.Key] = ti
						}
						// 执行回调
						go func(key string, spec int, addTime time.Time, lastExecuteTime time.Time, count int, err error, mp map[string]interface{}, i int) {
							for _, cb := range t.executeCallBack {
								cb(ExecuteArgs{key, spec, addTime, lastExecuteTime, count, err, mp, i})
							}
						}(ti.Key, ti.Spec, ti.AddTime, ti.LastExecuteTime, ti.Time, err, mp, i)
					}()
					break
				}
			}
		}(i)
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
	ticker := time.NewTicker(unitTime * time.Duration(second))
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
				t.l.Lock()
				// 接收到任务停止信号 取消任务
				ticker.Stop()
				// 任务正式取消成功 该不会再任务执行

				t.l.Unlock()
				// 退出线程
				return
			}
		}
	}()
	// 添加任务成功
	// 添加任务时更新任务信息表的状态
	t.taskInfo[key] = TaskInfo{Key: key, Spec: second, Task: task, AddTime: time.Now()}
	// 触发回调
	go func(key string, now time.Time) {
		for _, cb := range t.addCallBack {
			cb(ProcessArgs{key, now})
		}
	}(key, time.Now())
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
		close(v) // 取消信号发送成功
		// 取消信号发送成功时更新任务信息表的状态
		delete(t.taskInfo, key)
		// 取消信号发送成功成功触发回调方法
		go func(key string, now time.Time) {
			for _, cb := range t.cancelCallBack {
				cb(ProcessArgs{key, now})
			}
		}(key, time.Now())
		return nil
	}
}

// 任务已经被禁封时返回error
func (t *ReadyHandleTable) ban(key string) (error) {
	if !t.isBan(key) {
		t.banInfo[key] = BanInfo{key, time.Now()}
		// 禁封成功
		// 触发禁封回调
		go func(key string, now time.Time) {
			for _, cb := range t.banCallBack {
				cb(ProcessArgs{key, now})
			}
		}(key, time.Now())
		return nil
	} else {
		return errors.New(fmt.Sprintf("task prohibited, can not ban again, key = %s\r\n", key))
	}
}

func (t *ReadyHandleTable) isBan(key string) bool {
	_, ok := t.banInfo[key]
	return ok
}

// 任务本身没被禁封时 返回错误
func (t *ReadyHandleTable) unBan(key string) (error) {
	if t.isBan(key) {
		// 解封方法
		delete(t.banInfo, key)
		// 触发回调
		go func(key string, now time.Time) {
			for _, cb := range t.unBanCallBack {
				cb(ProcessArgs{key, now})
			}
		}(key, time.Now())
		return nil
	} else {
		return errors.New(fmt.Sprintf("unban task, key = %s\r\n", key))
	}
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
	t.l.Lock()
	defer t.l.Unlock()
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
// 任务存在才能重启 配置照旧
func (t *ReadyHandleTable) Restart(key string) {
	t.l.Lock()
	defer t.l.Unlock()
	// 任务存在才能重启
	if _, ok := t.get(key); ok {
		task := t.taskInfo[key].Task
		spec := t.taskInfo[key].Spec
		t.cancel(key)
		t.add(key, task, spec)
	} else {
		//任务不存在
	}
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
	t.l.RLock()
	defer t.l.RUnlock()
	v := deepcopy.Copy(t.taskInfo)
	r, _ := v.(map[string]TaskInfo)
	return r
}

// 获取被禁任务的信息
func (t *ReadyHandleTable) GetBanInfo() (map[string]BanInfo) {
	t.l.RLock()
	defer t.l.RUnlock()
	v := deepcopy.Copy(t.banInfo)
	r, _ := v.(map[string]BanInfo)
	return r
}
