package pool

import (
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// 单个线程信息总结（线程安全）
type GoroutineSettle struct {
	l          sync.RWMutex
	s          atomic.Value  // 线程当前状态 GoroutineStatus
	m          *LatencyMap   // 各个状态汇总记录 map[GoroutineStatus]*Latency
	lr         *RecentRecord // 最近记录
	createTime time.Time     // 线程创建时间
}

func NewGoroutineSettle(d time.Duration) *GoroutineSettle {
	s := atomic.Value{}
	s.Store(GoroutineStatusNone)
	return &GoroutineSettle{
		l:          sync.RWMutex{},
		s:          s,
		m:          NewLatencyMap(),
		lr:         NewRecentRecord(d),
		createTime: time.Now(),
	}
}

func (g *GoroutineSettle) getStatus() GoroutineStatus {
	return g.s.Load().(GoroutineStatus)
}

func (g *GoroutineSettle) setStatus(s GoroutineStatus) {
	g.s.Store(s)
}

// 获取线程存活时间
func (g *GoroutineSettle) getSurvivalDuration() time.Duration {
	return time.Now().Sub(g.createTime)
}

func (g *GoroutineSettle) GetSurvivalDuration() time.Duration {
	g.l.RLock()
	defer g.l.RUnlock()
	return g.getSurvivalDuration()
}

// 获取当前状态
func (g *GoroutineSettle) GetStatus() GoroutineStatus {
	g.l.RLock()
	defer g.l.RUnlock()
	return g.getStatus()
}

// 切换线程状态
func (g *GoroutineSettle) AutoSwitchGoRoutineStatus() GoroutineStatus {
	g.l.Lock()
	defer g.l.Unlock()
	preStatus := g.getStatus()
	status := SwitchStatus(preStatus)
	g.setStatus(status)
	if preStatus.IsValid() {
		g.m.Stop(preStatus)
	}
	g.m.Start(status)
	g.lr.AddSwitchRecord(preStatus, status)
	return status
}

// 获取所有时间状态总结
func (g *GoroutineSettle) GetStatusSettle() map[GoroutineStatus]time.Duration {
	g.l.RLock()
	defer g.l.RUnlock()
	return g.m.GetAll()
}

// 获取最近时间状态总结
func (g *GoroutineSettle) GetRecentStatusSettle() map[GoroutineStatus]time.Duration {
	g.l.RLock()
	defer g.l.RUnlock()
	return g.lr.GetRecentSettle()
}

// 获取最近活跃占比
func (g *GoroutineSettle) GetRecentActiveRatio() float64 {
	g.l.RLock()
	defer g.l.RUnlock()
	m := g.lr.GetRecentSettle()
	c := m[GoroutineStatusActive]
	s := g.getSurvivalDuration()
	s = time.Duration(int64(math.Min(float64(int64(CaseRecentDuration)), float64(int64(s)))))
	return float64(c) / float64(s)
}
