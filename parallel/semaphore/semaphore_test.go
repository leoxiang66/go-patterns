package semaphore

import (
	"testing"
	"time"
)

func TestSemaphore(t *testing.T) {
	sem := NewSemaphore(2)
	sem.Acquire()
	sem.Acquire()
	done := make(chan struct{})
	go func() {
		sem.Release()
		sem.Release()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Semaphore release timeout")
	}
}
