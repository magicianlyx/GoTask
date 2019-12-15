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
	return &GoroutinePool{
		c: make(chan TaskObj, options.TaskChannelSize),
		m: NewDynamicPoolMonitor(options),
		o: options,
	}
}

func (g *GoroutinePool) Put(obj TaskObj) {
	g.c <- obj
	g.CheckPressure()
}

func (g *GoroutinePool) Stop() {
	close(g.e)
}

// 根据压力尝试创建线程
func (g *GoroutinePool) CheckPressure() {
	g.l.Lock()
	defer g.l.Unlock()
	gCnt := g.m.GetGoroutineCount()
	if gCnt == 0 {
		g.createGoroutine()
	} else if (float64(len(g.c))/float64(g.o.TaskChannelSize)) > g.o.NewGreaterThanF && gCnt < g.o.GoroutineLimit {
		g.createGoroutine()
	}
}

// 根据压力尝试关闭线程
func (g *GoroutinePool) TryDestroyGoroutine(gid GoroutineUID) {
	g.l.Lock()
	defer g.l.Unlock()
	g.m.TryDestroyGoroutine(gid)
}

// 新建一个线程
func (g *GoroutinePool) createGoroutine() (GoroutineUID, chan<- struct{}) {
	gid := g.m.Construct()
	c := make(chan struct{})
	t := time.NewTicker(g.o.AutoMonitorDuration)
	go func(gid GoroutineUID) {
		for {
			select {
			case task := <-g.c:
				// 执行任务task
				g.m.SwitchGoRoutineStatus(gid)
				task(gid)
				g.m.SwitchGoRoutineStatus(gid)
			case <-t.C:
				// 压力检测 尝试自杀
				t.Stop()
				t = time.NewTicker(g.o.AutoMonitorDuration)
				g.TryDestroyGoroutine(gid)
			case <-c:
				// 单线程主动关闭
				g.m.Destroy(gid)
				return
			case <-g.e:
				// 主线程主动关闭
				g.m.Destroy(gid)
				return
			}
		}
	}(gid)
	t.Stop()
	return gid, c
}
