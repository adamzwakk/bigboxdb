package db

import (
    "context"
	"os"
    "encoding/json"
    "time"

    "github.com/redis/go-redis/v9"
)

var (
    Ctx = context.Background()
    Rdb *redis.Client
)

func InitRedis() {
    Rdb = redis.NewClient(&redis.Options{
        Addr: os.Getenv("REDIS_ADDR")+":6379",
    })
}

func GetOrSetCache[T any](key string, ttl time.Duration, fetcher func() (T, error)) (T, error) {
    var result T

    // Try cache first
    cached, err := Rdb.Get(Ctx, key).Result()
    if err == nil {
        if err := json.Unmarshal([]byte(cached), &result); err == nil {
            return result, nil
        }
    }

    // Cache miss - fetch from DB
    result, err = fetcher()
    if err != nil {
        return result, err
    }

    // Store in cache
    if data, err := json.Marshal(result); err == nil {
        Rdb.Set(Ctx, key, data, ttl)
    }

    return result, nil
}

func Invalidate(keys ...string) {
    Rdb.Del(Ctx, keys...)
}