package initializers

import (
	"context"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	"os"

	"cloud.google.com/go/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"

	"github.com/redis/go-redis/v9"
)

func initGCSClient() error {
	stClient, err := storage.NewClient(context.Background(), option.WithAPIKey(os.Getenv("GCS_API_KEY")))
	if err != nil {
		return err
	}

	appGlobals.GCSClient = stClient

	return nil
}

func initDBPool() error {
	pool, err := pgxpool.New(context.Background(), os.Getenv("PGDATABASE_URL"))
	if err != nil {
		return err
	}

	appGlobals.DBPool = pool

	return nil
}

func initRedisClient() error {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       0,
	})

	appGlobals.RedisClient = client

	return nil
}

func InitApp() error {

	if os.Getenv("GO_ENV") == "" {
		if err := godotenv.Load(".env"); err != nil {
			return err
		}
	}

	if err := initNeo4jDriver(); err != nil {
		return err
	}

	if err := initDBPool(); err != nil {
		return err
	}

	if err := initGCSClient(); err != nil {
		return err
	}

	if err := initRedisClient(); err != nil {
		return err
	}

	return nil
}

func CleanUp() {
	if err := appGlobals.Neo4jDriver.Close(context.TODO()); err != nil {
		helpers.LogError(err)
	}

	appGlobals.DBPool.Close()
}
