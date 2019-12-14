package pool

import (
	"math"
	"sync"
	"sync/atomic"
	"time"
)

const (
	GoroutineStatusNone   GoroutineStatus = 0
	GoroutineStatusActive GoroutineStatus = 1
	GoroutineStatusSleep  GoroutineStatus = 2
)

type GoroutineStatus int

func (s GoroutineStatus) ToString() string {
	if s == 0 {
		return "none"
	} else if s == 1 {
		return "active"
	} else {
		return "sleep"
	}
}

func (s GoroutineStatus) IsValid() bool {
	return s != GoroutineStatusNone
}

func SwitchStatus(p GoroutineStatus) (s GoroutineStatus) {
	switch p {
	case GoroutineStatusNone:
		return GoroutineStatusActive
	case GoroutineStatusSleep:
		return GoroutineStatusActive
	default:
		return GoroutineStatusSleep
	}
}

const (
	CaseRecentDuration = time.Second * 60
)

// 切换状态信息
type StatusSwitch struct {
	Time      time.Time       // 切换时间
	PreStatus GoroutineStatus // 切换前状态
	Status    GoroutineStatus // 切换后状态
}

func NewStatusSwitch(preStatus, status GoroutineStatus) *StatusSwitch {
	return &StatusSwitch{
		Time:      time.Now(),
		PreStatus: preStatus,
		Status:    status,
	}
}

func (s *StatusSwitch) Equal(t time.Time) bool {
	return s.Time.UnixNano() == t.UnixNano()
}

func (s *StatusSwitch) Lt(t time.Time) bool {
	return s.Time.UnixNano() < t.UnixNano()
}

func (s *StatusSwitch) Gte(t time.Time) bool {
	return s.Time.UnixNano() > t.UnixNano()
}

func (s *StatusSwitch) Clone() *StatusSwitch {
	return &StatusSwitch{
		Time:      s.Time,
		PreStatus: s.PreStatus,
		Status:    s.Status,
	}
}

// 存储最后d时长的切换记录
type RecentRecord struct {
	l sync.RWMutex    // 锁
	d time.Duration   // 最近有效时长
	m []*StatusSwitch // 存储切换记录
}

func NewRecentRecord(d time.Duration) *RecentRecord {
	return &RecentRecord{
		l: sync.RWMutex{},
		d: d,
		m: make([]*StatusSwitch, 0),
	}
}

// 获取最近的状态总结
func (l *RecentRecord) GetRecentSettle() map[GoroutineStatus]time.Duration {
	l.AdjustRecord()
	m := l.Clone() // 创建一个副本再去统计
	mLen := len(m)
	e := time.Now()

	o := make(map[GoroutineStatus]time.Duration)
	for i := mLen - 1; i >= 0; i-- {
		v := m[i]
		s := v.Time
		d := e.Sub(s)
		o[v.Status] += d
		e = v.Time
	}
	return o
}

// 调整 去除过期的切换记录
func (l *RecentRecord) AdjustRecord() {
	// 计算有效时间最早时刻
	now := time.Now()
	limit := now.Add(-l.d)

	l.l.Lock()
	// 获取有效时间内记录
	sLen := len(l.m)
	var limIdx = sLen
	for limIdx = 0; limIdx < sLen; limIdx++ {
		x := l.m[limIdx]
		if x.Gte(limit) {
			break
		}
	}
	// 0-limiIdx为超出有效时间的记录
	if limIdx == 0 {
		// 所有记录在有效时间内
		// 全部保留
	} else if limIdx == sLen {
		// 全部记录不在有效时间内
		// 保留当前状态 持续时间为有效时长
		prev := l.m[limIdx-1]
		prev.Time = limit
		l.m = []*StatusSwitch{prev}
	} else {
		x := l.m[limIdx]
		if x.Equal(limit) {
			l.m = l.m[limIdx:]
		} else {
			prev := l.m[limIdx-1]
			prev.Time = limit
			l.m = append([]*StatusSwitch{prev}, l.m[limIdx:]...)
		}
	}

	l.l.Unlock()
}

