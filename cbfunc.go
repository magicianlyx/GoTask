package GoTaskv1

import (
	"sync"
	"reflect"
)

type funcMap struct {
	fMap    sync.Map
}

func newFuncMap() *funcMap {
	return &funcMap{sync.Map{}}
}

func (fm *funcMap) add(i interface{}) {
	p := reflect.ValueOf(i).Pointer()
	fm.fMap.Store(p, i)
}

func (fm *funcMap) del(i interface{}) {
	p := reflect.ValueOf(i).Pointer()
	fm.fMap.Delete(p)
}

func (fm *funcMap) getAll(out interface{}) int {
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
