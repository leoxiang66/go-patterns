package pipeline

import (
	"fmt"
	"sync"
)

// FanIn 多通道合并：将多个<-chan X类型的输入通道，合并到一个chan X输出通道
func FanIn[X any](q <-chan struct{}, inputs ...<-chan X) chan X {
    out := make(chan X)
    var wg sync.WaitGroup

    for _, in := range inputs {
        wg.Add(1)
        // 直接操作输入通道，无需经过AddOnPipe
        go func(input <-chan X) {
            defer wg.Done()
            for {
                select {
                case <-q: // 响应终止信号
                    return
                case x, ok := <-input: // 直接从输入通道读取
                    if !ok { // 输入通道关闭，退出当前goroutine
                        return
                    }
                    // 发送到输出通道（同时监听终止信号）
                    select {
                    case <-q:
                        return
                    case out <- x:
                    }
                }
            }
        }(in) // 注意：循环中传递当前in的副本，避免闭包捕获变量问题
    }

    // 等待所有输入处理完毕后关闭输出通道
    go func() {
        wg.Wait()
        close(out)
        fmt.Println("FanIn: 所有输入处理完毕，关闭输出通道")
    }()

    return out
}


// FanOut 数据分发：将单一输入通道的数据复制到多个输出通道
// 参数：
//   q: 终止信号通道（关闭时组件退出）
//   in: 输入通道（单一数据源）
//   num: 输出通道数量
// 返回：
//   输出通道切片（每个通道都会收到输入的完整数据副本）
func FanOut[X any](q <-chan struct{}, in <-chan X, num int) []chan X {
    if num <= 0 {
        panic("FanOut: 输出通道数量必须大于0")
    }

    // 创建指定数量的输出通道
    outs := make([]chan X, num)
    for i := range outs {
        outs[i] = make(chan X)
    }

    var wg sync.WaitGroup
    wg.Add(num) // 每个输出通道对应一个goroutine

    // 为每个输出通道启动goroutine，负责复制数据
    for i := 0; i < num; i++ {
        go func(outChan chan<- X) {
            defer func() {
                wg.Done()
                close(outChan) // 确保每个输出通道最终会关闭
            }()

            // 从输入通道读取数据，复制到当前输出通道
            for {
                select {
                case <-q: // 响应终止信号
                    return
                case x, ok := <-in: // 读取输入数据
                    if !ok { // 输入通道关闭，退出
                        return
                    }
                    // 发送数据到当前输出通道（同时监听终止信号）
                    select {
                    case <-q:
                        return
                    case outChan <- x:
                    }
                }
            }
        }(outs[i])
    }

    // 可选：等待所有输出goroutine结束后打印日志（非必需）
    go func() {
        wg.Wait()
        fmt.Printf("FanOut: 所有%d个输出通道已处理完毕并关闭\n", num)
    }()

    return outs
}
