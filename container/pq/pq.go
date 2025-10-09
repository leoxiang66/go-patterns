package pq

import (
	"errors"
)

// PriorityQueue 是一个有容量限制的优先队列，支持泛型
type PriorityQueue[T any] struct {
	data     []T
	capacity int
	better   func(a, b T) bool // 比较函数，返回true表示a应该排在b前面
}

// NewPriorityQueue 创建一个新的优先队列
// capacity为0表示没有容量限制
// 时间复杂度: O(1)
func NewPriorityQueue[T any](capacity int, better func(a, b T) bool) (*PriorityQueue[T], error) {
	if capacity < 0 {
		return nil, errors.New("capacity must be non-negative")
	}
	if better == nil {
		return nil, errors.New("better function cannot be nil")
	}
	return &PriorityQueue[T]{
		data:     make([]T, 0, capacity),
		capacity: capacity,
		better:   better,
	}, nil
}

// Len 返回队列当前长度
// 时间复杂度: O(1)
func (pq *PriorityQueue[T]) Len() int {
	return len(pq.data)
}

// Enqueue 添加元素到队列中，使用二分查找找到插入位置
// 如果超出capacity，则丢弃优先级最低的元素(队尾)
// 时间复杂度: O(logN) 用于二分查找，O(N) 用于元素移动（最坏情况）
// 整体平均时间复杂度: O(N)
func (pq *PriorityQueue[T]) Enqueue(item T) error {

	if pq.capacity >0 && len(pq.data) == pq.capacity {
		if !pq.better(item,pq.data[pq.capacity-1]) {
			return nil
		}
	}

	// 使用二分查找找到插入位置
	insertPos := pq.binarySearch(item)

	// 插入元素
	pq.data = append(pq.data, item)
	if insertPos < len(pq.data)-1 {
		// 移动元素以保持有序
		copy(pq.data[insertPos+1:], pq.data[insertPos:len(pq.data)-1])
		pq.data[insertPos] = item
	}

	// 检查队列是否已满（只有当capacity>0时才检查）
	if pq.capacity > 0 && len(pq.data) > pq.capacity {
		// 清零将被丢弃的尾部元素，避免内存保持
		var zero T
		for i := pq.capacity; i < len(pq.data); i++ {
			pq.data[i] = zero
		}
		pq.data = pq.data[:pq.capacity]
	}

	return nil
}

// Dequeue 移除并返回队列头部元素（优先级最高的元素）
// 时间复杂度: O(1)
func (pq *PriorityQueue[T]) Dequeue() (T, error) {
	var zero T
	if len(pq.data) == 0 {
		return zero, errors.New("queue is empty")
	}

	// 获取头部元素
	item := pq.data[0]

	// 移除头部元素（O(1)操作）
	pq.data[0] = zero
	pq.data = pq.data[1:]

	return item, nil
}

// binarySearch 使用二分查找找到插入位置
// 时间复杂度: O(logN)
func (pq *PriorityQueue[T]) binarySearch(item T) int {
	low, high := 0, len(pq.data)-1
	for low <= high {
		mid := low + (high-low)/2
		if pq.better(item, pq.data[mid]) {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return low
}
