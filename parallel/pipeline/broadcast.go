package pipeline

import "sync"

// Broadcast 广播组件：将输入通道in的消息广播到所有订阅的输出通道
type Broadcast[X any] struct {
	mu          sync.RWMutex    // 保护订阅者列表
	subscribers []chan X        // 订阅者通道列表
	in          <-chan X        // 输入通道
	quit        <-chan struct{} // 终止信号
}

// NewBroadcast 初始化广播组件
func NewBroadcast[X any](q <-chan struct{}, in <-chan X) *Broadcast[X] {
	b := &Broadcast[X]{
		subscribers: make([]chan X, 0),
		in:          in,
		quit:        q,
	}
	return b
}

// Subscribe 订阅广播（返回一个接收广播消息的通道）
func (b *Broadcast[X]) Subscribe() <-chan X {
	ch := make(chan X)
	b.mu.Lock()
	b.subscribers = append(b.subscribers, ch)
	b.mu.Unlock()
	return ch
}

// Run 核心广播逻辑：从输入通道读消息，发送到所有订阅者
func (b *Broadcast[X]) Run() {
	for {
		select {
		case <-b.quit: // 响应终止信号，关闭所有订阅者通道
			b.mu.Lock()
			for _, ch := range b.subscribers {
				close(ch)
			}
			b.subscribers = nil
			b.mu.Unlock()
			return
		case x, ok := <-b.in: // 读取输入消息
			if !ok { // 输入通道关闭，退出广播
				b.mu.Lock()
				for _, ch := range b.subscribers {
					close(ch)
				}
				b.subscribers = nil
				b.mu.Unlock()
				return
			}
			// 广播消息到所有订阅者
			b.mu.RLock()
			for _, ch := range b.subscribers {
				select {
				case <-b.quit:
					b.mu.RUnlock()
					return
				case ch <- x: // 发送消息到订阅者
				}
			}
			b.mu.RUnlock()
		}
	}
}