// 添加切换记录
func (l *RecentRecord) AddSwitchRecord(preStatus, status GoroutineStatus) {
	l.l.Lock()
	l.m = append(l.m, NewStatusSwitch(preStatus, status))
	l.l.Unlock()
}

// 生成副本
func (l *RecentRecord) Clone() []*StatusSwitch {
	r := make([]*StatusSwitch, 0)
	l.l.RLock()
	for i := range l.m {
		s := l.m[i]
		r = append(r, s.Clone())
	}
	l.l.RUnlock()
	return r
}

// 动态线程池监控器
type DynamicPoolMonitor struct {
	cl sync.RWMutex // 锁

	currentActiveCount int64 // 当前活跃线程数
	activeCountPeak    int64 // 活跃线程最高峰值

	//sl sync.RWMutex // 锁
	s sync.Map // 各个状态汇总记录 map[GoroutineStatus]time.Duration

	//ml sync.RWMutex // 锁
	m sync.Map // map[goroutine_id]*GoroutineSettle
}

func (m *DynamicPoolMonitor) AddSettle(sl map[GoroutineStatus]*Latency) {
	for k, v := range sl {
		//m.sl.Lock()
		var se time.Duration
		if v, ok := m.s.Load(k); ok {
			se = v.(time.Duration)
		}
		se += v.AmountDuration
		m.s.Store(k, se)
		m.sl.Unlock()
	}
}

// 获取各个状态的占用时长
func (m *DynamicPoolMonitor) GetSettle() map[GoroutineStatus]time.Duration {
	o := make(map[GoroutineStatus]time.Duration)
	gids := m.GetAllGoroutine()
	for _, gid := range gids {
		sm := m.GetGoroutineSettle(gid)
		for status := range sm {
			settle := sm[status]
			o[status] += settle.AmountDuration
		}
	}

	m.s.Range(func(key, value interface{}) bool {
		status := key.(GoroutineStatus)
		elapse := value.(time.Duration)
		o[status] = elapse
		return true
	})
	return o
}

func NewDynamicPoolMonitor() *DynamicPoolMonitor {
	return &DynamicPoolMonitor{
		cl:                 sync.RWMutex{},
		currentActiveCount: 0,
		activeCountPeak:    0,
		//ml:                 sync.RWMutex{},
		m: sync.Map{},
	}
}

// 创建一个新的线程id
func (m *DynamicPoolMonitor) newGoroutineID() int {
	// 从0起获取一个尚未分配的线程id
	for x := 0; true; x++ {
		if _, ok := m.m.Load(x); !ok {
			return x
		}
	}
	return -1
}

// 清除一个线程记录
func (m *DynamicPoolMonitor) collectGoroutineID(gid int) {
	m.m.Delete(gid)
}

// 状态切换
// none->active
// sleep->active
// active->sleep
func (m *DynamicPoolMonitor) switchStatus(p GoroutineStatus) (s GoroutineStatus) {
	switch p {
	case GoroutineStatusNone:
		return GoroutineStatusActive
	case GoroutineStatusSleep:
		return GoroutineStatusActive
	default:
		return GoroutineStatusSleep
	}
}

// （保证增加操作原子执行）
// 减少一个活跃线程数
func (m *DynamicPoolMonitor) decActiveCount() {
	m.cl.Lock()
	defer m.cl.Unlock()
	atomic.AddInt64(&m.currentActiveCount, -1)
}

// 增加一个活跃线程数
func (m *DynamicPoolMonitor) adcActiveCount() {
	m.cl.Lock()
	defer m.cl.Unlock()
	ac := atomic.AddInt64(&m.currentActiveCount, 1)
	ap := atomic.LoadInt64(&m.activeCountPeak)
	nap := int64(math.Max(float64(ap), float64(ac)))
	atomic.StoreInt64(&m.activeCountPeak, nap)
}

