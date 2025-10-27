package eventTypes

import "time"

type NewPostEvent struct {
	ClientUsername string         `json:"clientUsername"`
	PostId         string         `json:"postId"`
	PostData       map[string]any `json:"postData"`
	Hashtags       []string       `json:"hashtags"`
	Mentions       []string       `json:"mentions"`
	At             time.Time      `json:"at"`
}
