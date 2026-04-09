package adapter

// Pool defines worker task pool interface Pool 定义工作任务池接口
type Pool interface {
	// Submit submits task for async execution Submit 提交一个任务以异步执行
	Submit(task func()) error
	// Stop stops pool and releases resources Stop 停止池并释放所有资源
	Stop()
	// Stats returns runtime statistics Stats 返回运行时统计信息
	Stats() (running, capacity int, usage float64)
}
