package barrier

import "time"

// EasyBarrier 是一个简单的同步机制，允许多个 goroutine 协调并等待彼此完成。
type EasyBarrier struct {
	num_workers int           // 当前需要同步的 worker 数量
	ready       chan struct{} // 用于同步的通道
}

// NewEasyBarrier 创建一个新的 EasyBarrier。
// 参数 num_workers 指定需要同步的 worker 数量。
func NewEasyBarrier(num_workers int) *EasyBarrier {
	return &EasyBarrier{
		num_workers: num_workers,
		ready:       make(chan struct{}, num_workers),
	}
}

// Done 表示一个 worker 已经完成。
// 调用此方法会向 ready 通道发送一个信号。
func (barrier *EasyBarrier) Done() {
	barrier.ready <- struct{}{}
}

// Sync 等待所有 worker 完成。
// 调用此方法会阻塞，直到所有 worker 都调用了 Done。
func (barrier *EasyBarrier) Sync() {
	for i := 0; i < barrier.num_workers; i++ {
		<-barrier.ready
	}
}

// Example 展示了 EasyBarrier 的使用示例。
// 创建一个 EasyBarrier 并启动多个 goroutine，等待它们完成后继续执行。
func Example() {
	barrier := NewEasyBarrier(5)
	for i := 0; i < 5; i++ {
		go func(i int) {
			println("goroutine ", i, " starts!")
			time.Sleep(time.Duration((i+1)*3) * time.Second)
			println("goroutine ", i, " finishes!")
			barrier.Done()
		}(i)
	}

	barrier.Sync()
	println("All goroutine finish!")
}
