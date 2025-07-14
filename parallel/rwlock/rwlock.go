package rwlock

import (
	"fmt"
	"sync"
	"time"

	"github.com/leoxiang66/go-patterns/parallel/mutex"
)

type RWLock struct {
	readers  int32
	noreader chan struct{}
	writer   chan struct{}
	mutex    *mutex.Mutex
}

func NewRWLock() *RWLock {
	sem := make(chan struct{}, 1)
	noreader := make(chan struct{}, 1)
	noreader <- struct{}{}
	return &RWLock{
		writer:   sem,
		noreader: noreader,
		mutex:    mutex.NewMutex(),
	}
}

func (rwlock *RWLock) RLock() {
	rwlock.writer <- struct{}{}
	rwlock.mutex.Lock()
	if rwlock.readers == 0 {
		<-rwlock.noreader
	}
	rwlock.readers++
	rwlock.mutex.Unlock()
	<-rwlock.writer
}

func (rwlock *RWLock) RUnlock() {
	rwlock.mutex.Lock()
	rwlock.readers--
	if rwlock.readers == 0 {
		rwlock.noreader <- struct{}{}
	}
	rwlock.mutex.Unlock()
}

func (rw *RWLock) WLock() {
	rw.writer <- struct{}{}
	<-rw.noreader
}

func (rw *RWLock) WUnlock() {
	rw.noreader <- struct{}{}
	<-rw.writer
}

func Example() {
	var wg sync.WaitGroup
	lock := NewRWLock()
	count := 0

	// 启动 3 个读者
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			lock.RLock()
			fmt.Printf("Reader %d reading %d...\n", id, count)
			time.Sleep(1000 * time.Millisecond)
			lock.RUnlock()
		}(i)
	}

	// 启动 1 个写者
	wg.Add(1)
	go func() {
		defer wg.Done()
		lock.WLock()
		fmt.Println("Writer writing...")
		count++
		time.Sleep(1000 * time.Millisecond)
		lock.WUnlock()
	}()

	wg.Wait()
}
