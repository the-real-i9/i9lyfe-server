package eventTypes

import "i9lyfe/src/types/appTypes"

type NewUserEvent struct {
	Username string `redis:"username"`
	UserData string `redis:"userData"`
}

type EditUserEvent struct {
	Username    string              `redis:"username"`
	UpdateKVMap appTypes.BinableMap `redis:"updateKVMap"`
}

type UserPresenceChangeEvent struct {
	Username string `redis:"username"`
	Presence string `redis:"presence"`
	LastSeen int64  `redis:"lastSeen"`
}

type UserFollowEvent struct {
	FollowerUser  string `redis:"followerUser"`
	FollowingUser string `redis:"followingUser"`
}

type UserUnfollowEvent struct {
	FollowerUser  string `redis:"followerUser"`
	FollowingUser string `redis:"followingUser"`
}

type NewPostEvent struct {
	OwnerUsername string `redis:"ownerUsername"`
	PostId        string `redis:"postId"`
	PostCursor    int64  `redis:"postCursor"`
}

type PostReactionEvent struct {
	PostId string `redis:"postId"`
}

type PostReactionRemovedEvent struct {
	PostId string `redis:"postId"`
}

type PostCommentEvent struct {
	PostId string `redis:"postId"`
}

type PostCommentRemovedEvent struct {
	PostId string `redis:"postId"`
}

type CommentReactionEvent struct {
	CommentId string `redis:"commentId"`
}

type CommentReactionRemovedEvent struct {
	CommentId string `redis:"commentId"`
}

type CommentCommentEvent struct {
	ParentCommentId string `redis:"parentCommentId"`
}

type CommentCommentRemovedEvent struct {
	ParentCommentId string `redis:"parentCommentId"`
}

type RepostEvent struct {
	PostId       string `redis:"postId"`
	ReposterUser string `redis:"reposterUser"`
	RepostId     string `redis:"repostId"`
	RepostCursor int64  `redis:"repostCursor"`
}

type PostSaveEvent struct {
	PostId string `redis:"postId"`
}

type PostUnsaveEvent struct {
	PostId string `redis:"postId"`
}

type NewMessageEvent struct {
	FirstFromUser bool   `redis:"ffu"`
	FirstToUser   bool   `redis:"ftu"`
	FromUser      string `redis:"fromUser"`
	ToUser        string `redis:"toUser"`
	CHEId         string `redis:"CHEId"`
	MsgData       string `redis:"msgData"`
	CHECursor     int64  `redis:"cheCursor"`
}

type NewMsgReactionEvent struct {
	FromUser  string `redis:"fromUser"`
	ToUser    string `redis:"toUser"`
	CHEId     string `redis:"CHEId"`
	RxnData   string `redis:"rxnData"`
	ToMsgId   string `redis:"toMsgId"`
	Emoji     string `redis:"emoji"`
	CHECursor int64  `redis:"cheCursor"`
}

type MsgsAckEvent struct {
	FromUser   string                `redis:"fromUser"`
	ToUser     string                `redis:"toUser"`
	CHEIds     appTypes.BinableSlice `redis:"cHEIds"`
	Ack        string                `redis:"ack"`
	At         int64                 `redis:"at"`
	ChatCursor int64                 `redis:"chatCursor"`
}

type MsgDeletionEvent struct {
	CHEId string `redis:"CHEId"`
	For   string `redis:"for"`
}

type MsgReactionRemovedEvent struct {
	FromUser string `redis:"fromUser"`
	ToUser   string `redis:"toUser"`
	ToMsgId  string `redis:"toMsgId"`
	CHEId    string `redis:"CHEId"`
}
