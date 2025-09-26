package utils

import (
	"fmt"
	"time"
)

// RetryWork 执行工作函数，捕获panic并最多重试retryTimes次
// work: 需要执行的工作函数
// retryTimes: 最大重试次数（不包括首次执行）
func RetryWork(work func(), retryTimes int) {
	// 总执行次数 = 1（首次） + retryTimes（重试）
	totalAttempts := retryTimes + 1
	
	for attempt := 0; attempt < totalAttempts; attempt++ {
		// 使用匿名函数封装单次执行，便于捕获panic
		func(attempt int) {
			defer func() {
				if r := recover(); r != nil {
					// 如果不是最后一次尝试，则提示将重试
					if attempt < totalAttempts-1 {
						fmt.Printf("尝试 %d 发生panic: %v，将进行重试...\n", attempt+1, r)
					} else {
						fmt.Printf("最后一次尝试 %d 发生panic: %v，已耗尽所有重试次数\n", attempt+1, r)
					}
				}
			}()
			
			// 执行工作函数
			work()
			
			// 如果执行到这里，说明没有panic，直接返回
			fmt.Printf("尝试 %d 成功执行\n", attempt+1)
			// 使用label跳出外层循环
			attempt = totalAttempts // 强制退出外层循环
		}(attempt)
		
		// 如果已经是最后一次尝试，不再等待
		if attempt < totalAttempts-1 {
			// 简单的退避等待，每次重试间隔1秒
			time.Sleep(1 * time.Second)
		}
	}
}