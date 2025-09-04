package mutex

import (
	"sync"
	"testing"
	"time"
)

func TestMutex_Race(t *testing.T) {
	m := NewMutex()
	var counter int
	const goroutines = 50
	const increments = 100

	wg := sync.WaitGroup{}
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < increments; j++ {
				m.Lock()
				counter++
				m.Unlock()
			}
		}()
	}

	wg.Wait()

	expected := goroutines * increments
	if counter != expected {
		t.Errorf("counter = %d, want %d", counter, expected)
	}
}

func TestMutex_Parallel(t *testing.T) {
	m := NewMutex()
	var maxConcurrent int
	var current int
	const goroutines = 20

	wg := sync.WaitGroup{}
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			m.Lock()
			current++
			if current > maxConcurrent {
				maxConcurrent = current
			}
			time.Sleep(10 * time.Millisecond)
			current--
			m.Unlock()
		}()
	}

	wg.Wait()

	if maxConcurrent > 1 {
		t.Errorf("maxConcurrent = %d, want <= 1", maxConcurrent)
	}
}
