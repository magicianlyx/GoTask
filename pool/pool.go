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

func (g *GoroutinePool) close() {
	atomic.StoreInt64(&g.s, 1)
}

func (g *GoroutinePool) isClose() bool {
	return atomic.LoadInt64(&g.s) == 1
}

func (g *GoroutinePool) Put(obj TaskObj) {
	if !g.isClose() {
		g.c <- obj
		g.checkPressure()
	}
}

func (g *GoroutinePool) Stop() {
	g.close()
	close(g.c)
	close(g.e)
}

func (g *GoroutinePool) GetWorkCount() int64 {
	if g.isClose() {
		return 0
	}
	return int64(len(g.c))
	
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
				} else if g.isClose(){
					t.Stop()
					return
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

func (g *GoroutinePool) GetStatusSettle() map[GoroutineStatus]time.Duration {
	return g.m.GetStatusSettle()
}

func (g *GoroutinePool) GetCurrentActiveCount() int64 {
	return g.m.GetCurrentActiveCount()
}

func (g *GoroutinePool) GetGoroutineCount() int64 {
	return g.m.GetGoroutineCount()
}

func (g *GoroutinePool) GetGoroutinePeak() int64 {
	return g.m.GetGoroutinePeak()
}
