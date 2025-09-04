package limiter

import (
	"sync"
	"time"
)

type StaticLimiter struct {
	ticker *time.Ticker
	mu     sync.Mutex
}

func NewStaticLimiter(interval time.Duration) *StaticLimiter {
	return &StaticLimiter{
		ticker: time.NewTicker(interval),
	}
}

func (l *StaticLimiter) GrantNextToken() {
	l.mu.Lock()
	defer l.mu.Unlock()
	<-l.ticker.C
}

func (l *StaticLimiter) Reset(interval time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ticker.Reset(interval)
}

func (l *StaticLimiter) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ticker.Stop()
}
