package mutex

import "time"

/*
Mutex： 同時間淨係得一個thread可以訪問， 相當於capacity = 1 嘅semaphore
*/

// Mutex 是一个互斥锁，同一时间仅允许一个线程访问共享资源。
// 它的实现基于容量为 1 的通道，类似于信号量。
type Mutex struct {
	container chan struct{} // 用于实现互斥的通道
}

// NewMutex 创建一个新的 Mutex。
// 返回一个初始化的互斥锁实例。
func NewMutex() *Mutex {
	return &Mutex{
		container: make(chan struct{}, 1),
	}
}

// Lock 获取互斥锁。
// 如果锁已被占用，则当前 goroutine 会阻塞直到锁被释放。
func (mutex *Mutex) Lock() {
	mutex.container <- struct{}{}
}

// Unlock 释放互斥锁。
// 调用此方法会解除对锁的占用，允许其他 goroutine 获取锁。
func (mutex *Mutex) Unlock() {
	<-mutex.container
}

// Example 展示了 Mutex 的使用示例。
// 创建一个 Mutex 并在多个 goroutine 中使用，确保同一时间只有一个 goroutine 执行关键代码。
func Example() {
	mutex := NewMutex()

	for {
		mutex.Lock()
		go func() {
			println(time.Now().Second())
			time.Sleep(3 * time.Second)
			mutex.Unlock()
		}()
	}
}
