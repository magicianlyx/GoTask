package task

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"time"
)

// 任务调度接口
type ISchedule interface {
	Expression(t *TaskInfo) (nt time.Time, isValid bool) // 表达式
	// Record(t *TaskInfo, cs ...int) (et []time.Time)      // 索引执行记录 返回执行时刻 time为zero时说明还没被执行
	ToString() string
}

// 指定时长循环调度
type SpecSchedule struct {
	spec time.Duration
}

func NewSpecSchedule(spec time.Duration) *SpecSchedule {
	return &SpecSchedule{spec: spec}
}

func (p *SpecSchedule) Expression(t *TaskInfo) (nt time.Time, isValid bool) {
	return t.AddTime.Add(time.Duration(t.Count+1) * p.spec), true
}

func (p *SpecSchedule) Record(t *TaskInfo, cs ...int) (et []time.Time) {
	et = make([]time.Time, 0)
	for i := range cs {
		c := cs[i]
		if t.Count >= c {
			et = append(et, t.AddTime.Add(time.Duration(c)*p.spec))
		}
		et = append(et, time.Time{})
	}
	return
}

func (p *SpecSchedule) ToString() string {
	s, _ := jsoniter.MarshalToString(map[string]interface{}{
		"spec": fmt.Sprintf("%.6fs", p.spec.Seconds()),
	})
	return s
}

// 指定时长指定次数调度器
type SpecTimeSchedule struct {
	spec time.Duration
	time int
}

func NewSpecTimeSchedule(spec time.Duration, time int) *SpecTimeSchedule {
	return &SpecTimeSchedule{spec: spec, time: time}
}

func (p *SpecTimeSchedule) Expression(t *TaskInfo) (nt time.Time, isValid bool) {
	if t.Count >= p.time {
		return time.Time{}, false
	}
	nt = t.AddTime.Add(time.Duration(t.Count+1) * p.spec)
	isValid = true
	return
}

func (p *SpecTimeSchedule) ToString() string {
	s, _ := jsoniter.MarshalToString(map[string]interface{}{
		"time": p.time,
		"spec": fmt.Sprintf("%.6fs", p.spec.Seconds()),
	})
	return s
}

// 指定时间点调度器
type PlanSchedule struct {
	tList []time.Time // 计划任务时间点 有次数限制
}

func NewPlanSchedule(tList []time.Time) *PlanSchedule {
	return &PlanSchedule{tList: tList}
}

func (p *PlanSchedule) Expression(t *TaskInfo) (nt time.Time, isValid bool) {
	if p.tList == nil || len(p.tList) == 0 || t.Count >= len(p.tList) {
		return time.Time{}, false
	} else {
		return p.tList[t.Count], true
	}
}

func (p *PlanSchedule) ToString() string {
	s, _ := jsoniter.MarshalToString(map[string]interface{}{
		"tList": p.tList,
	})
	return s
}

// 每日指定时刻调度器
type EveryDaySchedule struct {
	hour    int
	minute  int
	second  int
	mSecond int
}

func NewEveryDaySchedule(hour, minute, second, mSecond int) *EveryDaySchedule {
	return &EveryDaySchedule{hour, minute, second, mSecond}
}

func (e *EveryDaySchedule) Expression(t *TaskInfo) (nt time.Time, isValid bool) {
	now := time.Now()
	nt = time.Date(now.Year(), now.Month(), now.Day(), e.hour, e.minute, e.second, e.mSecond, time.Local)
	if nt.UnixNano() < now.UnixNano() {
		return nt.AddDate(0, 0, 1), true
	}
	return nt, true
}

func (e *EveryDaySchedule) ToString() string {
	return fmt.Sprintf("every day %d h %d m %d s %d ms", e.hour, e.minute, e.second, e.mSecond)
}
