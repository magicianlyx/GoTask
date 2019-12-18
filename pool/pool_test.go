package pool

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestNewGoroutinePool(t *testing.T) {
	options := &Options{}
	pool := NewGoroutinePool(options)
	fmt.Printf("%v\r\n", pool.o.GoroutineLimit)
	
	for i := 0; i < 100; i++ {
		go func(i int) {
			for {
				pool.Put(func(gid GoroutineUID) {
					if gid >= 18 {
						fmt.Printf("gid: %v\r\n", gid)
					}
				})
				time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			}
		}(i)
	}
	
	go func() {
		for {
			active := pool.GetCurrentActiveCount()
			count := pool.GetGoroutineCount()
			peak := pool.GetGoroutinePeak()
			settle := pool.GetStatusSettle()
			
		}
	}()
	
	time.Sleep(time.Hour)
}
