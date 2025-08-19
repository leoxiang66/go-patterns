package list

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
)

// List 是一个通用动态数组，模仿 Python list 与 JS Array 常用方法。
type List[T any] struct {
	data []T
}

// ---------- 基础 ----------

func New[T any]() *List[T] { return &List[T]{data: make([]T, 0)} }
func From[T any](xs []T) *List[T] {
	cp := make([]T, len(xs))
	copy(cp, xs)
	return &List[T]{data: cp}
}
func (l *List[T]) Len() int        { return len(l.data) }
func (l *List[T]) Cap() int        { return cap(l.data) }
func (l *List[T]) ToSlice() []T    { cp := make([]T, len(l.data)); copy(cp, l.data); return cp }
func (l *List[T]) String() string  { return fmt.Sprintf("%v", l.data) }
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

// Get / Set（越界会 panic，与 Python/JS 一致）
func (l *List[T]) Get(i int) T {
	ii := normIndex(l.Len(), i)
	if ii < 0 {
		panic("index out of range")
	}
	return l.data[ii]
}
func (l *List[T]) Set(i int, v T) {
	ii := normIndex(l.Len(), i)
	if ii < 0 {
		panic("index out of range")
	}
	l.data[ii] = v
}

// At：JS 的 at()，支持负索引。越界返回第二个值 false
func (l *List[T]) At(i int) (T, bool) {
	ii := normIndex(l.Len(), i)
	if ii < 0 {
		var zero T
		return zero, false
	}
	return l.data[ii], true
}

// With：JS 的 with()，返回修改某索引后的新副本
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

func (l *List[T]) Append(items ...T) { // Python: append / JS: push
	l.data = append(l.data, items...)
}
func (l *List[T]) Push(items ...T) { l.Append(items...) } // alias

func (l *List[T]) Extend(xs []T) { // Python: extend
	l.data = append(l.data, xs...)
}

func (l *List[T]) Unshift(items ...T) { // JS: unshift
	if len(items) == 0 {
		return
	}
	n := len(l.data)
	l.data = append(items, l.data...) // 直接前插
	if len(l.data) != n+len(items) {
		panic("unshift failed")
	}
}

func (l *List[T]) Shift() (T, bool) { // JS: shift
	if l.Len() == 0 {
		var zero T
		return zero, false
	}
	v := l.data[0]
	copy(l.data[0:], l.data[1:])
	l.data = l.data[:len(l.data)-1]
	return v, true
}

func (l *List[T]) Pop() (T, bool) { // Python: pop() / JS: pop()
	if l.Len() == 0 {
		var zero T
		return zero, false
	}
	last := l.data[len(l.data)-1]
	l.data = l.data[:len(l.data)-1]
	return last, true
}

