package list

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
)

// List 是一个通用的动态数组容器，模仿 Python 的 list 和 JavaScript 的 Array，支持常用的增删查改等操作。
type List[T any] struct {
	data []T
}

// ---------- 基础 ----------

// New 创建一个空的 List。
func New[T any]() *List[T] { return &List[T]{data: make([]T, 0)} }

// From 根据给定切片创建一个新的 List。
func From[T any](xs []T) *List[T] {
	cp := make([]T, len(xs))
	copy(cp, xs)
	return &List[T]{data: cp}
}

// Len 返回 List 的长度。
func (l *List[T]) Len() int { return len(l.data) }

// Cap 返回 List 的容量。
func (l *List[T]) Cap() int { return cap(l.data) }

// ToSlice 返回 List 的副本切片。
func (l *List[T]) ToSlice() []T { cp := make([]T, len(l.data)); copy(cp, l.data); return cp }

// String 返回 List 的字符串表示。
func (l *List[T]) String() string { return fmt.Sprintf("%v", l.data) }

// Clone 返回 List 的副本。
func (l *List[T]) Clone() *List[T] { return From(l.data) } // Python: copy(), JS: slice()

// ---------- 索引辅助（支持负索引/边界钳制） ----------

// 规范化单索引：支持负数；越界则返回 -1（表示无效）
func normIndex(length int, i int) int {
	if i < 0 {
		i = length + i
	}
	if i < 0 || i >= length {
		return -1
	}
	return i
}

// 规范化区间：支持负数与开闭区间钳制（行为接近 JS 的 slice / splice，end 为开区间）
func normRange(length int, start, end *int) (s, e int) {
	s = 0
	e = length
	if start != nil {
		s = *start
		if s < 0 {
			s = length + s
		}
		if s < 0 {
			s = 0
		}
		if s > length {
			s = length
		}
	}
	if end != nil {
		e = *end
		if e < 0 {
			e = length + e
		}
		if e < 0 {
			e = 0
		}
		if e > length {
			e = length
		}
	}
	if s > e {
		s, e = e, s // 统一成空区间
	}
	return
}

// ---------- 读写/访问 ----------

// Get 获取指定索引的元素，支持负索引，越界会 panic。
func (l *List[T]) Get(i int) T {
	ii := normIndex(l.Len(), i)
	if ii < 0 {
		panic("index out of range")
	}
	return l.data[ii]
}

// Set 设置指定索引的元素，支持负索引，越界会 panic。
func (l *List[T]) Set(i int, v T) {
	ii := normIndex(l.Len(), i)
	if ii < 0 {
		panic("index out of range")
	}
	l.data[ii] = v
}

// At 获取指定索引的元素，支持负索引，越界返回 false。
func (l *List[T]) At(i int) (T, bool) {
	ii := normIndex(l.Len(), i)
	if ii < 0 {
		var zero T
		return zero, false
	}
	return l.data[ii], true
}

// With 返回修改某索引后的新 List 副本，原 List 不变。
func (l *List[T]) With(i int, v T) *List[T] {
	ii := normIndex(l.Len(), i)
	if ii < 0 {
		panic("index out of range")
	}
	cp := l.ToSlice()
	cp[ii] = v
	return From(cp)
}

// ---------- 增删改（Python: append/extend/insert/remove/pop/clear；JS: push/pop/unshift/shift/splice/copyWithin/fill） ----------

// Append 在 List 末尾添加元素。
func (l *List[T]) Append(items ...T) {
	l.data = append(l.data, items...)
}

// Push 是 Append 的别名。
func (l *List[T]) Push(items ...T) { l.Append(items...) }

// Extend 扩展 List，添加切片中的所有元素。
func (l *List[T]) Extend(xs []T) {
	l.data = append(l.data, xs...)
}

// Unshift 在 List 前端插入元素。
func (l *List[T]) Unshift(items ...T) {
	if len(items) == 0 {
		return
	}
	n := len(l.data)
	l.data = append(items, l.data...) // 直接前插
	if len(l.data) != n+len(items) {
		panic("unshift failed")
	}
}

// Shift 移除并返回第一个元素。
func (l *List[T]) Shift() (T, bool) {
	if l.Len() == 0 {
		var zero T
		return zero, false
	}
	v := l.data[0]
	copy(l.data[0:], l.data[1:])
	l.data = l.data[:len(l.data)-1]
	return v, true
}

