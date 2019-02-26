package GoTaskv1

import (
	"sync"
	"reflect"
)

type funcMap struct {
	fmap    sync.Map
	typedef reflect.Type
}

func newFuncMap() *funcMap {
	return &funcMap{sync.Map{}, nil}
}

func (fm *funcMap) add(i interface{}) {
	t := reflect.TypeOf(i)
	p := reflect.ValueOf(i).Pointer()
	if fm.typedef != nil && fm.typedef != t {
		return
	} else if fm.typedef == nil {
		fm.typedef = t
	}
	fm.fmap.Store(p, i)
}

func (fm *funcMap) del(i interface{}) {
	p := reflect.ValueOf(i).Pointer()
	fm.fmap.Delete(p)
}

func (fm *funcMap) getAll(out interface{}) int {
	ro := reflect.ValueOf(out).Elem()
	eo := make([]reflect.Value, 0)
	
	n := 0
	
	fm.fmap.Range(func(key, value interface{}) bool {
		n += 1
		eo = append(eo, reflect.ValueOf(value))
		return true
	})
	ro.Set(reflect.Append(ro, eo...))
	return n
}
