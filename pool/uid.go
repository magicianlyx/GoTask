package pool

import "sync"

type GoroutineUID int

type IGenerator interface {
	Generate() GoroutineUID
	Collect(gid GoroutineUID)
}

type Generator struct {
	contains sync.Map
}

func NewGenerator() *Generator {
	return &Generator{
		contains: sync.Map{},
	}
}

func (i *Generator) Generate() GoroutineUID {
	for x := GoroutineUID(0); true; x++ {
		if _, ok := i.contains.Load(x); !ok {
			i.contains.Store(x, struct{}{})
			return x
		}
	}
	return -1
}

func (i *Generator) Collect(gid GoroutineUID) {
	i.contains.Delete(gid)
}
