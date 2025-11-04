package appGlobals

import (
	"cloud.google.com/go/storage"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/redis/go-redis/v9"

	"github.com/jackc/pgx/v5/pgxpool"
)

var GCSClient *storage.Client

var DBPool *pgxpool.Pool

var Neo4jDriver neo4j.DriverWithContext

var RedisClient *redis.Client
