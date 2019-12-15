package pool

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestNewCounter(t *testing.T) {
	c := NewCounter()
	size := 100
	wg := sync.WaitGroup{}
	wg.Add(size)
	for i := 0; i < size; i++ {
		go func() {
			time.Sleep(time.Duration(rand.Int()%100) * time.Millisecond)
			c.Inc()
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Printf("c:%v\r\n", c.Get())


	size = 46
	wg = sync.WaitGroup{}
	wg.Add(size)
	for i := 0; i < size; i++ {
		go func() {
			time.Sleep(time.Duration(rand.Int()%100) * time.Millisecond)
			c.Dec()
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Printf("c:%v\r\n", c.Get())
	fmt.Printf("max:%v\r\n", c.GetMax())
}
