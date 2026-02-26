package initializers

import (
	"context"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/backgroundWorkers"
	"i9lyfe/src/helpers"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

func initGCSClient() error {
	stClient, err := storage.NewClient(context.Background())
	if err != nil {
		return err
	}

	appGlobals.GCSClient = stClient

	return nil
}

func initDBPool() error {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("PGDATABASE_URL"))
	if err != nil {
		return err
	}

	if os.Getenv("GO_ENV") == "test" {
		_, err := pool.Exec(ctx /* sql */, `TRUNCATE users * CASCADE`)
		if err != nil {
			return err
		}
	}

	appGlobals.DBPool = pool

	return nil
}

func initRedisClient() error {
	client := redis.NewClient(&redis.Options{
		Addr:         os.Getenv("REDIS_ADDR"),
		Password:     os.Getenv("REDIS_PASS"),
		DB:           0,
		WriteTimeout: 10 * time.Second, // the likelihood of a big write pipeline

		// Explicitly disable maintenance notifications
		// This prevents the client from sending CLIENT MAINT_NOTIFICATIONS ON
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
	})

	if os.Getenv("GO_ENV") == "test" {
		err := client.FlushDB(context.Background()).Err()
		if err != nil {
			return err
		}
	}

	appGlobals.RedisClient = client

	backgroundWorkers.Start(client)

	return nil
}

func InitApp() error {

	if os.Getenv("GO_ENV") == "" {
		if err := godotenv.Load(".env"); err != nil {
			return err
		}
	}

	if os.Getenv("GO_ENV") == "test" {
		if err := godotenv.Load(".env.test"); err != nil {
			return err
		}
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
	appGlobals.DBPool.Close()

	if err := appGlobals.RedisClient.Close(); err != nil {
		helpers.LogError(err)
	}
}
