package pool

import (
	"sync"
)

type DynamicPoolMonitor struct {
	l sync.RWMutex
	k *Generator
	g *GoroutineSettleMap
	c *Counter
	o *Options
}

func NewDynamicPoolMonitor(o *Options) *DynamicPoolMonitor {
	return &DynamicPoolMonitor{
		l: sync.RWMutex{},
		k: NewGenerator(),
		g: NewGoroutineSettleMap(),
		c: NewCounter(),
		o: o,
	}
}

// 切换一个线程的状态
func (m *DynamicPoolMonitor) SwitchGoRoutineStatus(gid GoroutineUID) {
	m.l.Lock()
	defer m.l.Unlock()
	m.g.AutoSwitchGoRoutineStatus(gid)
}

// 构建一个新线程
func (m *DynamicPoolMonitor) construct() GoroutineUID {
	gid := m.k.Generate()
	m.c.Inc()
	m.g.NewGoroutineSettle(gid, NewGoroutineSettle(m.o.AutoMonitorDuration))
	return gid
}

// 强制构建一个新线程
func (m *DynamicPoolMonitor) Construct() GoroutineUID {
	m.l.Lock()
	defer m.l.Unlock()
	return m.construct()
}

// 销毁一个线程
func (m *DynamicPoolMonitor) destroy(gid GoroutineUID) {
	if m.g.GetCurrentStatus(gid) == GoroutineStatusActive {
		m.g.AutoSwitchGoRoutineStatus(gid)
	}
	m.k.Collect(gid)
	m.c.Dec()
	m.g.DeleteGoroutineSettle(gid)
}

// 强制销毁一个线程
func (m *DynamicPoolMonitor) Destroy(gid GoroutineUID) {
	m.l.Lock()
	defer m.l.Unlock()
	m.destroy(gid)
}

// 获取当前存活线程数
func (m *DynamicPoolMonitor) GetGoroutineCount() int64 {
	return m.c.Get()
}

// 获取当前活跃线程数
func (m *DynamicPoolMonitor) GetCurrentActiveCount() int64 {
	return m.g.GetActiveGoroutineCount()
}

//// 是否还能创建一个线程
//func (m *DynamicPoolMonitor) CanConstructGoroutine() bool {
//	m.l.Lock()
//	defer m.l.Unlock()
//	if m.c.Get() == 0 && m.c.Get() < int64(m.o.GoroutineLimit) {
//		return true
//	} else {
//		return false
//	}
//}
//
// 尝试关闭一个线程 如果线程最近活跃时长较短 则关闭线程
func (m *DynamicPoolMonitor) TryDestroyGoroutine(gid GoroutineUID) bool {
	m.l.Lock()
	defer m.l.Unlock()
	if m.g.GetRecentActiveRatio(gid) < m.o.CloseLessThanF {
		m.destroy(gid)
		return true
	}
	return false
}
