package utils

import (
	"fmt"
	"time"
)

// RetryWork 执行工作函数，捕获panic或error并最多重试retryTimes次
// work: 需要执行的工作函数
// retryTimes: 最大重试次数（不包括首次执行）
func RetryWork(work func() error, retryTimes int) {
	totalAttempts := retryTimes + 1
	for attempt := 0; attempt < totalAttempts; attempt++ {
		var err error
		func(attempt int) {
			defer func() {
				if r := recover(); r != nil {
					// 捕获panic，转换为错误
					err = fmt.Errorf("panic: %v", r)
				}
			}()
			// 执行工作函数并获取返回的错误
			err = work()
		}(attempt)

		// 判断是否需要重试
		if err == nil {
			LogMessage(fmt.Sprintf("尝试 %d 成功", attempt+1))
			return // 成功，退出重试
		}

		LogMessage(fmt.Sprintf("业务逻辑出现error/panic: %s", err.Error()))

		// 失败处理
		if attempt < totalAttempts-1 {
			LogMessage(fmt.Sprintf("尝试 %d 失败: %v，将重试...", attempt+1, err))
			time.Sleep(500 * time.Millisecond)
		} else {
			LogMessage(fmt.Sprintf("最后一次尝试 %d 失败: %v，已耗尽重试次数", attempt+1, err))
		}
	}
}
