package task

import "time"

type TimerObj func() (spec time.Duration, start, end time.Time)

// 计时器
func Timer() TimerObj {
	p1 := time.Now()
	return func() (spec time.Duration, start, end time.Time) {
		p2 := time.Now()
		spec = time.Now().Sub(p1)
		start = p1
		end = p2
		return
	}
}
