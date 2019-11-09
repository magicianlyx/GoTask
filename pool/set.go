package pool

import "sync"

type IDManager struct {
	contains sync.Map
}

func NewIDList() *IDManager {
	return &IDManager{
		contains: sync.Map{},
	}
}

func (i *IDManager) Create() int {
	for x := 0; true; x++ {
		if _, ok := i.contains.Load(x); !ok {
			i.contains.Store(x, struct{}{})
			return x
		}
	}
	return -1
}

func (i *IDManager) Remove(id int) {
	i.contains.Delete(id)
}
