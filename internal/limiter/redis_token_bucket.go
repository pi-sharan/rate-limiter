package limiter

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	goRedis "github.com/redis/go-redis/v9"
)

// Redis-backed token bucket limiter.
type redisLimiter struct {
	cfg       TokenBucketConfig
	picker    RedisClientPicker
	keyPrefix string
}

// NewRedisWithPicker creates a limiter that uses the provided client picker.
func NewRedisWithPicker(picker RedisClientPicker, cfg TokenBucketConfig) (Limiter, error) {
	if picker == nil {
		return nil, fmt.Errorf("redis client picker is nil")
	}

	if cfg.BucketCapacity <= 0 {
		cfg.BucketCapacity = 100
	}
	if cfg.RefillRate <= 0 {
		cfg.RefillRate = 50
	}
	if cfg.RefillInterval <= 0 {
		cfg.RefillInterval = time.Minute
	}

	return &redisLimiter{
		cfg:       cfg,
		picker:    picker,
		keyPrefix: "rate_limiter",
	}, nil
}

func (r *redisLimiter) Allow(ctx context.Context, key string) (Decision, error) {
	if key == "" {
		key = "default"
	}

	client, err := r.picker.Pick(key)
	if err != nil {
		return Decision{}, err
	}

	now := time.Now()
	intervalMs := int64(r.cfg.RefillInterval / time.Millisecond)

	if intervalMs <= 0 {
		intervalMs = 1000
	}

	res, err := redisTokenBucketScript.Run(ctx, client, []string{r.bucketKey(key)},
		r.cfg.BucketCapacity,
		r.cfg.RefillRate,
		intervalMs,
		now.UnixMilli(),
	).Result()

	if err != nil {
		return Decision{}, err
	}

	values, ok := res.([]interface{})
	if !ok || len(values) != 3 {
		return Decision{}, fmt.Errorf("unexpected script response: %v", res)
	}

	allowedInt, err := toInt64(values[0])
	if err != nil {
		return Decision{}, fmt.Errorf("parse allowed: %w", err)
	}

	tokens, err := toFloat64(values[1])
	if err != nil {
		return Decision{}, fmt.Errorf("parse tokens: %w", err)
	}

	retryAfterMs, err := toInt64(values[2])
	if err != nil {
		return Decision{}, fmt.Errorf("parse retryAfter: %w", err)
	}

	decision := Decision{
		Allowed:    allowedInt == 1,
		Remaining:  int(math.Floor(tokens)),
		RetryAfter: time.Duration(retryAfterMs) * time.Millisecond,
	}

	return decision, nil
}

func (r *redisLimiter) bucketKey(key string) string {
	return fmt.Sprintf("%s:%s", r.keyPrefix, key)
}

func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case int64:
		return float64(val), nil
	case float64:
		return val, nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("unsupported type %T", v)
	}
}

func toInt64(v interface{}) (int64, error) {
	switch val := v.(type) {
	case int64:
		return val, nil
	case float64:
		return int64(val), nil
	case string:
		parsed, err := strconv.ParseInt(val, 10, 64)
		return parsed, err
	default:
		return 0, fmt.Errorf("unsupported type %T", v)
	}
}

type staticPicker struct {
	client goRedis.Cmdable
}

func (s *staticPicker) Pick(key string) (goRedis.Cmdable, error) {
	if s.client == nil {
		return nil, fmt.Errorf("redis client is nil")
	}
	return s.client, nil
}

var redisTokenBucketScript = goRedis.NewScript(`
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local interval_ms = tonumber(ARGV[3])
local now_ms = tonumber(ARGV[4])

local data = redis.call("HMGET", key, "tokens", "last_refill")
local tokens = tonumber(data[1])
local last_refill = tonumber(data[2])

if tokens == nil or last_refill == nil then
  tokens = capacity
  last_refill = now_ms
else
  local elapsed = now_ms - last_refill
  if elapsed > 0 and interval_ms > 0 then
    local refill = (elapsed / interval_ms) * refill_rate
    if refill > 0 then
      tokens = math.min(capacity, tokens + refill)
      last_refill = now_ms
    end
  end
end

local allowed = 0
local retry_after = 0

if tokens >= 1 then
  tokens = tokens - 1
  allowed = 1
else
  if refill_rate > 0 then
    local missing = 1 - tokens
    retry_after = math.ceil((missing / refill_rate) * interval_ms)
  else
    retry_after = interval_ms
  end
end

redis.call("HMSET", key, "tokens", tokens, "last_refill", last_refill)
redis.call("PEXPIRE", key, math.max(interval_ms, 1000))

return {allowed, tokens, retry_after}
`)
