package pool

import (
	"runtime"
	"time"
)

type Options struct {
	AutoMonitorDuration time.Duration // 定时check时长
	CloseLessThanF      float64       // 定时check活跃线程比例 小于50%时 会关闭当前线程
	NewGreaterThanF     float64       // 活跃线程比例大于90%时 新任务会创建新线程去跑
	GoroutineLimit      int           // 线程上限数
	TaskChannelSize     int           // 任务channel尺寸
}

// 构建默认配置
func NewDefaultOptions() *Options {
	return &Options{
		AutoMonitorDuration: time.Second * 5 * 60,
		CloseLessThanF:      0.3,
		NewGreaterThanF:     0.001,
		GoroutineLimit:      runtime.NumCPU() * 3,
		TaskChannelSize:     runtime.NumCPU() * 100,
	}
}

// 填充参数
func (o *Options) fillDefaultOptions() {
	oDefault := NewDefaultOptions()
	if o.AutoMonitorDuration == time.Duration(0) {
		o.AutoMonitorDuration = oDefault.AutoMonitorDuration
	}
	if o.CloseLessThanF >= 1.0 || o.CloseLessThanF <= 0.0 {
		o.CloseLessThanF = oDefault.CloseLessThanF
	}
	if o.NewGreaterThanF >= 1.0 || o.NewGreaterThanF <= 0.0 {
		o.NewGreaterThanF = oDefault.NewGreaterThanF
	}
	if o.GoroutineLimit <= 0 {
		o.GoroutineLimit = oDefault.GoroutineLimit
	}
	if o.TaskChannelSize <= 0 {
		o.TaskChannelSize = oDefault.TaskChannelSize
	}
}

func (o *Options) Clone() *Options {
	return &Options{
		o.AutoMonitorDuration,
		o.CloseLessThanF,
		o.NewGreaterThanF,
		o.GoroutineLimit,
		o.TaskChannelSize,
	}
}
