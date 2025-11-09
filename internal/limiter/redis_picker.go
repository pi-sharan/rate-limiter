package limiter

import goRedis "github.com/redis/go-redis/v9"

// RedisClientPicker selects the redis client for a given key (supporting sharding).
type RedisClientPicker interface {
	Pick(key string) (goRedis.Cmdable, error)
}
