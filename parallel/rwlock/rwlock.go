package rwlock

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// RWLock 是一个写优先的写锁，允许多个读者同时访问，但写者需要独占访问。
// 它通过通道和互斥锁实现，支持高效的读写操作。
type RWLock struct {
	readers  int32         // 当前活跃的读者数量
	noreader chan struct{} // 用于指示是否有读者的通道
	writer   chan struct{} // 用于控制写者访问的通道
}

// NewRWLock 创建一个新的 RWLock。
// 返回一个初始化的读写锁实例。
func NewRWLock() *RWLock {
	sem := make(chan struct{}, 1)
	noreader := make(chan struct{}, 1)
	noreader <- struct{}{}
	return &RWLock{
		writer:   sem,
		noreader: noreader,
	}
}

// RLock 获取读锁。
// 如果当前有写者正在访问，调用此方法的 goroutine 会阻塞。
func (rwlock *RWLock) RLock() {
	rwlock.writer <- struct{}{}
	if atomic.LoadInt32(&rwlock.readers) == 0 {
		<-rwlock.noreader // 第一个读者抢走 noreader token
	}
	atomic.AddInt32(&rwlock.readers, 1)
	<-rwlock.writer
}

// RUnlock 释放读锁。
// 如果这是最后一个读者，会通知等待的写者。
func (rwlock *RWLock) RUnlock() {
	n := atomic.AddInt32(&rwlock.readers, -1)
	if n == 0 {
		rwlock.noreader <- struct{}{}
	}
	
}

// WLock 获取写锁。
// 如果当前有读者或写者正在访问，调用此方法的 goroutine 会阻塞。
func (rw *RWLock) WLock() {
	rw.writer <- struct{}{}
	<-rw.noreader
}

// WUnlock 释放写锁。
// 允许其他读者或写者继续访问。
func (rw *RWLock) WUnlock() {
	rw.noreader <- struct{}{}
	<-rw.writer
}

// Example 展示了 RWLock 的使用示例。
// 启动多个读者和一个写者，演示读写锁的行为。
func Example() {
	var wg sync.WaitGroup
	lock := NewRWLock()
	count := 0

	// 启动多个读者
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			lock.RLock()
			fmt.Printf("Reader %d reading %d...\n", id, count)
			time.Sleep(1000 * time.Millisecond)
			lock.RUnlock()
		}(i)
	}

	// 启动一个写者
	wg.Add(1)
	go func() {
		defer wg.Done()
		lock.WLock()
		fmt.Println("Writer writing...")
		count++
		time.Sleep(1000 * time.Millisecond)
		lock.WUnlock()
	}()

	wg.Wait()
}