// Pop 移除并返回最后一个元素。
func (l *List[T]) Pop() (T, bool) {
	if l.Len() == 0 {
		var zero T
		return zero, false
	}
	last := l.data[len(l.data)-1]
	l.data = l.data[:len(l.data)-1]
	return last, true
}

// Insert 在指定位置插入元素，支持负索引。
func (l *List[T]) Insert(i int, v T) {
	// Insert 前位置，支持负索引（-1 表示末尾前，即与 Python 行为一致）
	if i < 0 {
		i = l.Len() + i
		if i < 0 {
			i = 0
		}
	}
	if i > l.Len() {
		i = l.Len()
	}
	l.data = append(l.data, v)                  // 扩容 1
	copy(l.data[i+1:], l.data[i:len(l.data)-1]) // 右移
	l.data[i] = v
}

// RemoveFirst 移除第一个等于 value 的元素，eq 为比较器。
func (l *List[T]) RemoveFirst(value T, eq func(a, b T) bool) bool {
	for i, x := range l.data {
		if eq(x, value) {
			copy(l.data[i:], l.data[i+1:])
			l.data = l.data[:len(l.data)-1]
			return true
		}
	}
	return false
}

// RemoveAt 移除指定索引的元素，返回被移除的值和是否成功。
func (l *List[T]) RemoveAt(i int) (T, bool) {
	ii := normIndex(l.Len(), i)
	if ii < 0 {
		var zero T
		return zero, false
	}
	v := l.data[ii]
	copy(l.data[ii:], l.data[ii+1:])
	l.data = l.data[:len(l.data)-1]
	return v, true
}

// Clear 清空 List。
func (l *List[T]) Clear() {
	l.data = l.data[:0]
}

// Splice 删除并插入元素，返回被删除的元素列表，并在当前 List 上做就地修改。
func (l *List[T]) Splice(start, deleteCount int, items ...T) *List[T] {
	n := l.Len()
	// 处理 start（允许超界，按 JS 规则钳制到 [0, n]）
	if start < 0 {
		start = n + start
	}
	if start < 0 {
		start = 0
	}
	if start > n {
		start = n
	}
	// 处理 deleteCount（小于 0 视为 0；超过可删长度则钳制）
	if deleteCount < 0 {
		deleteCount = 0
	}
	if deleteCount > n-start {
		deleteCount = n - start
	}

	removed := make([]T, deleteCount)
	copy(removed, l.data[start:start+deleteCount])

	// 构造新切片：前段 + items + 后段
	newData := make([]T, 0, n-deleteCount+len(items))
	newData = append(newData, l.data[:start]...)
	newData = append(newData, items...)
	newData = append(newData, l.data[start+deleteCount:]...)
	l.data = newData
	return From(removed)
}

// CopyWithin 将指定区间的元素复制到目标位置。
func (l *List[T]) CopyWithin(target int, start int, endOpt ...int) {
	n := l.Len()
	end := n
	if len(endOpt) > 0 {
		end = endOpt[0]
	}
	s, e := normRange(n, &start, &end)
	if target < 0 {
		target = n + target
	}
	if target < 0 {
		target = 0
	}
	if target > n {
		target = n
	}
	if s >= e {
		return
	}
	// 复制时要考虑重叠
	cnt := e - s
	if target+cnt > n {
		cnt = n - target
	}
	if cnt <= 0 {
		return
	}
	// 根据相对位置决定复制方向
	if target < s {
		for i := 0; i < cnt; i++ {
			l.data[target+i] = l.data[s+i]
		}
	} else {
		for i := cnt - 1; i >= 0; i-- {
			l.data[target+i] = l.data[s+i]
		}
	}
}

// Fill 用指定值填充区间元素。
func (l *List[T]) Fill(value T, startEnd ...int) {
	var start, end *int
	if len(startEnd) >= 1 {
		start = &startEnd[0]
	}
	if len(startEnd) >= 2 {
		end = &startEnd[1]
	}
	s, e := normRange(l.Len(), start, end)
	for i := s; i < e; i++ {
		l.data[i] = value
	}
}

// ---------- 查询/搜索（includes/index/count/find/findIndex 等） ----------

