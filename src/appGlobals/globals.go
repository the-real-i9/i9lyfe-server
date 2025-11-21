package appGlobals

import (
	"cloud.google.com/go/storage"
	"github.com/redis/go-redis/v9"

	"github.com/jackc/pgx/v5/pgxpool"
)

var GCSClient *storage.Client

var DBPool *pgxpool.Pool

var RedisClient *redis.Client
