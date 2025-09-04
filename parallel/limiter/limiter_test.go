package limiter

import (
	"testing"
	"time"
)

func TestStaticLimiter(t *testing.T) {
	limiter := NewStaticLimiter(10 * time.Millisecond)
	start := time.Now()
	limiter.GrantNextToken()
	limiter.GrantNextToken()
	elapsed := time.Since(start)
	if elapsed < 20*time.Millisecond {
		t.Errorf("limiter did not wait enough, elapsed=%v", elapsed)
	}
	limiter.Stop()
}
