# Rate Limit API

## Verification

**Protected Route**
hey -n 1000 -c 50 http://localhost:8080/resource

Summary:
  Total:        0.0247 secs  
  Slowest:      0.0063 secs  
  Fastest:      0.0001 secs  
  Average:      0.0012 secs  
  Requests/sec: 40558.2173  
  
  Total data:   53935 bytes
  Size/request: 53 bytes

Response time histogram:
  0.000 [1]     |  
  0.001 [453]   |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■  
  0.001 [288]   |■■■■■■■■■■■■■■■■■■■■■■■■■  
  0.002 [102]   |■■■■■■■■■  
  0.003 [28]    |■■  
  0.003 [32]    |■■■  
  0.004 [38]    |■■■  
  0.004 [28]    |■■  
  0.005 [14]    |■  
  0.006 [6]     |■  
  0.006 [10]    |■  


Latency distribution:
  10% in 0.0002 secs  
  25% in 0.0004 secs  
  50% in 0.0007 secs  
  75% in 0.0014 secs  
  90% in 0.0032 secs  
  95% in 0.0040 secs  
  99% in 0.0057 secs

Details (average, fastest, slowest):  
  DNS+dialup:   0.0001 secs, 0.0001 secs, 0.0063 secs  
  DNS-lookup:   0.0001 secs, 0.0000 secs, 0.0020 secs  
  req write:    0.0000 secs, 0.0000 secs, 0.0012 secs  
  resp wait:    0.0009 secs, 0.0001 secs, 0.0038 secs  
  resp read:    0.0001 secs, 0.0000 secs, 0.0012 secs  

Status code distribution:  
  [200] 5 responses  
  [429] 995 responses  


**Unprotected Route**  
hey -n 1000 -c 50 http://localhost:8080/healthz  

Summary:  
  Total:        0.0207 secs  
  Slowest:      0.0050 secs  
  Fastest:      0.0001 secs  
  Average:      0.0009 secs  
  Requests/sec: 48193.3517  
  
  Total data:   15000 bytes  
  Size/request: 15 bytes

Response time histogram:
  0.000 [1]     |  
  0.001 [422]   |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■  
  0.001 [305]   |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■  
  0.002 [134]   |■■■■■■■■■■■■■  
  0.002 [35]    |■■■  
  0.003 [34]    |■■■  
  0.003 [17]    |■■  
  0.004 [29]    |■■■  
  0.004 [9]     |■  
  0.005 [9]     |■  
  0.005 [5]     |  


Latency distribution:  
  10% in 0.0002 secs  
  25% in 0.0004 secs  
  50% in 0.0006 secs  
  75% in 0.0011 secs  
  90% in 0.0021 secs  
  95% in 0.0031 secs  
  99% in 0.0044 secs  

Details (average, fastest, slowest):  
  DNS+dialup:   0.0001 secs, 0.0001 secs, 0.0050 secs  
  DNS-lookup:   0.0001 secs, 0.0000 secs, 0.0021 secs  
  req write:    0.0000 secs, 0.0000 secs, 0.0008 secs  
  resp wait:    0.0007 secs, 0.0000 secs, 0.0040 secs  
  resp read:    0.0001 secs, 0.0000 secs, 0.0011 secs  

Status code distribution:  
  [200] 1000 responses



## Redis Sharding Quickstart

- Default backend is in-memory. Switch to Redis with `RATE_LIMIT_BACKEND=redis`.
- Describe shards via `REDIS_SHARDS` (comma-separated host:port pairs). If omitted, the service falls back to a single shard built from `REDIS_ADDR`, `REDIS_USERNAME`, `REDIS_PASSWORD`, `REDIS_DB`.
- Optional: `REDIS_HASH_REPLICAS` tunes virtual nodes for the consistent-hash ring (defaults to 128).
- Example with two shards:
  ```
  RATE_LIMIT_BACKEND=redis \
  REDIS_SHARDS=127.0.0.1:6379,127.0.0.1:6380 \
  REDIS_HASH_REPLICAS=256 \
  go run ./cmd/api
  ```
- Every request key is hashed to exactly one shard; keep shard lists identical across app replicas to ensure consistent routing when you scale out.

## Architecture Overview
- **Gin API service**: Runs in multiple pods/VMs, loads config at boot, and exposes `/resource` behind a middleware pipeline.
- **Middleware path**: For each request we derive a key (`client-id:route`), invoke the limiter, and emit `429` or forward to the handler.
- **Limiter abstraction**: In-memory and Redis implementations satisfy the same `Limiter` interface, so deployments can swap backends via config.
- **Redis sharding**: When `RATE_LIMIT_BACKEND=redis`, the service builds a consistent-hash ring (CRC32 + virtual nodes). Every request key is hashed on the application side to pick a shard, ensuring that all replicas of the Go service route that client to the same Redis node without coordination services.
- **Redis nodes**: Each shard stores token-bucket state using a Lua script for atomic refill/consume; TTL keeps idle keys lightweight. You can start with one node and scale horizontally by adding shard addresses.
- **Scaling strategy**: Add more API replicas behind a load balancer. Because limiter state sits in Redis shards keyed via consistent hashing, any replica can process a request while still enforcing global per-client limits. Future additions like ZooKeeper/etcd would only be needed if shard membership must change dynamically without restart.