func (l *List[T]) Insert(i int, v T) { // Python: insert(i, x)
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

func (l *List[T]) RemoveFirst(value T, eq func(a, b T) bool) bool { // Python: remove(x)
	for i, x := range l.data {
		if eq(x, value) {
			copy(l.data[i:], l.data[i+1:])
			l.data = l.data[:len(l.data)-1]
			return true
		}
	}
	return false
}

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

func (l *List[T]) Clear() { // Python: clear()
	l.data = l.data[:0]
}

// Splice：JS 的 splice(start, deleteCount, ...items)
// 返回被删除的元素列表，并在当前 List 上做就地修改
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

// CopyWithin：JS 的 copyWithin(target, start, end?)
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

// Fill：JS 的 fill(value, start?, end?)
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

// Includes：JS includes(value)，需要 eq 比较器
func (l *List[T]) Includes(value T, eq func(a, b T) bool) bool {
	for _, x := range l.data {
		if eq(x, value) {
			return true
		}
	}
	return false
}

// IndexOf / LastIndexOf（JS 行为）
func (l *List[T]) IndexOf(value T, eq func(a, b T) bool) int {
	for i, x := range l.data {
		if eq(x, value) {
			return i
		}
	}
	return -1
}
func (l *List[T]) LastIndexOf(value T, eq func(a, b T) bool) int {
	for i := len(l.data) - 1; i >= 0; i-- {
		if eq(l.data[i], value) {
			return i
		}
	}
	return -1
}

// Count：Python count(x)
func (l *List[T]) Count(value T, eq func(a, b T) bool) int {
	c := 0
	for _, x := range l.data {
		if eq(x, value) {
			c++
		}
	}
	return c
}

// Find / FindIndex / FindLast / FindLastIndex（JS 行为）
func (l *List[T]) Find(pred func(v T, i int) bool) (T, bool) {
	for i, x := range l.data {
		if pred(x, i) {
			return x, true
		}
	}
	var zero T
	return zero, false
}
func (l *List[T]) FindIndex(pred func(v T, i int) bool) int {
	for i, x := range l.data {
		if pred(x, i) {
			return i
		}
	}
	return -1
}
func (l *List[T]) FindLast(pred func(v T, i int) bool) (T, bool) {
	for i := len(l.data) - 1; i >= 0; i-- {
		if pred(l.data[i], i) {
			return l.data[i], true
		}
	}
	var zero T
	return zero, false
}
func (l *List[T]) FindLastIndex(pred func(v T, i int) bool) int {
	for i := len(l.data) - 1; i >= 0; i-- {
		if pred(l.data[i], i) {
			return i
		}
	}
	return -1
}

// ---------- 遍历/变换（forEach/map/filter/reduce/some/every 等） ----------

// ForEach: read only iterations
func (l *List[T]) ForEach(fn func(v T, i int)) {
	for i := 0; i < len(l.data); i++ {
		fn(l.data[i], i)
	}
}

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

// T -> R, return new
func Map[T any, R any](l *List[T], fn func(v T, i int) R) *List[R] {
	out := make([]R, l.Len())
	for i := 0; i < len(l.data); i++ {
		out[i] = fn(l.data[i], i)
	}
	return From(out)
}

// T -> any, return new
func (l *List[T]) Map(fn func(v T, i int) any) *List[any] {
	out := make([]any, l.Len())
	for i := 0; i < len(l.data); i++ {
		out[i] = fn(l.data[i], i)
	}
	return From(out)
}

// T -> T, in place
func (l *List[T]) MapInPlace(fn func(v T, i int) T) {
	for i := 0; i < len(l.data); i++ {
		l.data[i] = fn(l.data[i], i)
	}
}

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

func (l *List[T]) Filter(pred func(v T, i int) bool) *List[T] {
	out := make([]T, 0, l.Len())
	for i, x := range l.data {
		if pred(x, i) {
			out = append(out, x)
		}
	}
	return From(out)
}

func Reduce[T any, R any](l *List[T], init R, fn func(acc R, v T, i int) R) R {
	acc := init
	for i, x := range l.data {
		acc = fn(acc, x, i)
	}
	return acc
}

func (l *List[T]) Some(pred func(v T, i int) bool) bool {
	for i, x := range l.data {
		if pred(x, i) {
			return true
		}
	}
	return false
}
func (l *List[T]) Every(pred func(v T, i int) bool) bool {
	for i, x := range l.data {
		if !pred(x, i) {
			return false
		}
	}
	return true
}

// ---------- 切片/排序/反转（slice/sort/reverse 以及不可变版本） ----------

// Slice：JS slice(start?, end?)（不修改原 List）
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

// Reverse（就地）/ ToReversed（返回新副本）
func (l *List[T]) Reverse() {
	for i, j := 0, len(l.data)-1; i < j; i, j = i+1, j-1 {
		l.data[i], l.data[j] = l.data[j], l.data[i]
	}
}
func (l *List[T]) ToReversed() *List[T] {
	cp := l.ToSlice()
	for i, j := 0, len(cp)-1; i < j; i, j = i+1, j-1 {
		cp[i], cp[j] = cp[j], cp[i]
	}
	return From(cp)
}

// Sort（就地）/ ToSorted（返回新副本）
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
func (l *List[T]) ToSorted(less func(a, b T) bool) *List[T] {
	cp := l.ToSlice()
	tmp := &List[T]{data: cp}
	tmp.Sort(less)
	return tmp
}

// ToSpliced：JS 的 toSpliced（返回新副本，不改原 List）
func (l *List[T]) ToSpliced(start, deleteCount int, items ...T) *List[T] {
	cp := l.Clone()
	cp.Splice(start, deleteCount, items...)
	return cp
}

// Join：JS join / Python join（需要 toStr 转换）
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
