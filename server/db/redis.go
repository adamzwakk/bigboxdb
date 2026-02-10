package db

import (
    "context"
	"os"

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