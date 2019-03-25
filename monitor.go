package GoTaskv1

import (
	"time"
	"sync"
)

// 执行记录
type ExecuteRecord struct {
	StartTime  time.Time
	EndTime    time.Time
	ElapseTime int // 单位毫秒
}

func (e *ExecuteRecord) Clone() *ExecuteRecord {
	return &ExecuteRecord{
		e.StartTime,
		e.EndTime,
		e.ElapseTime,
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
	GoroutineStatusSleep  = 0
	GoroutineStatusActive = 1
)

type GoroutineInfo struct {
	l           sync.RWMutex
	ID          int                 // 线程ID
	Status      int                 // 当前状态
	Key         string              // 正在执行的任务key
	StartTime   time.Time           // 任务开始时间
	LastNRecord *ExecuteRecordQueue // 线程最后n个执行记录
}

func (gi *GoroutineInfo) Clone() *GoroutineInfo {
	return &GoroutineInfo{
		ID:          gi.ID,
		Status:      gi.Status,
		Key:         gi.Key,
		StartTime:   gi.StartTime,
		LastNRecord: gi.LastNRecord.Clone(),
	}
}

type Monitor struct {
	Gis []*GoroutineInfo
}

func NewMonitor(count int) *Monitor {
	gis := make([]*GoroutineInfo, count)
	for i, _ := range gis {
		gis[i] = &GoroutineInfo{
			sync.RWMutex{},
			i,
			GoroutineStatusSleep,
			"",
			time.Time{},
			NewExecuteRecordQueue(10),
		}
	}
	return &Monitor{gis}
}

func (m *Monitor) SetGoroutineSleep(id int) {
	gis := m.Gis[id]
	gis.l.Lock()
	defer gis.l.Unlock()
	gis.Status = GoroutineStatusSleep
	gis.Key = ""
	now := time.Now()
	gis.LastNRecord.push(&ExecuteRecord{
		gis.StartTime,
		now,
		int(now.Sub(gis.StartTime).Nanoseconds() / 1e6),
	})
}

func (m *Monitor) SetGoroutineRunning(id int, key string) {
	gis := m.Gis[id]
	gis.l.Lock()
	defer gis.l.Unlock()
	gis.Status = GoroutineStatusActive
	gis.Key = key
	gis.StartTime = time.Now()
}

func (m *Monitor) Clone() *Monitor {
	gis := []*GoroutineInfo{}
	for _, gi := range m.Gis {
		gis = append(gis, gi.Clone())
	}
	
	return &Monitor{
		gis,
	}
}

func (m *Monitor) GetGoroutineStatus(id int) *GoroutineInfo {
	return m.Gis[id].Clone()
}

func (m *Monitor) GetAllGoroutineStatus() *Monitor {
	return m.Clone()
}
