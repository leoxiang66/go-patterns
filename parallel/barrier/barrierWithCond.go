package barrier

import "sync"

// LightBarrier 使用 sync.Cond 和计数器实现的同步机制。
type LightBarrier struct {
	mu    sync.Mutex
	cond  *sync.Cond
	count int
}

// NewLightBarrier 创建一个新的 BarrierWithCond。
func NewLightBarrier() *LightBarrier {
	b := &LightBarrier{
		count: 0,
	}
	b.cond = sync.NewCond(&b.mu)
	return b
}

// Done 表示一个 worker 已经完成。
func (b *LightBarrier) Done() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.count--
	if b.count <= 0 {
		b.cond.Broadcast() // 唤醒所有等待的 goroutine
	}
}

// Sync 等待所有 worker 完成。
func (b *LightBarrier) Sync() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for b.count > 0 {
		b.cond.Wait() // 阻塞直到被唤醒
	}
}

// Add 增加一个 worker。
func (b *LightBarrier) Add() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.count++
}
