package pool

//var c chan<- taskobj
//
//func goroutine(n int) (chan<- taskobj) {
//	c := make(chan taskobj, n*5)
//	for i := 0; i < n; i++ {
//		go func() {
//			for {
//				task := <-c
//				task()
//			}
//		}()
//	}
//	return c
//}
//
//func put(task taskobj){
//
//}