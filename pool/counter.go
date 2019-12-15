package pool

import "sync"

// 计数器（线程安全）
type Counter struct {
	x   int64
	max int64
	l   sync.RWMutex
}

func NewCounter() *Counter {
	return &Counter{
		x:   0,
		max: 0,
		l:   sync.RWMutex{},
	}
}

// 自增
func (c *Counter) Inc() int64 {
	c.l.Lock()
	defer c.l.Unlock()
	c.x += 1
	if c.max < c.x {
		c.max = c.x
	}
	return c.x
}

// 自减
func (c *Counter) Dec() int64 {
	c.l.Lock()
	defer c.l.Unlock()
	c.x -= 1
	return c.x
}

// 获取峰值
func (c *Counter) GetMax() int64 {
	c.l.RLock()
	defer c.l.RUnlock()
	return c.max
}

// 获取值
func (c *Counter) Get() int64 {
	c.l.RLock()
	defer c.l.RUnlock()
	return c.x
}
