package pool

import (
	"sync"
	"time"
)

// 多状态总结（线程安全）
type LatencyMap struct {
	l sync.RWMutex
	m map[GoroutineStatus]*Latency
}

func NewLatencyMap() *LatencyMap {
	return &LatencyMap{
		sync.RWMutex{},
		make(map[GoroutineStatus]*Latency),
	}
}

func (g *LatencyMap) getOrCreate(status GoroutineStatus) *Latency {
	if v, ok := g.m[status]; ok {
		return v
	} else {
		l := NewLatency(status)
		g.m[status] = l
		return l
	}
}

func (g *LatencyMap) set(status GoroutineStatus, l *Latency) {
	g.m[status] = l
}

func (g *LatencyMap) Set(status GoroutineStatus, l *Latency) {
	g.l.Lock()
	defer g.l.Unlock()
	g.set(status, l)
}

func (g *LatencyMap) getAll() map[GoroutineStatus]time.Duration {
	r := make(map[GoroutineStatus]time.Duration)
	for status := range g.m {
		latency := g.m[status]
		latency = latency.AmountDurationOfNow()
		r[status] = latency.amount
	}
	return r
}
func (g *LatencyMap) GetAll() map[GoroutineStatus]time.Duration {
	g.l.RLock()
	defer g.l.RUnlock()
	return g.getAll()
}

func (g *LatencyMap) Clone() *LatencyMap {
	g.l.RLock()
	defer g.l.RUnlock()
	r := NewLatencyMap()
	for status := range g.m {
		latency := g.m[status]
		latency = latency.AmountDurationOfNow()
		r.set(status, latency)
	}
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
