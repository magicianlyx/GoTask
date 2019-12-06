package pool

import "time"

func Timer() func() (start, end time.Time, duration time.Duration) {
	p1 := time.Now()
	return func() (start, end time.Time, duration time.Duration) {
		start = p1
		end = time.Now()
		return start, end, end.Sub(start)
	}
}
