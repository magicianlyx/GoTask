package structure

import (
	"sync"
)

// 无序集合 线程安全
type Set struct {
	s sync.Map
}

func NewSet() *Set {
	return &Set{sync.Map{}}
}

func (s *Set) IsExist(key interface{}) bool {
	_, isExist := s.s.Load(key)
	return isExist
}

func (s *Set) Add(key interface{}) {
	s.s.Store(key, struct{}{})
}

func (s *Set) Delete(key interface{}) {
	s.s.Delete(key)
}
