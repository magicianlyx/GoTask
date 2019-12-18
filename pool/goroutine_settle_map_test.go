package pool

import (
	"github.com/magicianlyx/GoTask/utils"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestLatencyParallel(t *testing.T) {
	active := GoroutineStatusActive
	l := NewLatency(active)
	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			l.Start()
		}
	}()
	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			l.Stop()
		}
	}()
	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			_ = l.Clone()
		}
	}()
	
	func() {
		for {
			fmt.Printf("%v\r\n", l.AmountDurationOfNow().amount)
			fmt.Printf("%v\r\n", l.IsStart())
		}
	}()
}

func TestNewLatency(t *testing.T) {
	active := GoroutineStatusActive
	l := NewLatency(active)
	l.Start()
	fmt.Printf("%v\r\n", l.IsStart())
	l.Stop()
	fmt.Printf("%v\r\n", l.IsStart())
	fmt.Printf("%v\r\n", l.AmountDurationOfNow().amount)
}

func TestNewLatencyMapParallel(t *testing.T) {
	active := GoroutineStatusActive
	sleep := GoroutineStatusSleep
	status := []GoroutineStatus{active, sleep}
	randStatus := func() GoroutineStatus {
		rd := rand.Intn(2)
		return status[rd]
	}
	
	m := NewLatencyMap()
	
	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			m.Start(randStatus())
		}
	}()
	
	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			m.Stop(randStatus())
		}
	}()
	
	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			_ = m.Clone()
		}
	}()
	
	func() {
		for {
			time.Sleep(500 * time.Millisecond)
			
			// gsd := m.GetAll()
			// settleMap := map[string]string{}
			// for i := range gsd {
			// 	status := i.ToString()
			// 	duration := fmt.Sprintf("%v", gsd[i])
			// 	settleMap[status] = duration
			// }
			// fmt.Printf("%s\r\n",
			// 	utils.ToJson(settleMap),
			// )
			
			gsd := m.Clone()
			settleMap := map[string]string{}
			for i := range gsd.m {
				status := i.ToString()
				duration := fmt.Sprintf("%v", gsd.m[i].amount)
				settleMap[status] = duration
			}
			fmt.Printf("%s\r\n",
				utils.ToJson(settleMap),
			)
			
		}
	}()
	
}

func TestNewLatencyMap(t *testing.T) {
	
	active := GoroutineStatusActive
	sleep := GoroutineStatusSleep
	
	m := NewLatencyMap()
	m.Start(active)
	m.Start(sleep)
	time.Sleep(time.Millisecond * 200)
	m.Start(active)
	m.Start(sleep)
	time.Sleep(time.Millisecond * 200)
	m.Stop(active)
	m.Stop(sleep)
	time.Sleep(time.Millisecond * 200)
	m.Stop(active)
	m.Stop(sleep)
	time.Sleep(time.Millisecond * 200)
	m.Start(active)
	m.Start(sleep)
	time.Sleep(time.Millisecond * 200)
	m.Stop(active)
	m.Stop(sleep)
	
	gsd := m.GetAll()
	settleMap := map[string]string{}
	for i := range gsd {
		status := i.ToString()
		duration := fmt.Sprintf("%v", gsd[i])
		settleMap[status] = duration
	}
	fmt.Printf("%s\r\n",
		utils.ToJson(settleMap),
	)
}

func TestNewStatusSettle(t *testing.T) {
	active := GoroutineStatusActive
	activeSettle := NewStatusSettle(active, 0)
	activeSettle.AddDuration(time.Millisecond * 300)
	fmt.Printf("active: %v\r\n", activeSettle.GetDuration())
	activeSettle.AddDuration(time.Millisecond * 300)
	fmt.Printf("active: %v\r\n", activeSettle.GetDuration())
	activeSettle.AddDuration(time.Millisecond * 300)
	fmt.Printf("active: %v\r\n", activeSettle.GetDuration())
	activeSettle.AddDuration(time.Millisecond * 300)
	fmt.Printf("active: %v\r\n", activeSettle.GetDuration())
}


func TestNewStatusSettleParallel(t *testing.T) {
	active := GoroutineStatusActive
	sleep := GoroutineStatusSleep
	
	activeSettle := NewStatusSettle(active, 0)
	sleepSettle := NewStatusSettle(sleep, 0)
	
	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			activeSettle.AddDuration(time.Duration(rand.Intn(1000)) * time.Millisecond)
		}
	}()
	
	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			sleepSettle.AddDuration(time.Duration(rand.Intn(1000)) * time.Millisecond)
		}
	}()
	
	func() {
		for {
			time.Sleep(500 * time.Millisecond)
			gsd := activeSettle.GetDuration()
			fmt.Printf("active: %v\r\n", gsd)
			gsd = sleepSettle.GetDuration()
			fmt.Printf("sleep: %v\r\n", gsd)
			
		}
	}()
}

func TestNewGoroutineSettle(t *testing.T) {
	unit := time.Millisecond * 100
	gs := NewGoroutineSettle(unit * 12)
	
	gs.AutoSwitchGoRoutineStatus() // active+=2
	time.Sleep(time.Millisecond * 200)
	gs.AutoSwitchGoRoutineStatus() // sleep+=1
	time.Sleep(time.Millisecond * 100)
	gs.AutoSwitchGoRoutineStatus() // active+=3
	time.Sleep(time.Millisecond * 300)
	gs.AutoSwitchGoRoutineStatus() // sleep+=1
	time.Sleep(time.Millisecond * 100)
	gs.AutoSwitchGoRoutineStatus() // active+=1
	time.Sleep(time.Millisecond * 100)
	gs.AutoSwitchGoRoutineStatus() // sleep+=4
	time.Sleep(time.Millisecond * 400)
	gs.AutoSwitchGoRoutineStatus() // active+=1
	time.Sleep(time.Millisecond * 100)
	
	gsd := gs.GetStatusSettle()
	settleMap := map[string]string{}
	for i := range gsd {
		status := i.ToString()
		duration := fmt.Sprintf("%v", gsd[i])
		settleMap[status] = duration
	}
	fmt.Printf("%s\r\n",
		utils.ToJson(settleMap),
	)
}
