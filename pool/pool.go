package pool

import "time"

// 任务函数 gid为执行该任务的线程id
type TaskObj func(gid int)

type GoroutinePool struct {
	c         chan TaskObj
	mainClose chan struct{} // 停止所有线程信号
	M         *DynamicPoolMonitor
	options   *Options
}

func NewGoroutinePool(options *Options) *GoroutinePool {
	options = options.Clone()
	options.fillDefaultOptions()
	return &GoroutinePool{
		c:       make(chan TaskObj, options.TaskChannelSize),
		M:       NewDynamicPoolMonitor(),
		options: options,
	}
}

func (g *GoroutinePool) Put(obj TaskObj) {
	g.c <- obj
	if g.M.GetCurrentActiveCount() == 0 {
		// 无活动线程时新建线程
		g.tryNewGoroutine()
	} else if (float64(len(g.c))/float64(g.options.TaskChannelSize)) > g.options.NewGreaterThanF && g.M.GetGoroutineCount() < g.options.GoroutineLimit {
		// 必要时新建线程
		g.tryNewGoroutine()
	}
}
func (g *GoroutinePool) Stop() {
	close(g.mainClose)
}


// FIXME 此处需要改进
//  对监控结构体M的创建新线程 在结构体内部的方法去进行限制 判断和创建应为原子操作
//  同时销毁线程的逻辑也需要做这样的处理
// 新建一个线程
func (g *GoroutinePool) tryNewGoroutine() (int, chan<- struct{}) {
	// FIXME 需改进
	gid := g.M.Construct()
	close := make(chan struct{})
	t := time.NewTicker(g.options.AutoMonitorDuration)
	go func(gid int) {
		for {
			select {
			case task := <-g.c: // 执行任务task
				g.M.Switch(gid)
				task(gid)
				g.M.Switch(gid)
			case <-close: // 主动关闭
				g.M.Destroy(gid)
				return
			case <-t.C: // 定时检测活跃时长比例 条件关闭
				t.Stop()
				t = time.NewTicker(g.options.AutoMonitorDuration)
				// FIXME 需改进
				if g.M.GetRecentActiveRatio(gid) < g.options.CloseLessThanF {
					g.M.Destroy(gid)
					return
				}
			case <-g.mainClose: // 主线程退出信号
				g.M.Destroy(gid)
				return
			}
		}
	}(gid)
	t.Stop()
	return gid, close
}
//
//// 新建一个线程
//func (g *GoroutinePool) newGoroutine() (int, chan<- struct{}) {
//	gid := g.M.Construct()
//	close := make(chan struct{})
//	t := time.NewTicker(g.options.AutoMonitorDuration)
//	go func(gid int) {
//		for {
//			select {
//			case task := <-g.c: // 执行任务task
//				g.M.Switch(gid)
//				task(gid)
//				g.M.Switch(gid)
//			case <-close: // 主动关闭
//				g.M.Destroy(gid)
//				return
//			case <-t.C: // 定时检测活跃时长比例 条件关闭
//				t.Stop()
//				t = time.NewTicker(g.options.AutoMonitorDuration)
//				if g.M.GetRecentActiveRatio(gid) < g.options.CloseLessThanF {
//					g.M.Destroy(gid)
//					return
//				}
//			case <-g.mainClose: // 主线程退出信号
//				g.M.Destroy(gid)
//				return
//			}
//		}
//	}(gid)
//	t.Stop()
//	return gid, close
//}
