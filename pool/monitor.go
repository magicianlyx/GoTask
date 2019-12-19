package pool

import (
	"sync"
	"time"
)

type DynamicPoolMonitor struct {
	l sync.RWMutex        // 互斥锁
	k *Generator          // ID发生器
	s *StatusSettleMap    // 状态总结（只统计已经消亡的线程）
	g *GoroutineSettleMap // 线程总结 只存储当前存活线程
	c *Counter            // 计数器 用于加速获取当前存活线程数及记录线程存活峰值数
	o *Options            // 配置
}

func NewDynamicPoolMonitor(o *Options) *DynamicPoolMonitor {
	s := NewStatusSettleMap()
	g := NewGoroutineSettleMap()
	c := NewCounter()
	return &DynamicPoolMonitor{
		l: sync.RWMutex{},
		k: NewGenerator(),
		s: s,
		g: g,
		c: c,
		o: o,
	}
}

// 切换一个线程的状态
func (m *DynamicPoolMonitor) SwitchGoRoutineStatus(gid GoroutineUID) {
	m.l.Lock()
	defer m.l.Unlock()
	m.g.AutoSwitchGoRoutineStatus(gid)
}

// 获取当前存活线程数
func (m *DynamicPoolMonitor) GetGoroutineCount() int {
	m.l.RLock()
	defer m.l.RUnlock()
	// FIXME
	//  return m.c.Get()
	return len(m.g.m)
}

// 获取存活线程数最高峰值
func (m *DynamicPoolMonitor) GetGoroutinePeak() int {
	m.l.RLock()
	defer m.l.RUnlock()
	return int(m.c.GetMax())
}

// 获取当前活跃线程数
func (m *DynamicPoolMonitor) GetCurrentActiveCount() int {
	m.l.RLock()
	defer m.l.RUnlock()
	return m.g.GetActiveGoroutineCount()
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

// 尝试创建一个线程
func (m *DynamicPoolMonitor) TryConstruct(want bool) (GoroutineUID, bool) {
	m.l.Lock()
	defer m.l.Unlock()
	gc := m.c.Get()
	if gc == 0 {
		// 无存活线程 必须创建
		gid := m.construct()
		return gid, true
	} else if want && gc < int64(m.o.GoroutineLimit) {
		// 存活线程数没有溢出 可以创建
		gid := m.construct()
		return gid, true
	}
	return 0, false
	
}

// 销毁一个线程
func (m *DynamicPoolMonitor) destroy(gid GoroutineUID) {
	if m.g.GetCurrentStatus(gid) == GoroutineStatusActive {
		m.g.AutoSwitchGoRoutineStatus(gid)
	}
	
	// 将线程的状态信息累加到监控器的状态信息上
	settle := m.g.GetStatusSettle(gid)
	m.s.AddMultiStatusDuration(settle)
	
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

// 尝试关闭一个线程 如果线程最近活跃时长较短 则关闭线程
func (m *DynamicPoolMonitor) TryDestroy(gid GoroutineUID) bool {
	m.l.Lock()
	defer m.l.Unlock()
	if m.g.GetRecentActiveRatio(gid) < m.o.CloseLessThanF {
		m.destroy(gid)
		return true
	}
	return false
}

// 获取状态总结
func (m *DynamicPoolMonitor) GetStatusSettle() map[GoroutineStatus]time.Duration {
	m.l.RLock()
	defer m.l.RUnlock()
	sMap := m.s.GetAllStatusDuration()
	gMap := m.g.GetAllGoroutineStatusDuration()
	r := make(map[GoroutineStatus]time.Duration)
	for status := range sMap {
		r[status] = r[status] + sMap[status]
	}
	for status := range gMap {
		r[status] = r[status] + gMap[status]
	}
	return r
}
