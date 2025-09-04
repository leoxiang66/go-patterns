package rwlock

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 并发读写测试：并发多个 reader 和 writer。
// 用 go test -race 可以检测到 RWLock 内部的 data race（如果实现有问题的话）。
func TestRWLock_ConcurrentReadersWriters(t *testing.T) {
	lock := NewRWLock()
	var wg sync.WaitGroup

	// 共享数据，只有写者会修改
	var data int

	const (
		nReaders   = 50
		nWriters   = 10
		iterReader = 100
		iterWriter = 100
	)

	// 启动读者
	for i := 0; i < nReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterReader; j++ {
				lock.RLock()
				_ = data // 读取共享数据
				// 模拟读操作耗时，让并发更容易重叠
				time.Sleep(time.Millisecond)
				lock.RUnlock()
			}
		}()
	}

	// 启动写者
	for i := 0; i < nWriters; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterWriter; j++ {
				lock.WLock()
				data++ // 写操作
				time.Sleep(time.Millisecond) // 模拟写耗时
				lock.WUnlock()
			}
		}()
	}

	wg.Wait()

	// 正确的写次数应等于 nWriters * iterWriter
	expected := nWriters * iterWriter
	if data != expected {
		t.Fatalf("unexpected data: got %d want %d", data, expected)
	}
}

// 调试用并发测试：进度打印 + 超时 watchdog。
// 先用较小规模跑，确认没有死锁，然后再放大规模并加 -race。
func TestRWLock_Debug(t *testing.T) {
	lock := NewRWLock()
	var wg sync.WaitGroup

	var data int64
	var readsDone int64
	var writesDone int64

	const (
		nReaders   = 20
		nWriters   = 5
		iterReader = 50
		iterWriter = 50
	)

	// 监视器：每 500ms 打印一次进度
	stopMonitor := make(chan struct{})
	go func() {
		tick := time.NewTicker(500 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				r := atomic.LoadInt64(&readsDone)
				w := atomic.LoadInt64(&writesDone)
				fmt.Printf("progress: readsDone=%d writesDone=%d\n", r, w)
			case <-stopMonitor:
				return
			}
		}
	}()

	// Watchdog：30s 内未完成就失败（避免长时间挂起）
	done := make(chan struct{})
	go func() {
		select {
		case <-time.After(30 * time.Second):
			t.Fatal("test timed out after 30s (watchdog)")
		case <-done:
			// 正常退出
		}
	}()

	// 启动读者
	for i := 0; i < nReaders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterReader; j++ {
				lock.RLock()
				_ = atomic.LoadInt64(&data)
				// 模拟读耗时：把 sleep 尽量短以减少总时长
				time.Sleep(1 * time.Millisecond)
				lock.RUnlock()
				atomic.AddInt64(&readsDone, 1)
			}
		}(i)
	}

	// 启动写者
	for i := 0; i < nWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterWriter; j++ {
				lock.WLock()
				atomic.AddInt64(&data, 1)
				time.Sleep(1 * time.Millisecond)
				lock.WUnlock()
				atomic.AddInt64(&writesDone, 1)
			}
		}(i)
	}

	wg.Wait()
	close(done)
	close(stopMonitor)

	expected := int64(nWriters * iterWriter)
	if atomic.LoadInt64(&data) != expected {
		t.Fatalf("unexpected data: got %d want %d", data, expected)
	}
}