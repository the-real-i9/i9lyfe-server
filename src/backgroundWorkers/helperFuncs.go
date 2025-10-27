package backgroundWorkers

import (
	"context"
	"encoding/json"
	"fmt"
	"i9lyfe/src/helpers"
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
)

func loadUndoneTasks(ctx context.Context, rdb *redis.Client, streamName, streamMsgId string, numTasks int, undoneTasksMan *sync.Map) error {
	undTasks, err := rdb.SMembers(ctx, fmt.Sprintf("stream:%s:msg:%s:undone_tasks", streamName, streamMsgId)).Result()
	if err == redis.Nil {
		for i := range numTasks {
			// by default all tasks are undone
			undoneTasksMan.Store(fmt.Sprintf("task%s", i+1), true)
		}
		return nil
	}

	if err != nil {
		helpers.LogError(err)
		return err
	}

	for _, t := range undTasks {
		undoneTasksMan.Store(t, true)
	}

	return nil
}

func saveUndoneTasks(ctx context.Context, rdb *redis.Client, streamName, streamMsgId string, undoneTasks []any) error {
	rdb.Del(ctx, fmt.Sprintf("stream:%s:msg:%s:undone_tasks", streamName, streamMsgId))

	if err := rdb.SAdd(ctx, fmt.Sprintf("stream:%s:msg:%s:undone_tasks", streamName, streamMsgId), undoneTasks...).Err(); err != nil {
		helpers.LogError(err)
		return err
	}
	return nil
}

func msgProcessingCleanup(ctx context.Context, rdb *redis.Client, streamName, streamMsgId string) {
	rdb.Del(ctx, fmt.Sprintf("stream:%s:msg:%s:undone_tasks", streamName, streamMsgId))
	rdb.Del(ctx, fmt.Sprintf("stream:%s:msg:%s:retry_count", streamName, streamMsgId))
}

func handleFailedMessage(ctx context.Context, rdb *redis.Client, groupName, streamName string, maxRetries int64, dlqName string, stmsg redis.XMessage, err error) {
	retryKey := fmt.Sprintf("stream:%s:msg:%s:retry_count", streamName, stmsg.ID)
	retries, _ := rdb.Incr(ctx, retryKey).Result()
	if retries <= maxRetries {
		log.Printf("will retry (%d/%d): %v", retries, maxRetries, err)
		return // let it stay pending; we’ll pick it again later
	}

	// After max retries, move to DLQ
	b, _ := json.Marshal(stmsg)
	if err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: dlqName,
		Values: map[string]any{"error": err.Error(), "data": string(b)},
	}).Err(); err != nil {
		helpers.LogError(err)
	}

	// Acknowledge to clear from main queue
	if err := rdb.XAck(ctx, streamName, groupName, stmsg.ID).Err(); err != nil {
		helpers.LogError(err)
	}
}
