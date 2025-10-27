package realtimeService

import (
	"i9lyfe/src/appGlobals"
	"sync"
)

var rdb = appGlobals.RedisClient

var AllClientSockets = &sync.Map{}
