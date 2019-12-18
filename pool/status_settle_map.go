package pool

import (
	"sync"
	"time"
)

// 多线程状态总结（线程安全）
type StatusSettleMap struct {
	l sync.RWMutex
	m map[GoroutineStatus]*StatusSettle
}

func NewStatusSettleMap() *StatusSettleMap {
	return &StatusSettleMap{
		l: sync.RWMutex{},
		m: make(map[GoroutineStatus]*StatusSettle),
	}
}

func (m *StatusSettleMap) getOrCreate(status GoroutineStatus, duration time.Duration) *StatusSettle {
	if settle, ok := m.m[status]; ok {
		return settle
	} else {
		settle := NewStatusSettle(status, duration)
		m.m[status] = settle
		return settle
	}
}

func (m *StatusSettleMap) GetOrCreate(status GoroutineStatus, duration time.Duration) *StatusSettle {
	m.l.Lock()
	defer m.l.Unlock()
	return m.getOrCreate(status, duration)
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

func (m *StatusSettleMap) getAllStatusDuration() map[GoroutineStatus]time.Duration {
	multiStatusDuration := make(map[GoroutineStatus]time.Duration)
	for status := range m.m {
		duration := m.m[status].GetDuration()
		multiStatusDuration[status] = multiStatusDuration[status] + duration
	}
	return multiStatusDuration
}

func (m *StatusSettleMap) GetAllStatusDuration() map[GoroutineStatus]time.Duration {
	m.l.RLock()
	defer m.l.RUnlock()
	return m.getAllStatusDuration()
}

