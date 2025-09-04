package barrier

import (
	"testing"
	"time"
)

func TestEasyBarrier(t *testing.T) {
	barrier := NewEasyBarrier(2)
	done := make(chan struct{})
	go func() {
		barrier.Done()
		barrier.Done()
		close(done)
	}()
	barrier.Sync()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("barrier.Sync timeout")
	}
}

func TestLightBarrier(t *testing.T) {
	b := NewLightBarrier()
	b.Add()
	b.Add()
	done := make(chan struct{})
	go func() {
		b.Done()
		b.Done()
		close(done)
	}()
	b.Sync()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("LightBarrier.Sync timeout")
	}
}
