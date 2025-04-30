package realtimeService

import (
	"i9lyfe/src/appTypes"
	"i9lyfe/src/services/contentRecommendationService"
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

var (
	AllClientSockets         = &sync.Map{}
	PostUpdateSubscribers    = &sync.Map{}
	CommentUpdateSubscribers = &sync.Map{}
)

func BroadcastNewPost(postId, ownerUsername string) {
	AllClientSockets.Range(func(key, value any) bool {
		clientUsername := key.(string)
		clientSocket := value.(*websocket.Conn)

		if ownerUsername == clientUsername {
			return true
		}

		post := contentRecommendationService.RecommendPost(clientUsername, postId)

		if w_err := clientSocket.WriteJSON(post); w_err != nil {
			log.Println(w_err)
		}

		return true
	})
}

func SendPostUpdate(data any) {
	PostUpdateSubscribers.Range(func(key, value any) bool {
		clientSocket := value.(*websocket.Conn)

		clientSocket.WriteJSON(appTypes.ServerWSMsg{
			Event: "latest post update",
			Data:  data,
		})

		return true
	})
}

func SendCommentUpdate(data any) {
	CommentUpdateSubscribers.Range(func(key, value any) bool {
		clientSocket := value.(*websocket.Conn)

		clientSocket.WriteJSON(appTypes.ServerWSMsg{
			Event: "latest comment update",
			Data:  data,
		})

		return true
	})
}