// 获取指定线程的总结
// 线程安全
func (m *DynamicPoolMonitor) getM(gid int) *GoroutineSettle {
	if v, ok := m.m.Load(gid); ok {
		if s, ok := v.(*GoroutineSettle); ok {
			return s
		} else {
			// 数据错误 删除线程信息
			m.m.Delete(gid)
		}
	}
	return nil
}

// 设置指定线程的总结
func (m *DynamicPoolMonitor) setM(gid int, gs *GoroutineSettle) {
	m.m.Store(gid, gs)
}

// 获取活跃线程峰值
func (m *DynamicPoolMonitor) GetActiveCountPeak() int64 {
	return atomic.LoadInt64(&m.activeCountPeak)
}

// 获取当前活跃线程数
func (m *DynamicPoolMonitor) GetCurrentActiveCount() int64 {
	return atomic.LoadInt64(&m.currentActiveCount)
}

// 获取所有线程ID
func (m *DynamicPoolMonitor) GetAllGoroutine() []int {
	gids := make([]int, 0)
	m.m.Range(func(key, value interface{}) bool {
		gids = append(gids, key.(int))
		return true
	})
	return gids
}

// 获取当前存在线程数
func (m *DynamicPoolMonitor) GetGoroutineCount() int {
	cnt := 0
	m.m.Range(func(key, value interface{}) bool {
		cnt += 1
		return true
	})
	return cnt
}

// 获取指定线程的总结
func (m *DynamicPoolMonitor) GetGoroutineSettle(gid int) map[GoroutineStatus]*Latency {
	if v := m.getM(gid); v != nil {
		return v.GetSettle()
	}
	return nil
}

// 获取指定线程的近期总结
func (m *DynamicPoolMonitor) GetRecentSettle(gid int) map[GoroutineStatus]time.Duration {
	if v := m.getM(gid); v != nil {
		return v.GetRecentSettle()
	}
	return nil
}

// 获取指定线程的最近活跃时长比例
func (m *DynamicPoolMonitor) GetRecentActiveRatio(gid int) float64 {

	if v := m.getM(gid); v != nil {
		m := v.GetRecentSettle()
		c := m[GoroutineStatusActive]

		s := v.GetSurvivalDuration()
		s = time.Duration(int64(math.Min(float64(int64(CaseRecentDuration)), float64(int64(s)))))

		return float64(c) / float64(s)
	}
	return 0.0
}

// 获取指定线程的当前状态
func (m *DynamicPoolMonitor) GetCurrentStatus(gid int) GoroutineStatus {
	var status = GoroutineStatusNone
	if s := m.getM(gid); s != nil {
		status = s.GetCurrentStatus()
	}
	return status
}

// 切换某线程状态
func (m *DynamicPoolMonitor) Switch(gid int) GoroutineStatus {
	var nstatus GoroutineStatus
	m.ml.Lock()
	if s := m.getM(gid); s != nil {
		status := s.GetCurrentStatus()
		nstatus = m.switchStatus(status)
		s.SwitchRoutineStatus(nstatus)
		m.setM(gid, s)
		if nstatus == GoroutineStatusActive {
			m.adcActiveCount()
		} else {
			m.decActiveCount()
		}
	}
	m.ml.Unlock()
	return nstatus
}

// 构建一个新线程 新线程处于为启动状态
func (m *DynamicPoolMonitor) Construct() int {
	gid := m.newGoroutineID()
	s := NewGoroutineSettle(CaseRecentDuration)

	m.ml.Lock()
	m.setM(gid, s)
	m.ml.Unlock()
	return gid
}

// 销毁一个线程
func (m *DynamicPoolMonitor) Destroy(gid int) {
	if m.GetCurrentStatus(gid) == GoroutineStatusActive {
		m.Switch(gid)
	}
	m.AddSettle(m.GetGoroutineSettle(gid))
	m.ml.Lock()
	m.collectGoroutineID(gid)
	m.ml.Unlock()
}

func (g *GoroutinePool) tryConstructAdjust() {

}

func (g *GoroutinePool) tryDestroyAdjust(gid int) {

}
