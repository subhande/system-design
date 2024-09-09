package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var QUORUM_COUNT int = 3
var SERVER_COUNT int = 5

// RedisLock represents a Redis remote lock
type RedisLock struct {
	key    string
	client *redis.Client
	id     string
}

type RedisLockQuorum struct {
	key     string
	clients []*redis.Client
	id      string
}

func NewRedisLock(client *redis.Client, key string, id string) *RedisLock {
	return &RedisLock{
		key:    key,
		client: client,
		id:     id,
	}
}

func NewRedisLockQuorum(clients []*redis.Client, key string, id string) *RedisLockQuorum {
	return &RedisLockQuorum{
		key:     key,
		clients: clients,
		id:      id,
	}
}

func (r *RedisLock) Acquire() bool {
	// Try to acquire the lock
	result := r.client.SetNX(context.Background(), r.key, r.id, 10*time.Second).Val()
	return result
}

func (r *RedisLock) Release() bool {
	// Release the lock
	result := r.client.Eval(context.Background(), "if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) else return 0 end", []string{r.key}, r.id).Val()
	resultStr := int(result.(int64))
	return resultStr == 1
}

func (r *RedisLockQuorum) Acquire() bool {
	// Try to acquire the lock
	count := 0
	for _, client := range r.clients {
		result := client.SetNX(context.Background(), r.key, r.id, 10*time.Second).Val()
		if result {
			count++
		}
	}
	return count >= QUORUM_COUNT
}

func (r *RedisLockQuorum) Release() bool {
	// Release the lock
	count := 0
	for _, client := range r.clients {
		result := client.Eval(context.Background(), "if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) else return 0 end", []string{r.key}, r.id).Val()
		resultStr := int(result.(int64))
		if resultStr == 1 {
			count++
		}
	}
	return count >= QUORUM_COUNT
}

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	lock := NewRedisLock(client, "my_lock", id)
	acquired := lock.Acquire()
	if acquired {
		fmt.Println("Lock acquired")
	} else {
		fmt.Println("Failed to acquire lock")
	}

	id2 := fmt.Sprintf("%d", time.Now().UnixNano())
	lock2 := NewRedisLock(client, "my_lock", id2)
	release2 := lock2.Release()

	if release2 {
		fmt.Println("Lock released")
	} else {
		fmt.Println("Failed to release lock")
	}

	release := lock.Release()
	if release {
		fmt.Println("Lock released")
	} else {
		fmt.Println("Failed to release lock")
	}

}
