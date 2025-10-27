package backgroundWorkers

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"
	"i9lyfe/src/models/postModel"
	"i9lyfe/src/services/contentRecommendationService"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

func newPostsStreamBgWorker(rdb *redis.Client) {
	var (
		streamName         = "new_posts"
		groupName          = "new_post_listeners"
		consumerName       = "worker-1"
		numTasks           = 5
		maxRetries   int64 = 3
		dlqName            = "new_posts_dlq"
	)

	ctx := context.Background()

	err := rdb.XGroupCreateMkStream(ctx, streamName, groupName, "$").Err()
	if err != nil && (err.Error() != "BUSYGROUP Consumer Group name already exists") {
		helpers.LogError(err)
		log.Fatal()
	}

	go func() {
		// cache new post and publish to subscribers
		for {
			streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    groupName,
				Consumer: consumerName,
				Streams:  []string{streamName, ">"},
				Count:    50,
				Block:    0,
			}).Result()

			if err != nil {
				helpers.LogError(err)
				continue
			}

			for _, stmsg := range streams[0].Messages {
				if err := processMessage(ctx, rdb, streamName, stmsg, numTasks); err != nil {
					handleFailedMessage(ctx, rdb, groupName, streamName, maxRetries, dlqName, stmsg, err)
					continue
				}

				// acknowledge message
				if err := rdb.XAck(ctx, streamName, groupName, stmsg.ID).Err(); err != nil {
					helpers.LogError(err)
				}
			}
		}
	}()
}

func processMessage(ctx context.Context, rdb *redis.Client, streamName string, stmsg redis.XMessage, numTasks int) error {

	var msg eventTypes.NewPostEvent
	helpers.ToStruct(stmsg.Values, &msg)

	// undone tasks manager
	var undoneTasksMan = new(sync.Map)

	if err := loadUndoneTasks(ctx, rdb, streamName, stmsg.ID, numTasks, undoneTasksMan); err != nil {
		return err
	}

	var wg = new(sync.WaitGroup)

	// all caching operations are idempotent
	wg.Go(func() {
		// task1: cache new post

		// skip task if done
		if _, undone := undoneTasksMan.Load("task1"); !undone {
			return
		}

		if err := rdb.Set(ctx, fmt.Sprintf("post:%s", msg.PostId), msg.PostData, 0).Err(); err != nil {
			helpers.LogError(err)

			undoneTasksMan.Store("task1", true)
			return
		}

		undoneTasksMan.Delete("task1")
	})

	wg.Go(func() {
		// task2: cache hashtag posts

		// skip task if done
		if _, undone := undoneTasksMan.Load("task2"); !undone {
			return
		}

		for _, ht := range msg.Hashtags {
			if err := rdb.SAdd(ctx, fmt.Sprintf("hastag:%s:posts", ht), msg.PostId).Err(); err != nil {
				helpers.LogError(err)

				undoneTasksMan.Store("task2", true)
				return
			}
		}

		undoneTasksMan.Delete("task2")
	})

	wg.Go(func() {
		// task3: cache user mentioned posts and their notifications

		// skip task if done
		if _, undone := undoneTasksMan.Load("task3"); !undone {
			return
		}

		for _, mu := range msg.Mentions {
			if err := rdb.SAdd(ctx, fmt.Sprintf("user:%s:mentioned_posts", mu), msg.PostId).Err(); err != nil {
				helpers.LogError(err)

				undoneTasksMan.Store("task3", true)
				return
			}

			notifUniqueId := fmt.Sprintf("user_%s_mentioned_in_post_%s", mu, msg.PostId)
			notification := helpers.BuildPostMentionNotification(notifUniqueId, msg.PostId, msg.ClientUsername, msg.At)
			if err := rdb.HSet(ctx, fmt.Sprintf("notification:%s", notifUniqueId), notification).Err(); err != nil {
				helpers.LogError(err)

				undoneTasksMan.Store("task3", true)
				return
			}

			if err := rdb.ZAdd(ctx, fmt.Sprintf("user:%s:notifications:%s-%s", mu, time.Now().Year(), time.Now().Month()), redis.Z{
				Score:  float64(time.Now().Unix()),
				Member: notifUniqueId,
			}).Err(); err != nil {
				helpers.LogError(err)

				undoneTasksMan.Store("task3", true)
				return
			}
		}

		undoneTasksMan.Delete("task3")
	})

	wg.Go(func() {
		// task4: create post mentions and hashtags in DB

		// skip task if done
		if _, undone := undoneTasksMan.Load("task4"); !undone {
			return
		}

		err := postModel.NewPostExtras(ctx, msg.ClientUsername, msg.PostId, msg.Mentions, msg.Hashtags, msg.At)
		if err != nil {
			undoneTasksMan.Store("task4", true)
			return
		}

		undoneTasksMan.Delete("task4")
	})

	wg.Go(func() {
		// task5: publish new post to subscribers | publish notifications event message

		// skip task if done
		if _, undone := undoneTasksMan.Load("task5"); !undone {
			return
		}

		contentRecommendationService.FanOutPost(msg.PostId)

		for _, mu := range msg.Mentions {
			notifUniqueId := fmt.Sprintf("mentioned_%s_in_post_%s", mu, msg.PostId)
			notification := helpers.BuildPostMentionNotification(notifUniqueId, msg.PostId, msg.ClientUsername, msg.At)

			notification["notif"] = helpers.Json2Map(notification["notif"].(string))

			realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
				Event: "new notification",
				Data:  notification,
			})
		}

		undoneTasksMan.Delete("task5")
	})
	wg.Wait()

	stillUndoneTasks := []any{}

	undoneTasksMan.Range(func(key, value any) bool {
		stillUndoneTasks = append(stillUndoneTasks, key)
		return true
	})

	if len(stillUndoneTasks) > 0 {
		if err := saveUndoneTasks(ctx, rdb, streamName, stmsg.ID, stillUndoneTasks); err != nil {
			return err
		}

		return fmt.Errorf("undone tasks")
	}

	msgProcessingCleanup(ctx, rdb, streamName, stmsg.ID)

	return nil
}
