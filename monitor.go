package GoTaskv1

import (
	"time"
	"sync"
)

// 执行记录
type ExecuteRecord struct {
	StartTime  time.Time
	EndTime    time.Time
	ElapseTime int    // 单位毫秒
	Key        string // 执行的任务键
}

func (e *ExecuteRecord) Clone() *ExecuteRecord {
	return &ExecuteRecord{
		e.StartTime,
		e.EndTime,
		e.ElapseTime,
		e.Key,
	}
}

type ExecuteRecordQueue struct {
	l    sync.RWMutex
	List []*ExecuteRecord
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

func (q *ExecuteRecordQueue) push(v *ExecuteRecord) {
	q.l.Lock()
	defer q.l.Unlock()
	if len(q.List) >= q.size {
		q.List = q.List[1:]
		q.List = append(q.List, v)
	} else {
		q.List = append(q.List, v)
	}
}

func (q *ExecuteRecordQueue) pop() *ExecuteRecord {
	q.l.Lock()
	defer q.l.Unlock()
	
	if len(q.List) >= 1 {
		o := q.List[0]
		q.List = q.List[1:]
		return o
	} else {
		return nil
	}
}

func (q *ExecuteRecordQueue) peek() *ExecuteRecord {
	q.l.RLock()
	defer q.l.RUnlock()
	if len(q.List) >= 1 {
		return q.List[0]
	} else {
		return nil
	}
}

func (q *ExecuteRecordQueue) Clone() *ExecuteRecordQueue {
	q.l.RLock()
	defer q.l.RUnlock()
	list := []*ExecuteRecord{}
	for _, er := range q.List {
		list = append(list, er.Clone())
	}
	return &ExecuteRecordQueue{List: list}
}

const (
	GoroutineStatusSleep  = "Sleep"
	GoroutineStatusActive = "Active"
)

type GoroutineInfo struct {
	l           sync.RWMutex
	ID          int                 // 线程ID
	Status      string              // 当前状态
	Key         string              // 正在执行的任务key
	startTime   time.Time           // 任务开始时间
	StartTime   time.Time           // 线程启动时间
	BusyTime    int                 // 忙碌时间 单位毫秒
	LastNRecord *ExecuteRecordQueue // 线程最后n个执行记录
	TaskCount   int                 // 执行任务数
}

func (gi *GoroutineInfo) Clone() *GoroutineInfo {
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
	for i, _ := range gis {
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
	now := time.Now()
	elapseTime := int(now.Sub(gi.startTime).Nanoseconds() / 1e6)
	gi.LastNRecord.push(&ExecuteRecord{
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
	gis := []*GoroutineInfo{}
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