// Includes 判断 List 是否包含指定元素，eq 为比较器。
func (l *List[T]) Includes(value T, eq func(a, b T) bool) bool {
	for _, x := range l.data {
		if eq(x, value) {
			return true
		}
	}
	return false
}

// IndexOf 返回第一个等于 value 的索引，未找到返回 -1。
func (l *List[T]) IndexOf(value T, eq func(a, b T) bool) int {
	for i, x := range l.data {
		if eq(x, value) {
			return i
		}
	}
	return -1
}

// LastIndexOf 返回最后一个等于 value 的索引，未找到返回 -1。
func (l *List[T]) LastIndexOf(value T, eq func(a, b T) bool) int {
	for i := len(l.data) - 1; i >= 0; i-- {
		if eq(l.data[i], value) {
			return i
		}
	}
	return -1
}

// Count 返回等于 value 的元素个数。
func (l *List[T]) Count(value T, eq func(a, b T) bool) int {
	c := 0
	for _, x := range l.data {
		if eq(x, value) {
			c++
		}
	}
	return c
}

// Find 返回第一个满足条件的元素。
func (l *List[T]) Find(pred func(v T, i int) bool) (T, bool) {
	for i, x := range l.data {
		if pred(x, i) {
			return x, true
		}
	}
	var zero T
	return zero, false
}

// FindIndex 返回第一个满足条件的元素索引。
func (l *List[T]) FindIndex(pred func(v T, i int) bool) int {
	for i, x := range l.data {
		if pred(x, i) {
			return i
		}
	}
	return -1
}

// FindLast 返回最后一个满足条件的元素。
func (l *List[T]) FindLast(pred func(v T, i int) bool) (T, bool) {
	for i := len(l.data) - 1; i >= 0; i-- {
		if pred(l.data[i], i) {
			return l.data[i], true
		}
	}
	var zero T
	return zero, false
}

// FindLastIndex 返回最后一个满足条件的元素索引。
func (l *List[T]) FindLastIndex(pred func(v T, i int) bool) int {
	for i := len(l.data) - 1; i >= 0; i-- {
		if pred(l.data[i], i) {
			return i
		}
	}
	return -1
}

// ---------- 遍历/变换（forEach/map/filter/reduce/some/every 等） ----------

// ForEach 只读遍历 List。
func (l *List[T]) ForEach(fn func(v T, i int)) {
	for i := 0; i < len(l.data); i++ {
		fn(l.data[i], i)
	}
}

// ForEachAsync 并发只读遍历 List，支持最大 goroutine 数。
func (l *List[T]) ForEachAsync(
	ctx context.Context,
	maxGoroutines int,
	fn func(v T, i int),
) error {
	if maxGoroutines <= 0 {
		maxGoroutines = runtime.GOMAXPROCS(0)
	}
	sem := make(chan struct{}, maxGoroutines)
	var wg sync.WaitGroup

	n := len(l.data)
	for i := 0; i < n; i++ {
		// 若已取消，则停止派发新的任务，但需要等待已在跑的任务收尾
		select {
		case <-ctx.Done():
			wg.Wait()
			return ctx.Err()
		case sem <- struct{}{}:
		}

		v := l.data[i] // 拷贝，避免 goroutine 里读到被外部改动的值
		wg.Add(1)
		go func(v T, i int) {
			defer func() {
				<-sem
				wg.Done()
			}()
			fn(v, i)
		}(v, i)
	}

	wg.Wait()
	return nil
}

// Map 返回一个新的 List，元素为 fn 映射结果。
func Map[T any, R any](l *List[T], fn func(v T, i int) R) *List[R] {
	out := make([]R, l.Len())
	for i := 0; i < len(l.data); i++ {
		out[i] = fn(l.data[i], i)
	}
	return From(out)
}

// T -> any, return new
// Map 返回一个新的 List，元素为 fn 映射结果（类型为 any）。
func (l *List[T]) Map(fn func(v T, i int) any) *List[any] {
	out := make([]any, l.Len())
	for i := 0; i < len(l.data); i++ {
		out[i] = fn(l.data[i], i)
	}
	return From(out)
}

// T -> T, in place
// MapInPlace 原地映射 List 元素。
func (l *List[T]) MapInPlace(fn func(v T, i int) T) {
	for i := 0; i < len(l.data); i++ {
		l.data[i] = fn(l.data[i], i)
	}
}

