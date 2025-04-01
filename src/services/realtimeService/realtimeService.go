package realtimeService

import (
	"sync"
)

var (
	AllClientSockets         = &sync.Map{}
	PostUpdateSubscribers    = &sync.Map{}
	CommentUpdateSubscribers = &sync.Map{}
)

func BroadcastNewPost(postId, ownerUsername string) {
	// AllClientSockets.Range()
}
