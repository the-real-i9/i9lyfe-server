// TODOs:
//   - Replace Neo4j caching with Redis
package realtimeService

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"

	"github.com/gofiber/contrib/websocket"
)

func publishContentMetric(ctx context.Context, data any, contentType string) {
	if err := rdb().Publish(ctx, "live_content_metrics", helpers.ToJson(appTypes.ServerEventMsg{
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

func SubscribeToLiveContentMetrics(ctx context.Context, clientUsername string, ctxCancel context.CancelFunc) {
	pubsub := rdb().Subscribe(ctx, "live_content_metrics")

	defer func() {
		if err := pubsub.Close(); err != nil {
			helpers.LogError(err)
		}
	}()

	go func(ctxCancel context.CancelFunc) {
		ch := pubsub.Channel()

		for msg := range ch {
			if userPipe, ok := AllClientSockets.Load(clientUsername); ok {
				pipe := userPipe.(*websocket.Conn)

				if err := pipe.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
					helpers.LogError(err)
					ctxCancel()
				}
			}
		}
	}(ctxCancel)
}

func PublishUserPresenceChange(ctx context.Context, targetUsername string, data map[string]any) {
	if err := rdb().Publish(ctx, fmt.Sprintf("user_%s_presence_change", targetUsername), helpers.ToJson(appTypes.ServerEventMsg{
		Event: "user presence changed",
		Data:  data,
	})).Err(); err != nil {
		helpers.LogError(err)
	}
}

func SubscribeToUserPresence(ctx context.Context, clientUsername string, targetUsername string, ctxCancel context.CancelFunc) {
	pubsub := rdb().Subscribe(ctx, fmt.Sprintf("user_%s_presence_change", targetUsername))

	defer func() {
		if err := pubsub.Close(); err != nil {
			helpers.LogError(err)
		}
	}()

	go func(ctxCancel context.CancelFunc) {
		ch := pubsub.Channel()

		for msg := range ch {
			if userPipe, ok := AllClientSockets.Load(clientUsername); ok {
				pipe := userPipe.(*websocket.Conn)

				if err := pipe.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
					helpers.LogError(err)
					ctxCancel()
				}
			}
		}
	}(ctxCancel)
}