// MapAsync 并发映射 List，返回新 List。
func MapAsync[T any, R any](
	ctx context.Context,
	l *List[T],
	maxGoroutines int,
	fn func(v T, i int) R,
) (*List[R], error) {
	if maxGoroutines <= 0 {
		maxGoroutines = runtime.GOMAXPROCS(0)
	}
	sem := make(chan struct{}, maxGoroutines)
	var wg sync.WaitGroup

	out := make([]R, len(l.data))

	for i := 0; i < len(l.data); i++ {
		select {
		case <-ctx.Done():
			wg.Wait()
			return nil, ctx.Err()
		case sem <- struct{}{}:
		}

		v := l.data[i]
		idx := i
		wg.Add(1)
		go func(v T, i int) {
			defer func() {
				<-sem
				wg.Done()
			}()
			out[i] = fn(v, i)
		}(v, idx)
	}

	wg.Wait()
	return From(out), nil
}

// Filter 返回所有满足条件的元素组成的新 List。
func (l *List[T]) Filter(pred func(v T, i int) bool) *List[T] {
	out := make([]T, 0, l.Len())
	for i, x := range l.data {
		if pred(x, i) {
			out = append(out, x)
		}
	}
	return From(out)
}

// Reduce 按照 fn 规则聚合 List 元素。
func Reduce[T any, R any](l *List[T], init R, fn func(acc R, v T, i int) R) R {
	acc := init
	for i, x := range l.data {
		acc = fn(acc, x, i)
	}
	return acc
}

// Some 判断是否存在满足条件的元素。
func (l *List[T]) Some(pred func(v T, i int) bool) bool {
	for i, x := range l.data {
		if pred(x, i) {
			return true
		}
	}
	return false
}

// Every 判断所有元素是否都满足条件。
func (l *List[T]) Every(pred func(v T, i int) bool) bool {
	for i, x := range l.data {
		if !pred(x, i) {
			return false
		}
	}
	return true
}

// ---------- 切片/排序/反转（slice/sort/reverse 以及不可变版本） ----------

// Slice 返回指定区间的新 List，不修改原 List。
func (l *List[T]) Slice(startEnd ...int) *List[T] {
	var start, end *int
	if len(startEnd) >= 1 {
		start = &startEnd[0]
	}
	if len(startEnd) >= 2 {
		end = &startEnd[1]
	}
	s, e := normRange(l.Len(), start, end)
	return From(l.data[s:e])
}

// Reverse 原地反转 List。
func (l *List[T]) Reverse() {
	for i, j := 0, len(l.data)-1; i < j; i, j = i+1, j-1 {
		l.data[i], l.data[j] = l.data[j], l.data[i]
	}
}

// ToReversed 返回反转后的新 List。
func (l *List[T]) ToReversed() *List[T] {
	cp := l.ToSlice()
	for i, j := 0, len(cp)-1; i < j; i, j = i+1, j-1 {
		cp[i], cp[j] = cp[j], cp[i]
	}
	return From(cp)
}

// Sort 使用插入排序原地排序 List。
func (l *List[T]) Sort(less func(a, b T) bool) {
	// 简单实现插入排序（小规模够用；需要高性能可改成 sort.SliceStable）
	n := len(l.data)
	for i := 1; i < n; i++ {
		j := i
		for j > 0 && less(l.data[j], l.data[j-1]) {
			l.data[j], l.data[j-1] = l.data[j-1], l.data[j]
			j--
		}
	}
}

// ToSorted 返回排序后的新 List。
func (l *List[T]) ToSorted(less func(a, b T) bool) *List[T] {
	cp := l.ToSlice()
	tmp := &List[T]{data: cp}
	tmp.Sort(less)
	return tmp
}

// ToSpliced 返回执行 splice 后的新 List，原 List 不变。
func (l *List[T]) ToSpliced(start, deleteCount int, items ...T) *List[T] {
	cp := l.Clone()
	cp.Splice(start, deleteCount, items...)
	return cp
}

// Join 用分隔符连接 List 元素，toStr 为元素转字符串函数。
func (l *List[T]) Join(sep string, toStr func(v T) string) string {
	if l.Len() == 0 {
		return ""
	}
	sb := strings.Builder{}
	for i, x := range l.data {
		if i > 0 {
			sb.WriteString(sep)
		}
		sb.WriteString(toStr(x))
	}
	return sb.String()
}
