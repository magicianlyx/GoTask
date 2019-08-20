package profile

import (
	"sync"
	"time"
)

// 执行记录
type ExecuteRecord struct {
	StartTime  time.Time // 任务执行开始时间
	EndTime    time.Time // 任务执行完毕时间
	ElapseTime int       // 执行消耗时间 单位毫秒
	Key        string    // 执行的任务键
}

func (e *ExecuteRecord) Clone() *ExecuteRecord {
	if e == nil {
		return nil
	}
	return &ExecuteRecord{
		e.StartTime,
		e.EndTime,
		e.ElapseTime,
		e.Key,
	}
}

// 任务执行记录队列
type ExecuteRecordQueue struct {
	l    sync.RWMutex
	list []*ExecuteRecord
	size int
}

func NewExecuteRecordQueue(size int) *ExecuteRecordQueue {
	if size == 0 {
		size = 10
	}
	return &ExecuteRecordQueue{
		sync.RWMutex{},
		[]*ExecuteRecord{},
		size,
	}
}

func (q *ExecuteRecordQueue) Push(v *ExecuteRecord) {
	q.l.Lock()
	defer q.l.Unlock()
	if len(q.list) >= q.size {
		q.list = q.list[1:]
	}
	q.list = append(q.list, v)
}

func (q *ExecuteRecordQueue) Pop() *ExecuteRecord {
	q.l.Lock()
	defer q.l.Unlock()

	if len(q.list) >= 1 {
		o := q.list[0]
		q.list = q.list[1:]
		return o
	} else {
		return nil
	}
}

func (q *ExecuteRecordQueue) Peek() *ExecuteRecord {
	q.l.RLock()
	defer q.l.RUnlock()
	if len(q.list) >= 1 {
		return q.list[0]
	} else {
		return nil
	}
}

func (q *ExecuteRecordQueue) Clone() *ExecuteRecordQueue {
	if q == nil {
		return nil
	}
	q.l.RLock()
	defer q.l.RUnlock()
	list := make([]*ExecuteRecord, 0)
	for _, er := range q.list {
		list = append(list, er.Clone())
	}
	return &ExecuteRecordQueue{list: list}
}

const (
	GoroutineStatusSleep  = "Sleep"
	GoroutineStatusActive = "Active"
)

// 线程信息
type GoroutineInfo struct {
	l           sync.RWMutex
	ID          int                 // 线程ID
	Status      string              // 当前状态
	Key         string              // 正在执行的任务key
	startTime   time.Time           // 任务开始时间 用来统计任务占用线程时间
	StartTime   time.Time           // 线程启动时间
	BusyTime    int                 // 忙碌时间 单位毫秒
	LastNRecord *ExecuteRecordQueue // 线程最后n个执行记录
	TaskCount   int                 // 执行任务数
}

func (gi *GoroutineInfo) Clone() *GoroutineInfo {
	if gi == nil {
		return nil
	}
	return &GoroutineInfo{
		ID:          gi.ID,
		Status:      gi.Status,
		Key:         gi.Key,
		StartTime:   gi.StartTime,
		LastNRecord: gi.LastNRecord.Clone(),
		BusyTime:    gi.BusyTime,
		TaskCount:   gi.TaskCount,
	}
}

type Monitor struct {
	GoroutineInfoList []*GoroutineInfo
}

func NewMonitor(routineCount int) *Monitor {
	gis := make([]*GoroutineInfo, routineCount)
	for i := range gis {
		gis[i] = &GoroutineInfo{
			sync.RWMutex{},
			i,
			GoroutineStatusSleep,
			"",
			time.Time{},
			time.Now(),
			0,
			NewExecuteRecordQueue(10),
			0,
		}
	}
	return &Monitor{gis}
}

func (m *Monitor) SetGoroutineSleep(id int) {
	gi := m.GoroutineInfoList[id]
	gi.l.Lock()
	defer gi.l.Unlock()
	// 计算刚执行完毕的任务总结
	now := time.Now()
	elapseTime := int(now.Sub(gi.startTime).Nanoseconds() / 1e6)
	gi.LastNRecord.Push(&ExecuteRecord{
		gi.startTime,
		now,
		elapseTime,
		gi.Key,
	})

	gi.Status = GoroutineStatusSleep
	gi.Key = ""
	gi.TaskCount = gi.TaskCount + 1
	gi.BusyTime = gi.BusyTime + elapseTime
	gi.startTime = time.Time{}
}

func (m *Monitor) SetGoroutineRunning(id int, key string) {
	gi := m.GoroutineInfoList[id]
	gi.l.Lock()
	defer gi.l.Unlock()
	gi.Status = GoroutineStatusActive
	gi.Key = key
	gi.startTime = time.Now()
}

func (m *Monitor) Clone() *Monitor {
	if m == nil {
		return nil
	}
	gis := make([]*GoroutineInfo, 0)
	for _, gi := range m.GoroutineInfoList {
		gis = append(gis, gi.Clone())
	}
	return &Monitor{
		gis,
	}
}

func (m *Monitor) GetGoroutineStatus(id int) *GoroutineInfo {
	return m.GoroutineInfoList[id].Clone()
}

func (m *Monitor) GetAllGoroutineStatus() *Monitor {
	return m.Clone()
}
