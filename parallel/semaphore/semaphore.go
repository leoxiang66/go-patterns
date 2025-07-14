package semaphore

import "time"

/*
Semaphore 限制同一時間可以訪問資源嘅數量
*/

type Semaphore struct {
	container chan struct{}
}

func NewSemaphore(capacity int) *Semaphore {
	return &Semaphore{
		container: make(chan struct{}, capacity),
	}
}

func (sem *Semaphore) Accquire() {
	sem.container <- struct{}{}
}

func (sem *Semaphore) Release() {
	<-sem.container
}

func Example() {
	sem := NewSemaphore(5)

	for {
		sem.Accquire()
		go func() {
			println(time.Now().Second())
			time.Sleep(3 * time.Second)
			sem.Release()
		}()
	}
}
