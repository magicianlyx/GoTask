package pool

import (
	"sync"
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
}

func NewGoroutinePool(options *Options) *GoroutinePool {
	options = options.Clone()
	options.fillDefaultOptions()
	m := NewDynamicPoolMonitor(options)
	return &GoroutinePool{
		c: make(chan TaskObj, options.TaskChannelSize),
		m: m,
		o: options,
	}
}

func (g *GoroutinePool) Put(obj TaskObj) {
	g.c <- obj
	g.checkPressure()
}

func (g *GoroutinePool) Stop() {
	close(g.e)
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
		for {
			t := time.NewTicker(g.o.AutoMonitorDuration)
			select {
			case task := <-g.c:
				// 执行任务task
				g.m.SwitchGoRoutineStatus(gid)
				task(gid)
				g.m.SwitchGoRoutineStatus(gid)
			case <-t.C:
				// 根据压力尝试关闭线程
				t.Stop()
				if g.m.TryDestroy(gid) {
					break
				} else {
					t = time.NewTicker(g.o.AutoMonitorDuration)
				}
			case <-c:
				// 单线程主动关闭
				g.m.Destroy(gid)
				break
			case <-g.e:
				// 主线程主动关闭
				g.m.Destroy(gid)
				break
			}
			t.Stop()
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
