package GoTask

import (
	"reflect"
	"sync"
)

type CbFuncMap struct {
	fMap sync.Map
}

func NewCbFuncMap() *CbFuncMap {
	return &CbFuncMap{sync.Map{}}
}

func (fm *CbFuncMap) Add(i interface{}) {
	p := reflect.ValueOf(i).Pointer()
	fm.fMap.Store(p, i)
}

func (fm *CbFuncMap) Del(i interface{}) {
	p := reflect.ValueOf(i).Pointer()
	fm.fMap.Delete(p)
}

func (fm *CbFuncMap) GetAll(out interface{}) int {
	ro := reflect.ValueOf(out).Elem()
	eo := make([]reflect.Value, 0)

	n := 0

	fm.fMap.Range(func(key, value interface{}) bool {
		n += 1
		eo = append(eo, reflect.ValueOf(value))
		return true
	})
	ro.Set(reflect.Append(ro, eo...))
	return n
}
