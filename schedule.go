package GoTaskv1

import (
	"time"
)

type ISchedule interface {
	expression(t *TaskInfo) (nt time.Time, isValid bool) // 表达式
	record(t *TaskInfo, cs ...int) (et []time.Time)      // 索引执行记录 返回执行时刻 time为zero时说明还没被执行
}

type SpecSchedule struct {
	spec time.Duration
}

func NewSpecSchedule(spec time.Duration) *SpecSchedule {
	return &SpecSchedule{spec: spec}
}

func (p *SpecSchedule) expression(t *TaskInfo) (nt time.Time, isValid bool) {
	return t.AddTime.Add(time.Duration(t.Count+1) * p.spec), true
}

func (p *SpecSchedule) record(t *TaskInfo, cs ...int) (et []time.Time) {
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

type SpecTimeSchedule struct {
	spec time.Duration
	time int
}

func NewSpecTimeSchedule(spec time.Duration, time int) *SpecTimeSchedule {
	return &SpecTimeSchedule{spec: spec, time: time}
}

func (p *SpecTimeSchedule) expression(t *TaskInfo) (nt time.Time, isValid bool) {
	if t.Count >= p.time {
		return time.Time{}, false
	}
	nt = t.AddTime.Add(time.Duration(t.Count+1) * p.spec)
	isValid = true
	return
}

func (p *SpecTimeSchedule) record(t *TaskInfo, cs ...int) (et []time.Time) {
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

type PlanSchedule struct {
	tList []time.Time // 计划任务时间点 有次数限制
}

func NewPlanSchedule(tList []time.Time) *PlanSchedule {
	return &PlanSchedule{tList: tList}
}

func (p *PlanSchedule) expression(t *TaskInfo) (nt time.Time, isValid bool) {
	if p.tList == nil || len(p.tList) == 0 || t.Count >= len(p.tList) {
		return time.Time{}, false
	} else {
		return p.tList[t.Count], true
	}
}

func (p *PlanSchedule) record(t *TaskInfo, cs ...int) (et []time.Time) {
	et = make([]time.Time, 0)
	for i := range cs {
		c := cs[i]
		if c <= t.Count && c > 0 {
			et = append(et, p.tList[c-1])
		}
		et = append(et, time.Time{})
	}
	return et
}

//type ExpressionSchedule struct {
//}
//
//// 任务计划
//type Schedule struct {
//	spec       time.Duration                       // 间隔时间
//	tList      []time.Time                         // 计划任务时间点 有次数限制
//	expression func(t *TaskInfo) (nt time.Time,isValid bool)  // 根据执行结果获取下一次执行时间的表达式
//}
