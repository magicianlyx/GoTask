package pool

import (
	"GoTask/utils"
	"fmt"
	"testing"
	"time"
)

func TestNewStatusSwitch(t *testing.T) {
	rr := NewRecentRecord(time.Millisecond *10200)
	none := GoroutineStatusNone
	active := GoroutineStatusActive
	sleep := GoroutineStatusSleep
	
	time.Sleep(time.Millisecond * 200)      // none+=2
	rr.AddSwitchRecord(none, active) // active+=2
	time.Sleep(time.Millisecond * 200)
	rr.AddSwitchRecord(active, sleep) // sleep+=1
	time.Sleep(time.Millisecond * 100)
	rr.AddSwitchRecord(sleep, active) // active+=3
	time.Sleep(time.Millisecond * 300)
	rr.AddSwitchRecord(active, sleep) // sleep+=1
	time.Sleep(time.Millisecond * 100)
	rr.AddSwitchRecord(sleep, active) // active+=1
	time.Sleep(time.Millisecond * 100)
	rr.AddSwitchRecord(active, sleep) // sleep+=4
	time.Sleep(time.Millisecond * 400)
	rr.AddSwitchRecord(sleep, active) // active+=1
	time.Sleep(time.Millisecond * 100)
	
	// sleep = 1+1+4 = 6
	// active = 2+3+1+1 = 7
	
	settle := rr.GetRecentSettle()
	settleMap := map[string]string{}
	for i := range settle {
		status := i.ToString()
		duration := fmt.Sprintf("%v", settle[i])
		settleMap[status] = duration
	}
	fmt.Printf("%s\r\n",
		utils.ToJson(settleMap),
	)
}
