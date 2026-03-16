package redislock

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisLocker holds the Redis client used for locking
type RedisLocker struct {
	client *redis.Client
}

// NewRedisLocker creates a new Locker instance
func NewRedisLocker(client *redis.Client) *RedisLocker {
	return &RedisLocker{
		client: client,
	}
}

// AcquireLock attempts to acquire a distributed lock for a specific key
// It uses SET NX EX to ensure atomicity.
// Returns true if acquired, false if it's already locked.
func (l *RedisLocker) AcquireLock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	// "SET key value EX duration NX"
	// Value doesn't matter much here since we just use it to block,
	// but a unique ID is better for strict release. For simplicity, we use "1".
	success, err := l.client.SetNX(ctx, "lock:seat:"+key, "1", expiration).Result()
	if err != nil {
		return false, err
	}
	return success, nil
}

func (l *RedisLocker) AcquireMultipleLocks(ctx context.Context, prefix string, keys []string, expiration time.Duration) (bool, []string, error) {
	var lockedKeys []string

	for _, key := range keys {
		lockKey := prefix + ":" + key
		success, err := l.client.SetNX(ctx, lockKey, "locked", expiration).Result()

		if err != nil || !success {
			// Rollback: if we fail to lock all, unlock the ones we already locked
			for _, lk := range lockedKeys {
				l.ReleaseLock(ctx, lk)
			}
			return false, nil, err
		}
		lockedKeys = append(lockedKeys, lockKey)
	}

	return true, lockedKeys, nil
}

// ReleaseLock removes the lock
func (l *RedisLocker) ReleaseLock(ctx context.Context, lockKey string) error {
	return l.client.Del(ctx, lockKey).Err()
}

// ReleaseMultipleLocks releases an array of locked keys
func (l *RedisLocker) ReleaseMultipleLocks(ctx context.Context, lockKeys []string) {
	for _, lk := range lockKeys {
		_ = l.client.Del(ctx, lk).Err()
	}
}
