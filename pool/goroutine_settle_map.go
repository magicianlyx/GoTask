package pool

import (
	"sync"
	"time"
)

// 多线程信息总结（线程安全）
type GoroutineSettleMap struct {
	l sync.RWMutex
	m map[GoroutineUID]*GoroutineSettle
}

func NewGoroutineSettleMap() *GoroutineSettleMap {
	return &GoroutineSettleMap{
		l: sync.RWMutex{},
		m: make(map[GoroutineUID]*GoroutineSettle),
	}
}

func (m *GoroutineSettleMap) get(gid GoroutineUID) (*GoroutineSettle, bool) {
	v, ok := m.m[gid]
	return v, ok
}

// 获取线程存活时间
func (m *GoroutineSettleMap) GetSurvivalDuration(gid GoroutineUID) time.Duration {
	m.l.RLock()
	defer m.l.RUnlock()
	if gs, ok := m.get(gid); ok {
		return gs.GetSurvivalDuration()
	}
	return 0
}

// 获取线程当前状态
func (m *GoroutineSettleMap) getCurrentStatus(gid GoroutineUID) GoroutineStatus {
	if gs, ok := m.get(gid); ok {
		return gs.GetStatus()
	}
	return GoroutineStatusNone
}

func (m *GoroutineSettleMap) GetCurrentStatus(gid GoroutineUID) GoroutineStatus {
	return m.getCurrentStatus(gid)
}

// 切换线程状态
func (m *GoroutineSettleMap) AutoSwitchGoRoutineStatus(gid GoroutineUID) GoroutineStatus {
	m.l.Lock()
	defer m.l.Unlock()
	
	if gs, ok := m.get(gid); ok {
		nstatus := gs.AutoSwitchGoRoutineStatus()
		return nstatus
	}
	return GoroutineStatusNone
}

// 获取指定线程所有时间状态总结
func (m *GoroutineSettleMap) GetStatusSettle(gid GoroutineUID) map[GoroutineStatus]time.Duration {
	m.l.RLock()
	defer m.l.RUnlock()
	if gs, ok := m.get(gid); ok {
		return gs.GetStatusSettle()
	}
	return nil
}

// 获取当前所有存活线程的所有状态总结
func (m *GoroutineSettleMap) GetAllGoroutineStatusDuration() map[GoroutineStatus]time.Duration {
	m.l.RLock()
	defer m.l.RUnlock()
	dsMap := make(map[GoroutineStatus]time.Duration)
	for gid := range m.m {
		gs := m.m[gid]
		gds := gs.GetStatusSettle()
		for status := range gds {
			dsMap[status] += gds[status]
		}
	}
	return dsMap
}

// 获取最近时间状态总结
func (m *GoroutineSettleMap) GetRecentStatusSettle(gid GoroutineUID) map[GoroutineStatus]time.Duration {
	m.l.RLock()
	defer m.l.RUnlock()
	if gs, ok := m.get(gid); ok {
		return gs.GetRecentStatusSettle()
	}
	return nil
}

// 获取指定线程最近活跃占比
func (m *GoroutineSettleMap) GetRecentActiveRatio(gid GoroutineUID) float64 {
	m.l.RLock()
	defer m.l.RUnlock()
	if gs, ok := m.get(gid); ok {
		return gs.GetRecentActiveRatio()
	}
	return 0.0
}

// 获取当前处于活跃状态的线程数
func (m *GoroutineSettleMap) GetActiveGoroutineCount() int {
	m.l.RLock()
	defer m.l.RUnlock()
	var cnt  = 0
	
	for gid := range m.m {
		if m.m[gid].GetStatus() == GoroutineStatusActive {
			cnt++
		}
	}
	return cnt
}

// 获取所有存活线程ID
func (m *GoroutineSettleMap) GetAllGoroutineUID() []GoroutineUID {
	m.l.RLock()
	defer m.l.RUnlock()
	l := make([]GoroutineUID, 0)
	for gid := range m.m {
		l = append(l, gid)
	}
	return l
}

// 构建一个新的线程
func (m *GoroutineSettleMap) NewGoroutineSettle(gid GoroutineUID, gs *GoroutineSettle) GoroutineUID {
	m.l.Lock()
	defer m.l.Unlock()
	m.m[gid] = gs
	return gid
}

// 终结一个线程
func (m *GoroutineSettleMap) DeleteGoroutineSettle(gid GoroutineUID) {
	m.l.Lock()
	defer m.l.Unlock()
	
	if gs, ok := m.get(gid); ok {
		if m.getCurrentStatus(gid) == GoroutineStatusActive {
			gs.AutoSwitchGoRoutineStatus()
		}
		delete(m.m, gid)
	}
}
