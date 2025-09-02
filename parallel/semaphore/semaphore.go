package semaphore

import "time"

/*
Semaphore 限制同一時間可以訪問資源嘅數量
*/

// Semaphore 是一个信号量，用于限制同时访问共享资源的 goroutine 数量。
// 它的实现基于带缓冲的通道。
type Semaphore struct {
	container chan struct{} // 用于控制并发访问的通道
}

// NewSemaphore 创建一个新的 Semaphore。
// 参数 capacity 指定信号量的容量，即允许同时访问的最大 goroutine 数量。
func NewSemaphore(capacity int) *Semaphore {
	return &Semaphore{
		container: make(chan struct{}, capacity),
	}
}

// Acquire 获取信号量。
// 如果信号量已满，调用此方法的 goroutine 会阻塞直到有可用的容量。
func (sem *Semaphore) Acquire() {
	sem.container <- struct{}{}
}

func (sem *Semaphore) TryAcquire() bool {
	select {
	case sem.container <- struct{}{}:
		return true
	default:
		return false
	}
}

// Release 释放信号量。
// 调用此方法会释放一个信号量容量，允许其他 goroutine 获取。
func (sem *Semaphore) Release() {
	<-sem.container
}

// Example 展示了 Semaphore 的使用示例。
// 创建一个 Semaphore 并限制同时运行的 goroutine 数量。
func Example() {
	sem := NewSemaphore(5)

	for {
		sem.Acquire()
		go func() {
			println(time.Now().Second())
			time.Sleep(3 * time.Second)
			sem.Release()
		}()
	}
}
