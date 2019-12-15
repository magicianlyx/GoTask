package pool

import (
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// 单种状态总结（线程安全）
type Latency struct {
	l      sync.RWMutex
	status GoroutineStatus // 状态
	amount time.Duration   // 总时长
	last   time.Time       // 最后一次启动时刻
}

func NewLatency(status GoroutineStatus) *Latency {
	return &Latency{
		l:      sync.RWMutex{},
		status: status,
		amount: 0,
		last:   time.Time{},
	}
}

func (l *Latency) Clone() *Latency {
	return &Latency{
		l:      sync.RWMutex{},
		status: l.status,
		amount: l.amount,
		last:   l.last,
	}
}

func (l *Latency) IsStart() bool {
	return !l.last.IsZero()
}

func (l *Latency) Start() {
	l.last = time.Now()
}

func (l *Latency) Stop() {
	if !l.IsStart() {
		return
	}
	now := time.Now()
	latency := now.Sub(l.last)
	l.amount += latency
	l.last = time.Time{}
}

func (l *Latency) AmountDurationOfNow() *Latency {
	return l.AmountDurationOfTime(time.Now())
}

func (l *Latency) AmountDurationOfTime(t time.Time) *Latency {
	if !l.IsStart() {
		return l.Clone()
	}
	latency := t.Sub(l.last)
	return &Latency{
		l:      sync.RWMutex{},
		status: l.status,
		amount: l.amount + latency,
		last:   l.last,
	}
}

// 多状态总结（线程安全）
type LatencyMap struct {
	l sync.RWMutex
	m sync.Map
}

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

func (g *LatencyMap) Set(status GoroutineStatus, l *Latency) {
	g.l.Lock()
	defer g.l.Unlock()
	g.set(status, l)
}

func (g *LatencyMap) set(status GoroutineStatus, l *Latency) {
	g.m.Store(status, l)
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
			printf("internal data error, type of `key` is not `GoroutineStatus`")
			g.m.Delete(key)
			return true
		}
		if v, ok := value.(*Latency); ok {
			sv = v.AmountDurationOfNow()
		} else {
			printf("internal data error, type of `value` is not `Latency`")
			g.m.Delete(key)
			return true
		}
		r[sk] = sv
		return true
	})
	return r
}

func (g *LatencyMap) Clone() *LatencyMap {
	g.l.RLock()
	defer g.l.RUnlock()
	r := &LatencyMap{}
	g.m.Range(func(key, value interface{}) bool {
		var sk GoroutineStatus
		var sv *Latency
		if k, ok := key.(GoroutineStatus); ok {
			sk = k
		} else {
			printf("internal data error, type of `key` is not `GoroutineStatus`")
			g.m.Delete(key)
			return true
		}
		if v, ok := value.(*Latency); ok {
			sv = v.AmountDurationOfNow()
		} else {
			printf("internal data error, type of `value` is not `Latency`")
			g.m.Delete(key)
			return true
		}
		r.set(sk, sv)
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

// 单线程状态总结（线程安全）
type StatusSettle struct {
	l        sync.RWMutex
	status   GoroutineStatus
	duration time.Duration
}

func NewStatusSettle(status GoroutineStatus, duration time.Duration) *StatusSettle {
	return &StatusSettle{
		l:        sync.RWMutex{},
		status:   status,
		duration: duration,
	}
}

func (s *StatusSettle) AddDuration(duration time.Duration) {
	s.l.Lock()
	defer s.l.Unlock()
	s.duration += duration
}

// 多线程状态总结（线程安全）
type StatusSettleMap struct {
	l sync.RWMutex
	m sync.Map
}

func NewStatusSettleMap() *StatusSettleMap {
	return &StatusSettleMap{
		l: sync.RWMutex{},
		m: sync.Map{},
	}
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

func (m *StatusSettleMap) AddMultiStatusDuration(multiStatusDuration map[GoroutineStatus]time.Duration) {
	m.l.Lock()
	defer m.l.Unlock()
	for status := range multiStatusDuration {
		duration := multiStatusDuration[status]
		m.getOrCreate(status, 0).AddDuration(duration)
	}
}

// 单个线程信息总结（线程安全 但非强一致）
// FIXME PS：目前没有保证同一时刻s，m，lr 信息一致；如果要求状态信息完全同步的需要在结构体加锁及时间变量统一
// FIXME 数据状态需要强一致
// FIXME 已处理
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
func (g *GoroutineSettle) GetSurvivalDuration() time.Duration {
	g.l.RLock()
	defer g.l.RUnlock()
	return time.Now().Sub(g.createTime)
}

// 获取当前状态
func (g *GoroutineSettle) GetCurrentStatus() GoroutineStatus {
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
func (g *GoroutineSettle) GetStatusSettle() *LatencyMap {
	g.l.RLock()
	defer g.l.RUnlock()
	return g.m.Clone()
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
	s := g.GetSurvivalDuration()
	s = time.Duration(int64(math.Min(float64(int64(CaseRecentDuration)), float64(int64(s)))))
	return float64(c) / float64(s)
}

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
func (m *GoroutineSettleMap) GetCurrentStatus(gid GoroutineUID) GoroutineStatus {
	m.l.RLock()
	defer m.l.RUnlock()
	if gs, ok := m.get(gid); ok {
		return gs.GetCurrentStatus()
	}
	return GoroutineStatusNone
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

// 获取所有时间状态总结
func (m *GoroutineSettleMap) GetStatusSettle(gid GoroutineUID) *LatencyMap {
	m.l.RLock()
	defer m.l.RUnlock()
	if gs, ok := m.get(gid); ok {
		return gs.GetStatusSettle()
	}
	return nil
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

// 获取线程最近活跃占比
func (m *GoroutineSettleMap) GetRecentActiveRatio(gid GoroutineUID) float64 {
	m.l.RLock()
	defer m.l.RUnlock()
	if gs, ok := m.get(gid); ok {
		return gs.GetRecentActiveRatio()
	}
	return 0.0
}

// 获取当前处于活跃状态的线程数
func (m *GoroutineSettleMap) GetActiveGoroutineCount() int64 {
	m.l.RLock()
	defer m.l.RUnlock()
	var cnt int64 = 0

	for gid := range m.m {
		if m.m[gid].getStatus() == GoroutineStatusActive {
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
func (m *GoroutineSettleMap) DeleteGoroutineSettle(gid GoroutineUID) *LatencyMap {
	m.l.Lock()
	defer m.l.Unlock()

	var settle *LatencyMap
	if gs, ok := m.get(gid); ok {
		if m.GetCurrentStatus(gid) == GoroutineStatusActive {
			gs.AutoSwitchGoRoutineStatus()
		}
		settle = m.GetStatusSettle(gid)
		delete(m.m, gid)
	}
	return settle
}
