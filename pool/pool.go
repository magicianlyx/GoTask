package pool

import (
	"sync"
	"sync/atomic"
	"time"
)

// 任务函数 gid为执行该任务的线程id
type TaskObj func(gid GoroutineUID)

type GoroutinePool struct {
	c chan TaskObj
	e chan struct{} // 停止所有线程信号
	l sync.RWMutex
	m *DynamicPoolMonitor
	o *Options
	s int64 // 0未关闭 1已关闭
}

func NewGoroutinePool(options *Options) *GoroutinePool {
	options = options.Clone()
	options.fillDefaultOptions()
	m := NewDynamicPoolMonitor(options)
	return &GoroutinePool{
		c: make(chan TaskObj, options.TaskChannelSize),
		e: make(chan struct{}),
		m: m,
		o: options,
		s: 0,
	}
}

// 设置关闭组件标识
func (g *GoroutinePool) close() {
	atomic.StoreInt64(&g.s, 1)
}

// 组件是否已关闭
func (g *GoroutinePool) isClose() bool {
	return atomic.LoadInt64(&g.s) == 1
}

// 向线程池推一个任务
func (g *GoroutinePool) Put(obj TaskObj) {
	if !g.isClose() {
		g.c <- obj
		g.checkPressure()
	}
}

// 关闭组件
func (g *GoroutinePool) Stop() {
	if !g.isClose() {
		g.close()
		close(g.c)
		close(g.e)
	}
}

// 根据压力尝试创建线程
func (g *GoroutinePool) checkPressure() {
	want := (float64(len(g.c)) / float64(g.o.TaskChannelSize)) > g.o.NewGreaterThanF
	if gid, ok := g.m.TryConstruct(want); ok {
		g.createGoroutine(gid)
	}
}

// 新建一个线程
func (g *GoroutinePool) createGoroutine(gid GoroutineUID) chan<- struct{} {
	c := make(chan struct{})
	go func(gid GoroutineUID) {
		t := time.NewTicker(g.o.AutoMonitorDuration)
		for {
			select {
			case task := <-g.c:
				if task != nil {
					// 执行任务task
					g.m.SwitchGoRoutineStatus(gid)
					task(gid)
					g.m.SwitchGoRoutineStatus(gid)
				} else if g.isClose() {
					// 主线程主动关闭
					break
				}
			case <-t.C:
				// 根据压力尝试关闭线程
				t.Stop()
				if g.m.TryDestroy(gid) {
					t.Stop()
					return
				} else {
					t = time.NewTicker(g.o.AutoMonitorDuration)
				}
			case <-c:
				// 单线程主动关闭
				g.m.Destroy(gid)
				t.Stop()
				return
			case <-g.e:
				// 主线程主动关闭
				g.m.Destroy(gid)
				t.Stop()
				return
			}
		}
	}(gid)
	return c
}

// 获取状态总结
func (g *GoroutinePool) GetStatusSettle() map[GoroutineStatus]time.Duration {
	return g.m.GetStatusSettle()
}

// 获取当前活跃线程数
func (g *GoroutinePool) GetCurrentActiveCount() int {
	return g.m.GetCurrentActiveCount()
}

// 获取当前存活线程数
func (g *GoroutinePool) GetGoroutineCount() int {
	return g.m.GetGoroutineCount()
}

// 获取线程池存活线程峰值
func (g *GoroutinePool) GetGoroutinePeak() int {
	return g.m.GetGoroutinePeak()
}

// 获取等待线程执行的任务数
func (g *GoroutinePool) GetWorkCount() int {
	if g.isClose() {
		return 0
	}
	return len(g.c)
}
