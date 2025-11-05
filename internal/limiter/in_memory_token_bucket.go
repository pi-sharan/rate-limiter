package limiter

import (
	"context"
	"math"
	"sync"
	"time"
)

type bucketState struct {
	tokens     float64
	lastRefill time.Time
}

// TokenBucketConfig groups token-bucket parameters.
type TokenBucketConfig struct {
	BucketCapacity int
	RefillRate     int
	RefillInterval time.Duration
}

type inMemoryLimiter struct {
	cfg     TokenBucketConfig
	mu      sync.Mutex
	buckets map[string]*bucketState
}

// NewInMemory returns an in-memory token bucket limiter.
func NewInMemory(cfg TokenBucketConfig) Limiter {
	if cfg.BucketCapacity <= 0 {
		cfg.BucketCapacity = 100
	}
	if cfg.RefillRate <= 0 {
		cfg.RefillRate = 50
	}
	if cfg.RefillInterval <= 0 {
		cfg.RefillInterval = time.Minute
	}

	return &inMemoryLimiter{
		cfg:     cfg,
		buckets: make(map[string]*bucketState),
	}
}

func (l *inMemoryLimiter) Allow(ctx context.Context, key string) (Decision, error) {
	_ = ctx // nothing to do with context yet

	l.mu.Lock()
	defer l.mu.Unlock()

	state, ok := l.buckets[key]
	if !ok {
		state = &bucketState{
			tokens:     float64(l.cfg.BucketCapacity - 1),
			lastRefill: time.Now(),
		}
		l.buckets[key] = state
		return Decision{Allowed: true, Remaining: int(state.tokens)}, nil
	}

	now := time.Now()
	elapsed := now.Sub(state.lastRefill)
	state.lastRefill = now

	refillFraction := elapsed.Seconds() / l.cfg.RefillInterval.Seconds()
	refillTokens := float64(l.cfg.RefillRate) * refillFraction
	if refillTokens > 0 {
		state.tokens = math.Min(float64(l.cfg.BucketCapacity), state.tokens+refillTokens)
	}

	if state.tokens >= 1 {
		state.tokens--
		return Decision{
			Allowed:   true,
			Remaining: int(math.Floor(state.tokens)),
		}, nil
	}

	retryAfterSeconds := (1 - state.tokens) / (float64(l.cfg.RefillRate) / l.cfg.RefillInterval.Seconds())

	return Decision{
		Allowed:    false,
		Remaining:  int(math.Floor(state.tokens)),
		RetryAfter: time.Duration(math.Ceil(retryAfterSeconds*1000)) * time.Millisecond,
	}, nil
}
