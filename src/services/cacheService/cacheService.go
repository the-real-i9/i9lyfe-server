package cacheService

import (
	"i9lyfe/src/appGlobals"

	"github.com/redis/go-redis/v9"
)

func rdb() *redis.Client {
	return appGlobals.RedisClient
}
