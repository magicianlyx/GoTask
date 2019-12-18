package pool

import (
	"sync"
	"time"
)

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

func (s *StatusSettle) GetDuration() time.Duration {
	s.l.RLock()
	defer s.l.RUnlock()
	return s.duration
}
