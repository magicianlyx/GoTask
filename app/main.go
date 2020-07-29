package main

import (
	"fmt"
	"gitee.com/magicianlyx/GoTask/pool"
	"math/rand"
	"time"
)

func main() {
	
	options := &pool.Options{}
	p := pool.NewGoroutinePool(options)
	
	for i := 0; i < 10; i++ {
		go func(i int) {
			for j := 0; j < 10000000; j++ {
				p.Put(func(gid pool.GoroutineUID) {
					if gid >= 18 {
						fmt.Printf("gid: %v\r\n", gid)
					}
					rdms := time.Duration(rand.Intn(10)) * time.Millisecond
					time.Sleep(rdms)
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
			active := p.GetCurrentActiveCount()
			// fmt.Printf("active := pool.GetCurrentActiveCount()\r\n")
			count := p.GetGoroutineCount()
			// fmt.Printf("count := pool.GetGoroutineCount()\r\n")
			peak := p.GetGoroutinePeak()
			// fmt.Printf("peak := pool.GetGoroutinePeak()\r\n")
			settle := p.GetStatusSettle()
			
			settleMap := map[string]string{}
			for i := range settle {
				status := i.ToString()
				duration := fmt.Sprintf("%v", settle[i])
				settleMap[status] = duration
			}
			
			fmt.Printf("active: %v  count: %v  peak: %v  channel: %v  active_duration: %v   sleep_duration: %v \r\n",
				active,
				count,
				peak,
				p.GetWorkCount(),
				settleMap["active"],
				settleMap["sleep"],
			)
			
			// fmt.Printf("%s\r\n",
			// 	utils.ToJson(settleMap),
			// )
			
			_ = active
			_ = count
			_ = peak
			_ = settle
			
		}
	}()
	
	go func() {
		time.Sleep(time.Second*20)
		p.Stop()
		p.Stop()
		p.Stop()
		fmt.Printf("关闭组件\r\n")
	}()
	
	// go func() {
	// 	for {
	// 		time.Sleep(time.Second)
	// 		fmt.Printf("channel: %v\r\n", len(pool.c))
	// 	}
	// }()
	
	time.Sleep(time.Hour)
}
