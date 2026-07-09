package pubsubService

import (
	"context"
	"fmt"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/services/sseService"

	"i9lyfe/src/helpers"
	"i9lyfe/src/types/appTypes"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/redis/go-redis/v9"
)

func rdb() *redis.Client {
	return appGlobals.RedisClient
}

func publishContentMetric(ctx context.Context, data any, contentType string) {
	if err := rdb().Publish(ctx, "live_content_metrics", helpers.ToMsgPack(appTypes.ServerEventMsg{
		Event: "latest " + contentType + " metric",
		Data:  data,
	})).Err(); err != nil {
		helpers.LogError(err)
	}
}

func PublishPostMetric(ctx context.Context, data any) {
	publishContentMetric(ctx, data, "post")
}

func PublishCommentMetric(ctx context.Context, data any) {
	publishContentMetric(ctx, data, "comment")
}

func SubscribeToLiveContentMetrics(ctx context.Context, clientUsername string) func() {
	pubsub := rdb().Subscribe(ctx, "live_content_metrics")

	closeFunc := func() {
		if err := pubsub.Close(); err != nil {
			helpers.LogError(err)
		}
	}

	go func(pubsub *redis.PubSub, closeFunc func()) {
		ch := pubsub.Channel()

		for msg := range ch {

			if userSock, ok := sseService.GetUserSocket(clientUsername); ok {
				sock := userSock.(*websocket.Conn)

				if err := sock.WriteMessage(websocket.BinaryMessage, []byte(msg.Payload)); err != nil {
					helpers.LogError(err)
					closeFunc()
				}
			} else {
				closeFunc()
			}
		}
	}(pubsub, closeFunc)

	return closeFunc
}

func PublishUserPresenceChange(ctx context.Context, targetUsername string, data map[string]any) {
	if err := rdb().Publish(ctx, fmt.Sprintf("user_%s_presence_change", targetUsername), helpers.ToMsgPack(appTypes.ServerEventMsg{
		Event: "user presence changed",
		Data:  data,
	})).Err(); err != nil {
		helpers.LogError(err)
	}
}

func SubscribeToUserPresence(ctx context.Context, clientUsername string, targetUsername string) func() {
	pubsub := rdb().Subscribe(ctx, fmt.Sprintf("user_%s_presence_change", targetUsername))

	closeFunc := func() {
		if err := pubsub.Close(); err != nil {
			helpers.LogError(err)
		}
	}

	go func(pubsub *redis.PubSub, closeFunc func()) {
		ch := pubsub.Channel()

		for msg := range ch {
			if userSock, ok := sseService.GetUserSocket(clientUsername); ok {
				sock := userSock.(*websocket.Conn)

				if err := sock.WriteMessage(websocket.BinaryMessage, []byte(msg.Payload)); err != nil {
					helpers.LogError(err)
					closeFunc()
				}
			} else {
				closeFunc()
			}
		}
	}(pubsub, closeFunc)

	return closeFunc
}
