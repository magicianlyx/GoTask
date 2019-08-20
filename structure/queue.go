package structure

import (
	"sync"
)

type Queue struct {
	l    sync.RWMutex
	list []interface{}
	size int
}

func NewQueue(size int) *Queue {
	if size <= 0 {
		size = 10
	}
	return &Queue{
		l:    sync.RWMutex{},
		list: make([]interface{}, size),
		size: size,
	}
}

func (q *Queue) Push(v ...interface{}) {
	q.l.Lock()
	defer q.l.Unlock()
	if len(q.list) >= q.size {
		q.list = q.list[1:]
	}
	q.list = append(q.list, v)
}

func (q *Queue) Pop() interface{} {
	q.l.Lock()
	defer q.l.Unlock()

	if len(q.list) >= 1 {
		o := q.list[0]
		q.list = q.list[1:]
		return o
	} else {
		return nil
	}
}

func (q *Queue) Peek() interface{} {
	q.l.RLock()
	defer q.l.RUnlock()
	if len(q.list) >= 1 {
		return q.list[0]
	} else {
		return nil
	}
}

func (q *Queue) Clone() *Queue {
	list := make([]interface{}, q.size)
	q.l.RLock()
	for i := range q.list {
		list[i] = q.list[i]
	}
	q.l.RUnlock()
	return &Queue{
		l:    sync.RWMutex{},
		list: list,
		size: q.size,
	}
}
