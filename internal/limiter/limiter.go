package limiter

import (
	"context"
	"time"
)

type Decision struct {
	Allowed    bool
	Remaining  int
	RetryAfter time.Duration
}

type Limiter interface {
	Allow(ctx context.Context, key string) (Decision, error)
}
