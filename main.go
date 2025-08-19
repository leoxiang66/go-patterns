package main

import (
	// "github.com/leoxiang66/go-patterns/parallel/barrier"
	"context"
	"fmt"
	"time"

	"github.com/leoxiang66/go-patterns/container/list"
)

// import "github.com/leoxiang66/go-patterns/parallel/semaphore"
// import "github.com/leoxiang66/go-patterns/parallel/mutex"
// import "github.com/leoxiang66/go-patterns/parallel/rwlock"

func main() {
	// semaphore.Example()
	// mutex.Example()
	// rwlock.Example()
	// barrier.Example()

	l1 := list.From([]int{1, 2, 3, 4, 5})

	l1.MapInPlace(func(v, i int) int {
		return v * 2
	})

	fmt.Println(l1.ToSlice())

	l2 := list.New[any]()

	l2.Append([]any{1, "hi"})

	fmt.Println(l2.ToSlice())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := l1.ForEachAsync(ctx, 2, func(v, i int) {
		time.Sleep(3 * time.Second)
		fmt.Println(v)
	})

	fmt.Println(err)

	l3 := l1.Map(func(v, i int) any {
		return v
	})

	fmt.Println(l3.ToSlice()...)
}
