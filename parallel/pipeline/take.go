package pipeline

// Take 从输入通道in中截取前n个数据，输出到out通道；截取完成后关闭out
// 参数：
//   q: 终止信号通道（关闭时提前退出）
//   n: 需截取的数据量
//   in: 输入通道（承接FanOut/FanIn的输出）
func Take[X any](q <-chan struct{}, n int, in <-chan X) chan X {
	out := make(chan X)
	count := 0 // 计数已截取的数据量

	go func() {
		defer close(out) // 确保输出通道最终关闭
		for {
			select {
			case <-q: // 响应终止信号，提前退出
				return
			case x, ok := <-in:
				if !ok { // 输入通道关闭，退出
					return
				}
				if count >= n { // 已截取够n个数据，退出
					return
				}
				// 发送数据到输出通道
				select {
				case <-q:
					return
				case out <- x:
					count++ // 计数+1
				}
			}
		}
	}()

	return out
}