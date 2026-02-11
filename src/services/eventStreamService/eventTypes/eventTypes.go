package eventTypes

import "i9lyfe/src/appTypes"

type NewUserEvent struct {
	Username string `redis:"username" json:"username"`
	UserData string `redis:"userData" json:"userData"`
}

type EditUserEvent struct {
	Username    string              `redis:"username" json:"username"`
	UpdateKVMap appTypes.BinableMap `redis:"updateKVMap" json:"updateKVMap"`
}

type UserPresenceChangeEvent struct {
	Username string `redis:"username" json:"username"`
	Presence string `redis:"presence" json:"presence"`
	LastSeen int64  `redis:"lastSeen" json:"lastSeen"`
}

type UserFollowEvent struct {
	FollowerUser  string `redis:"followerUser" json:"followerUser"`
	FollowingUser string `redis:"followingUser" json:"followingUser"`
	At            int64  `redis:"at" json:"at"`
}

type UserUnfollowEvent struct {
	FollowerUser  string `redis:"followerUser" json:"followerUser"`
	FollowingUser string `redis:"followingUser" json:"followingUser"`
}

type NewPostEvent struct {
	OwnerUser string                `redis:"ownerUser" json:"ownerUser"`
	PostId    string                `redis:"postId" json:"postId"`
	PostData  string                `redis:"postData" json:"postData"`
	Hashtags  appTypes.BinableSlice `redis:"hashtags" json:"hashtags"`
	Mentions  appTypes.BinableSlice `redis:"mentions" json:"mentions"`
	Score     float64               `redis:"score" json:"score"`
	At        int64                 `redis:"at" json:"at"`
}

type PostDeletionEvent struct {
	OwnerUser string                `redis:"ownerUser" json:"ownerUser"`
	PostId    string                `redis:"postId" json:"postId"`
	Mentions  appTypes.BinableSlice `redis:"mentions" json:"mentions"`
}

type PostReactionEvent struct {
	ReactorUser string `redis:"reactorUser" json:"reactorUser"`
	PostOwner   string `redis:"postOwner" json:"postOwner"`
	PostId      string `redis:"postId" json:"postId"`
	Emoji       string `redis:"emoji" json:"emoji"`
	At          int64  `redis:"at" json:"at"`
}

type PostReactionRemovedEvent struct {
	ReactorUser string `redis:"reactorUser" json:"reactorUser"`
	PostId      string `redis:"postId" json:"postId"`
}

type PostCommentEvent struct {
	CommenterUser string                `redis:"commenterUser" json:"commenterUser"`
	PostId        string                `redis:"postId" json:"postId"`
	PostOwner     string                `redis:"postOwner" json:"postOwner"`
	CommentId     string                `redis:"commentId" json:"commentId"`
	CommentData   string                `redis:"commentData" json:"commentData"`
	Mentions      appTypes.BinableSlice `redis:"mentions" json:"mentions"`
	At            int64                 `redis:"at" json:"at"`
	Score         float64               `redis:"score" json:"score"`
}

type PostCommentRemovedEvent struct {
	CommenterUser string `redis:"commenterUser" json:"commenterUser"`
	PostId        string `redis:"postId" json:"postId"`
	CommentId     string `redis:"commentId" json:"commentId"`
}

type CommentReactionEvent struct {
	ReactorUser  string `redis:"reactorUser" json:"reactorUser"`
	CommentId    string `redis:"commentId" json:"commentId"`
	CommentOwner string `redis:"commentOwner" json:"commentOwner"`
	Emoji        string `redis:"emoji" json:"emoji"`
	At           int64  `redis:"at" json:"at"`
}

type CommentReactionRemovedEvent struct {
	ReactorUser string `redis:"reactorUser" json:"reactorUser"`
	CommentId   string `redis:"commentId" json:"commentId"`
}

type CommentCommentEvent struct {
	CommenterUser      string                `redis:"commenterUser" json:"commenterUser"`
	ParentCommentId    string                `redis:"parentCommentId" json:"parentCommentId"`
	ParentCommentOwner string                `redis:"parentCommentOwner" json:"parentCommentOwner"`
	CommentId          string                `redis:"commentId" json:"commentId"`
	CommentData        string                `redis:"commentData" json:"commentData"`
	Mentions           appTypes.BinableSlice `redis:"mentions" json:"mentions"`
	At                 int64                 `redis:"at" json:"at"`
	Score              float64               `redis:"score" json:"score"`
}

type CommentCommentRemovedEvent struct {
	CommenterUser   string `redis:"commenterUser" json:"commenterUser"`
	ParentCommentId string `redis:"parentCommentId" json:"parentCommentId"`
	CommentId       string `redis:"commentId" json:"commentId"`
}

type RepostEvent struct {
	ReposterUser string  `redis:"reposterUser" json:"reposterUser"`
	PostId       string  `redis:"postId" json:"postId"`
	PostOwner    string  `redis:"postOwner" json:"postOwner"`
	RepostId     string  `redis:"repostId" json:"repostId"`
	RepostData   string  `redis:"repostData" json:"repostData"`
	Score        float64 `redis:"score" json:"score"`
	At           int64   `redis:"at" json:"at"`
}

type PostSaveEvent struct {
	SaverUser string `redis:"saverUser" json:"saverUser"`
	PostId    string `redis:"postId" json:"postId"`
}

type PostUnsaveEvent struct {
	SaverUser string `redis:"saverUser" json:"saverUser"`
	PostId    string `redis:"postId" json:"postId"`
}

type NewMessageEvent struct {
	FirstFromUser bool    `redis:"ffu" json:"ffu"`
	FirstToUser   bool    `redis:"ftu" json:"ftu"`
	FromUser      string  `redis:"fromUser" json:"fromUser"`
	ToUser        string  `redis:"toUser" json:"toUser"`
	CHEId         string  `redis:"CHEId" json:"CHEId"`
	MsgData       string  `redis:"msgData" json:"msgData"`
	Score         float64 `redis:"score" json:"score"`
}

type NewMsgReactionEvent struct {
	FromUser string  `redis:"fromUser" json:"fromUser"`
	ToUser   string  `redis:"toUser" json:"toUser"`
	CHEId    string  `redis:"CHEId" json:"CHEId"`
	RxnData  string  `redis:"rxnData" json:"rxnData"`
	ToMsgId  string  `redis:"toMsgId" json:"toMsgId"`
	Emoji    string  `redis:"emoji" json:"emoji"`
	Score    float64 `redis:"score" json:"score"`
}

type MsgsAckEvent struct {
	FromUser  string                `redis:"fromUser" json:"fromUser"`
	ToUser    string                `redis:"toUser" json:"toUser"`
	CHEIdList appTypes.BinableSlice `redis:"cheIdList" json:"cheIdList"`
	Ack       string                `redis:"ack" json:"ack"`
	At        int64                 `redis:"at" json:"at"`
	Score     float64               `redis:"score" json:"score"`
}

type MsgDeletionEvent struct {
	CHEId string `redis:"CHEId" json:"CHEId"`
	For   string `redis:"for" json:"for"`
}

type MsgReactionRemovedEvent struct {
	FromUser string `redis:"fromUser" json:"fromUser"`
	ToUser   string `redis:"toUser" json:"toUser"`
	ToMsgId  string `redis:"toMsgId" json:"toMsgId"`
	CHEId    string `redis:"CHEId" json:"CHEId"`
}
