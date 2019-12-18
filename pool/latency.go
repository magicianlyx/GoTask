package pool

import (
	"sync"
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

func (l *Latency) clone() *Latency {
	return &Latency{
		l:      sync.RWMutex{},
		status: l.status,
		amount: l.amount,
		last:   l.last,
	}
}

func (l *Latency) Clone() *Latency {
	l.l.RLock()
	defer l.l.RUnlock()
	return l.clone()
}

func (l *Latency) isStart() bool {
	return !l.last.IsZero()
}

func (l *Latency) IsStart() bool {
	l.l.RLock()
	defer l.l.RUnlock()
	return l.isStart()
}

// 如果原先状态为启动 那么不操作
func (l *Latency) Start() {
	l.l.Lock()
	defer l.l.Unlock()
	if !l.isStart() {
		l.last = time.Now()
	}
}

// 如果原先状态为停止 那么不操作
func (l *Latency) Stop() {
	l.l.Lock()
	defer l.l.Unlock()
	if l.isStart() {
		now := time.Now()
		latency := now.Sub(l.last)
		l.amount += latency
		l.last = time.Time{}
	}
}

func (l *Latency) amountDurationOfTime(t time.Time) *Latency {
	c := l.clone()
	if !l.isStart() {
		return c
	}
	latency := t.Sub(c.last)
	return &Latency{
		l:      sync.RWMutex{},
		status: l.status,
		amount: l.amount + latency,
		last:   l.last,
	}
}

func (l *Latency) AmountDurationOfTime(t time.Time) *Latency {
	l.l.RLock()
	defer l.l.RUnlock()
	return l.amountDurationOfTime(t)
}

func (l *Latency) AmountDurationOfNow() *Latency {
	l.l.RLock()
	defer l.l.RUnlock()
	return l.amountDurationOfTime(time.Now())
}
