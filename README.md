# go-patterns

`go-patterns` is a collection of concurrency patterns and data structures implemented in Go, designed to help developers better understand and utilize Go's concurrency features.

## Project Structure

This repository includes the following modules, each implementing a common concurrency pattern or data structure:

- **container/list**: Implements a generic dynamic array, similar to Python's `list` and JavaScript's `Array`.
- **parallel/barrier**: Provides an implementation of a Barrier for synchronizing multiple goroutines.
- **parallel/mutex**: Implements a simple mutex to ensure that only one goroutine accesses a shared resource at a time.
- **parallel/rwlock**: Implements a read-write lock, supporting multiple readers or a single writer for concurrent access.
- **parallel/semaphore**: Implements a semaphore to limit the number of goroutines accessing shared resources simultaneously.

## Usage

Each module includes example code demonstrating how to use the corresponding pattern or data structure. You can run these examples directly to learn their usage.

### Example

Below is an example of using the `parallel/mutex` module:

```go
package main

import (
	"github.com/leoxiang66/go-patterns/parallel/mutex"
	"time"
)

func main() {
	m := mutex.NewMutex()

	for i := 0; i < 5; i++ {
		go func(id int) {
			m.Lock()
			defer m.Unlock()
			println("Goroutine", id, "is running")
			time.Sleep(1 * time.Second)
		}(i)
	}

	time.Sleep(6 * time.Second)
}
```

## Contribution

Contributions are welcome! Feel free to suggest improvements or submit pull requests. If you have implementations of new concurrency patterns or data structures, we'd love to see them.

## License

This project is open-sourced under the MIT License. For more details, please refer to the [LICENSE](./LICENSE) file.


