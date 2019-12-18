package pool

import (
	"sync"
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

// 状态转移
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
	if preStatus == status {
		// 非法逻辑
		printf(
			"can not transfer status from `%s` to `%s`\r\n",
			preStatus.ToString(),
			status.ToString(),
		)
	}
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
	l.adjustRecord()
	lc := l.Clone() // 创建一个副本再去统计
	m := lc.m
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
func (l *RecentRecord) adjustRecord() {
	// 计算有效时间最早时刻
	now := time.Now()
	limit := now.Add(-l.d)

	l.l.Lock()
	defer l.l.Unlock()
	// 获取有效时间内记录
	sLen := len(l.m)
	var limIdx = sLen
	for limIdx = 0; limIdx < sLen; limIdx++ {
		x := l.m[limIdx]
		if x.Gte(limit) {
			break
		}
	}
	// 0-limIdx为超出有效时间的记录
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
}

// 添加切换记录
func (l *RecentRecord) AddSwitchRecord(preStatus, status GoroutineStatus) {
	l.l.Lock()
	defer l.l.Unlock()
	l.m = append(l.m, NewStatusSwitch(preStatus, status))
}

// 生成副本
func (l *RecentRecord) Clone() *RecentRecord {
	l.l.RLock()
	defer l.l.RUnlock()
	r := make([]*StatusSwitch, 0)
	for i := range l.m {
		s := l.m[i]
		r = append(r, s.Clone())
	}
	return &RecentRecord{
		l: sync.RWMutex{},
		d: l.d,
		m: r,
	}
}
