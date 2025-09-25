package pipeline

import (
	"fmt"
	"time"
)

// AddOnPipe 通用管道节点：接收X类型输入，通过函数f转换为Y类型，输出到Y通道；支持通过quit通道终止
// 参数：
//   q: 终止信号通道（关闭时组件退出）
//   f: 数据转换函数（业务逻辑由外部注入，组件不关心具体实现）
//   in: 输入通道（接收X类型数据）
// 返回：
//   输出通道（发送Y类型数据）
func AddOnPipe[X, Y any](q <-chan struct{}, f func(X) Y, in <-chan X) chan Y {
	out := make(chan Y) // 初始化输出通道（无缓冲，确保同步性；也可根据需求设为带缓冲）

	go func() {
		// 关键：defer关闭输出通道，避免下游永久阻塞
		defer close(out)

		for {
			select {
			// 1. 监听终止信号：quit关闭则退出循环
			case <-q:
				fmt.Println("AddOnPipe: 收到终止信号，退出")
				return

			// 2. 接收输入数据，处理后输出
			case input, ok := <-in:
				// 若输入通道关闭（上游无数据），则退出循环
				if !ok {
					fmt.Println("AddOnPipe: 输入通道关闭，退出")
					return
				}
				// 调用外部注入的转换函数f，处理数据
				output := f(input)
				// 将结果发送到输出通道（若下游未接收，会阻塞，符合管道同步特性）
				out <- output
			}
		}
	}()

	return out
}


func example() {
    // 1. 准备输入通道和终止通道
    input := make(chan int)
    quit := make(chan struct{})
    defer close(quit) // 主函数退出时关闭quit，确保所有组件终止
    
    // 2. 定义业务逻辑函数（注入到AddOnPipe）：将int转换为string（模拟“准备托盘”“添加配料”等业务）
    intToString := func(x int) string {
        return fmt.Sprintf("处理后的数据：%d", x)
    }
    
    // 3. 创建管道节点，串联业务逻辑
    pipe := AddOnPipe(quit, intToString, input)
    
    // 4. 启动goroutine：向输入通道发送测试数据
    go func() {
        for i := 0; i < 5; i++ {
            input <- i
            time.Sleep(500 * time.Millisecond) // 模拟数据产生间隔
        }
        close(input) // 数据发送完毕，关闭输入通道
    }()
    
    // 5. 接收管道输出，验证结果
    for output := range pipe {
        fmt.Println("主函数收到：", output)
    }
    fmt.Println("测试完成")
}