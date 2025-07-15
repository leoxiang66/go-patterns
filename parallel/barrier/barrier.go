package barrier

import "time"

type Barrier struct {
	num_workers int
	ready       chan struct{}
}

func NewBarrier(num_workers int) *Barrier {
	return &Barrier{
		num_workers: num_workers,
		ready:       make(chan struct{}, num_workers),
	}
}

func (barrier *Barrier) Done() {
	barrier.ready <- struct{}{}
}

func (barrier *Barrier) Sync() {
	for i := 0; i < barrier.num_workers; i++ {
		<-barrier.ready
	}
}


func Example()  {
	barrier := NewBarrier(5)
	for i := 0; i < 5; i++ {
		go func(i int) {
			println("goroutine ",i, " starts!")
			time.Sleep(time.Duration((i+1)*3)*time.Second)
			println("goroutine ",i, " finishes!")
			barrier.Done()
		}(i)
	}

	barrier.Sync()
	println("All goroutine finish!")
}
