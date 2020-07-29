package pool

import (
	"fmt"
	"GoTask/utils"
	"math/rand"
	"testing"
	"time"
)

func TestNewGoroutinePool(t *testing.T) {
	options := &Options{}
	pool := NewGoroutinePool(options)
	fmt.Printf("%v\r\n", utils.ToJson(pool.o))
	
	for i := 0; i < 10; i++ {
		go func(i int) {
			for j := 0; j < 10000000; j++ {
				pool.Put(func(gid GoroutineUID) {
					if gid >= 18 {
						fmt.Printf("gid: %v\r\n", gid)
					}
					time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
				})
				if j%1000 == 0 {
					time.Sleep(10 * time.Second)
					
				}
				time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
			}
		}(i)
	}
	
	go func() {
		i := 0
		for {
			time.Sleep(time.Second)
			i++
			// fmt.Printf("time.Sleep(time.Second)\r\n")
			active := pool.GetCurrentActiveCount()
			// fmt.Printf("active := pool.GetCurrentActiveCount()\r\n")
			count := pool.GetGoroutineCount()
			// fmt.Printf("count := pool.GetGoroutineCount()\r\n")
			peak := pool.GetGoroutinePeak()
			// fmt.Printf("peak := pool.GetGoroutinePeak()\r\n")
			settle := pool.GetStatusSettle()
			fmt.Printf("active: %v  count: %v  peak: %v  channel: %v  \r\n",
				active,
				count,
				peak,
				len(pool.c),
			)
			
			
			settleMap := map[string]string{}
			for i := range settle {
				status := i.ToString()
				duration := fmt.Sprintf("%v", settle[i])
				settleMap[status] = duration
			}
			fmt.Printf("%s\r\n",
				utils.ToJson(settleMap),
			)
			
			_ = settle
			
		}
	}()
	
	// go func() {
	// 	for {
	// 		time.Sleep(time.Second)
	// 		fmt.Printf("channel: %v\r\n", len(pool.c))
	// 	}
	// }()
	
	time.Sleep(time.Hour)
}
