package pool

import (
	"sync"
	"sync/atomic"
	"time"
)

// 单种状态总结（非线程安全）
type Latency struct {
	Status         GoroutineStatus // 状态
	AmountDuration time.Duration   // 总时长
	LastStart      time.Time       // 最后一次启动时刻
}

func NewLatency(status GoroutineStatus) *Latency {
	return &Latency{Status: status}
}

func (l *Latency) Clone() *Latency {
	return &Latency{
		Status:         l.Status,
		AmountDuration: l.AmountDuration,
		LastStart:      l.LastStart,
	}
}

func (l *Latency) IsStart() bool {
	return !l.LastStart.IsZero()
}

func (l *Latency) Start() {
	l.LastStart = time.Now()
}

func (l *Latency) Stop() {
	if !l.IsStart() {
		return
	}
	now := time.Now()
	latency := now.Sub(l.LastStart)
	l.AmountDuration += latency
	l.LastStart = time.Time{}
}

func (l *Latency) AmountDurationOfNow() *Latency {
	return l.AmountDurationOTime(time.Now())
}

func (l *Latency) AmountDurationOTime(t time.Time) *Latency {
	if !l.IsStart() {
		return l.Clone()
	}
	latency := t.Sub(l.LastStart)
	return &Latency{
		Status:         l.Status,
		AmountDuration: l.AmountDuration + latency,
		LastStart:      l.LastStart,
	}
}

type LatencyMap struct {
	l sync.RWMutex
	m sync.Map
}

// 多状态总结（非线程安全）
func NewLatencyMap() *LatencyMap {
	return &LatencyMap{
		sync.RWMutex{},
		sync.Map{},
	}
}

func (g *LatencyMap) getOrCreate(status GoroutineStatus) *Latency {
	if v, ok := g.m.Load(status); ok {
		if l, ok := v.(*Latency); ok {
			return l
		} else {
			g.m.Delete(status)
		}
	}
	l := NewLatency(status)
	g.m.Store(status, l)
	return l
}

func (g *LatencyMap) set(status GoroutineStatus, l *Latency) {
	g.m.Store(status, l)
}

func (g *LatencyMap) delete(status GoroutineStatus) {
	g.m.Delete(status)
}

func (g *LatencyMap) GetAll() map[GoroutineStatus]*Latency {
	g.l.RLock()
	defer g.l.RUnlock()
	return g.getAll()
}

func (g *LatencyMap) getAll() map[GoroutineStatus]*Latency {
	r := make(map[GoroutineStatus]*Latency)
	g.m.Range(func(key, value interface{}) bool {
		var sk GoroutineStatus
		var sv *Latency
		if k, ok := key.(GoroutineStatus); ok {
			sk = k
		} else {
			g.m.Delete(key)
			return true
		}
		if v, ok := value.(*Latency); ok {
			sv = v.AmountDurationOfNow()
		} else {
			g.m.Delete(key)
			return true
		}
		r[sk] = sv
		return true
	})
	return r
}

func (g *LatencyMap) IsStart(status GoroutineStatus) bool {
	g.l.Lock()
	defer g.l.Unlock()
	return g.getOrCreate(status).IsStart()
}

func (g *LatencyMap) Start(status GoroutineStatus) {
	g.l.Lock()
	defer g.l.Unlock()
	g.getOrCreate(status).Start()
}

func (g *LatencyMap) Stop(status GoroutineStatus) {
	g.l.Lock()
	defer g.l.Unlock()
	g.getOrCreate(status).Stop()
}

// 单线程状态总结（非线程安全）
type StatusSettle struct {
	status   GoroutineStatus
	duration time.Duration
}

func NewStatusSettle(status GoroutineStatus, duration time.Duration) *StatusSettle {
	return &StatusSettle{
		status:   status,
		duration: duration,
	}
}

func (s *StatusSettle) AddDuration(duration time.Duration) {
	s.duration += duration
}

// 多线程状态总结（线程安全）
type StatusSettleMap struct {
	l sync.RWMutex
	m sync.Map
}

func (m *StatusSettleMap) getOrCreate(status GoroutineStatus, duration time.Duration) *StatusSettle {
	if v, ok := m.m.Load(status); ok {
		if l, ok := v.(*StatusSettle); ok {
			return l
		} else {
			m.m.Delete(status)
		}
	}
	settle := NewStatusSettle(status, duration)
	m.m.Store(status, settle)
	return settle
}

func (m *StatusSettleMap) set(status GoroutineStatus, settle *StatusSettle) {
	m.m.Store(status, settle)
}

func (m *StatusSettleMap) delete(status GoroutineStatus) {
	m.m.Delete(status)
}

func (m *StatusSettleMap) AddStatusDuration(status GoroutineStatus, duration time.Duration) {
	m.l.Lock()
	defer m.l.Unlock()
	m.getOrCreate(status, 0).AddDuration(duration)
}

// 单个线程信息总结（非强一致）
// FIXME PS：目前没有保证同一时刻s，m，lr 信息一致；如果要求状态信息完全同步的需要在结构体加锁及时间变量统一
type GoroutineSettle struct {
	s          atomic.Value  // 线程当前状态 GoroutineStatus
	m          *LatencyMap   // 各个状态汇总记录 map[GoroutineStatus]*Latency
	lr         *RecentRecord // 最近记录 FIXME
	createTime time.Time     // 线程创建时间
}

func NewGoroutineSettle(d time.Duration) *GoroutineSettle {
	s := atomic.Value{}
	s.Store(GoroutineStatusNone)
	return &GoroutineSettle{
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
func (g *GoroutineSettle) GetSurvivalDuration() time.Duration {
	return time.Now().Sub(g.createTime)
}

// 获取当前状态
func (g *GoroutineSettle) GetCurrentStatus() GoroutineStatus {
	return g.getStatus()
}

// 切换线程状态
func (g *GoroutineSettle) SwitchGoRoutineStatus(status GoroutineStatus) {
	if !status.IsValid() {
		return
	}
	preStatus := g.getStatus()
	g.setStatus(status)
	if preStatus.IsValid() {
		g.m.Stop(preStatus)
	}
	g.lr.AddSwitchRecord(preStatus, status)
}

// 获取所有时间状态总结
func (g *GoroutineSettle) GetStatusSettle() map[GoroutineStatus]*Latency {
	return g.m.GetAll()
}

// 获取最近时间状态总结
func (g *GoroutineSettle) GetRecentStatusSettle() map[GoroutineStatus]time.Duration {
	return g.lr.GetRecentSettle()
}
