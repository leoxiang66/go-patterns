package semaphore

import "sync"

type SemaphoreByCond struct {
	numTokens int
	mu        sync.Mutex
	cond      *sync.Cond
}

func NewSemaphoreByCond(capacity int) *SemaphoreByCond {
	ret := &SemaphoreByCond{
		numTokens: capacity,
	}
	ret.cond = sync.NewCond(&ret.mu)
	return ret
}

func (sm *SemaphoreByCond) Acquire() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for sm.numTokens == 0 {
		sm.cond.Wait()
	}
	sm.numTokens--
}

func (sm *SemaphoreByCond) TryAcquire() bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.numTokens == 0 {
		return false
	} else {
		sm.numTokens--
		return true
	}
}

func (sm *SemaphoreByCond) Release() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.numTokens++
	sm.cond.Broadcast()
}
