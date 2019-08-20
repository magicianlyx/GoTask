package structure

import (
	"net"
	"sync"
)

// 无序集合 线程安全
type Set struct {
	s sync.Map
}

func NewSet() *Set {
	return &Set{sync.Map{}}
}

func (s *Set) IsExist(key string) bool {
	isExist := false
	s.s.Range(func(k, _ interface{}) bool {
		if kv, ok := k.(string); ok && kv == key {
			isExist = true
			return false
		} else {
			return true
		}
	})
	return isExist
}

func (s *Set) Add(key string) {
	s.s.Store(key, net.Interface{})
}

func (s *Set) Delete(key string) {
	s.s.Delete(key)
}
