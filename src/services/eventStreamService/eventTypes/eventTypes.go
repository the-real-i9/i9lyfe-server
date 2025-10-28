package eventTypes

import "time"

type NewPostEvent struct {
	OwnerUser string         `json:"ownerUser"`
	PostId    string         `json:"postId"`
	PostData  map[string]any `json:"postData"`
	Hashtags  []string       `json:"hashtags"`
	Mentions  []string       `json:"mentions"`
	At        time.Time      `json:"at"`
}

type PostReactionEvent struct {
	ReactorUser string    `json:"reactorUser"`
	PostId      string    `json:"postId"`
	Reaction    string    `json:"reaction"`
	At          time.Time `json:"at"`
}
