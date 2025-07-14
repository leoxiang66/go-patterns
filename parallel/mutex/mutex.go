package mutex

import "time"

/*
Mutex： 同時間淨係得一個thread可以訪問， 相當於capacity = 1 嘅semaphore
*/

type Mutex struct {
	container chan struct{}
}

func NewMutex() *Mutex {
	return &Mutex{
		container: make(chan struct{}, 1),
	}
}

func (mutex *Mutex) Lock() {
	mutex.container <- struct{}{}
}

func (mutex *Mutex) Unlock() {
	<-mutex.container
}

func Example() {
	mutex := NewMutex()

	for {
		mutex.Lock()
		go func() {
			println(time.Now().Second())
			time.Sleep(3 * time.Second)
			mutex.Unlock()
		}()
	}
}
