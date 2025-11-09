package rediscluster

import (
	"errors"
	"fmt"
	"hash/crc32"
	"sort"

	goRedis "github.com/redis/go-redis/v9"
)

// Node represents a redis shard in the hash ring.
type Node struct {
	ID     string
	Client goRedis.Cmdable
}

// ConsistentHash implements a basic consistent hashing ring.
type ConsistentHash struct {
	replicas int
	keys     []uint32
	ring     map[uint32]goRedis.Cmdable
}

// NewConsistentHash creates a consistent hash ring from the provided nodes.
func NewConsistentHash(nodes []Node, replicas int) (*ConsistentHash, error) {
	if len(nodes) == 0 {
		return nil, errors.New("no redis shards provided")
	}
	if replicas <= 0 {
		replicas = 128
	}

	ring := make(map[uint32]goRedis.Cmdable, len(nodes)*replicas)
	keys := make([]uint32, 0, len(nodes)*replicas)

	for _, node := range nodes {
		if node.ID == "" || node.Client == nil {
			return nil, fmt.Errorf("invalid node: %+v", node)
		}
		for i := 0; i < replicas; i++ {
			hash := hashKey(fmt.Sprintf("%s#%d", node.ID, i))
			ring[hash] = node.Client
			keys = append(keys, hash)
		}
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	return &ConsistentHash{
		replicas: replicas,
		keys:     keys,
		ring:     ring,
	}, nil
}

// Pick returns the redis client responsible for the provided key.
func (c *ConsistentHash) Pick(key string) (goRedis.Cmdable, error) {
	if len(c.keys) == 0 {
		return nil, errors.New("hash ring is empty")
	}

	hash := hashKey(key)
	idx := sort.Search(len(c.keys), func(i int) bool { return c.keys[i] >= hash })
	if idx == len(c.keys) {
		idx = 0
	}

	client := c.ring[c.keys[idx]]
	if client == nil {
		return nil, errors.New("no client found for key")
	}

	return client, nil
}

func hashKey(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}
