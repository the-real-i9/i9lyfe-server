package backgroundWorkers

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func postReactionsStreamBgWorker(rdb *redis.Client) {
	for {
		res, err := rdb.XRead(context.Background(), &redis.XReadArgs{
			Streams: []string{"post_reactions", "$"},
			Count:   1000,
			Block:   0,
		}).Result()

		if err != nil {
			log.Println("error reading stream:", err)
			continue
		}

		for _, stream := range res {
			for _, msg := range stream.Messages {
				// accumulate reactions_count for each unique post_id
				// so that we can update the total post's reactions_count at once in the cache
			}
		}

		// update each post's view in the cache with their reactions count
		// publish update post's reactions_count to subscibers

		timer := time.NewTimer(500 * time.Millisecond)

		<-timer.C
	}

}
