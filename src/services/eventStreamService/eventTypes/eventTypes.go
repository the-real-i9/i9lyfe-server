package eventTypes

import "i9lyfe/src/appTypes"

type NewUserEvent struct {
	Username string `redis:"username" msgpack:"username"`
	UserData string `redis:"userData" msgpack:"userData"`
}

type EditUserEvent struct {
	Username    string              `redis:"username" msgpack:"username"`
	UpdateKVMap appTypes.BinableMap `redis:"updateKVMap" msgpack:"updateKVMap"`
}

type UserPresenceChangeEvent struct {
	Username string `redis:"username" msgpack:"username"`
	Presence string `redis:"presence" msgpack:"presence"`
	LastSeen int64  `redis:"lastSeen" msgpack:"lastSeen"`
}

type UserFollowEvent struct {
	FollowerUser  string `redis:"followerUser" msgpack:"followerUser"`
	FollowingUser string `redis:"followingUser" msgpack:"followingUser"`
	At            int64  `redis:"at" msgpack:"at"`
	FollowCursor  int64  `redis:"followCursor" msgpack:"followCursor"`
}

type UserUnfollowEvent struct {
	FollowerUser  string `redis:"followerUser" msgpack:"followerUser"`
	FollowingUser string `redis:"followingUser" msgpack:"followingUser"`
}

type NewPostEvent struct {
	OwnerUser  string                `redis:"ownerUser" msgpack:"ownerUser"`
	PostId     string                `redis:"postId" msgpack:"postId"`
	PostData   string                `redis:"postData" msgpack:"postData"`
	Hashtags   appTypes.BinableSlice `redis:"hashtags" msgpack:"hashtags"`
	Mentions   appTypes.BinableSlice `redis:"mentions" msgpack:"mentions"`
	At         int64                 `redis:"at" msgpack:"at"`
	PostCursor int64                 `redis:"postCursor" msgpack:"postCursor"`
}

type PostDeletionEvent struct {
	OwnerUser string                `redis:"ownerUser" msgpack:"ownerUser"`
	PostId    string                `redis:"postId" msgpack:"postId"`
	Mentions  appTypes.BinableSlice `redis:"mentions" msgpack:"mentions"`
}

type PostReactionEvent struct {
	ReactorUser string `redis:"reactorUser" msgpack:"reactorUser"`
	PostOwner   string `redis:"postOwner" msgpack:"postOwner"`
	PostId      string `redis:"postId" msgpack:"postId"`
	Emoji       string `redis:"emoji" msgpack:"emoji"`
	At          int64  `redis:"at" msgpack:"at"`
	RxnCursor   int64  `redis:"rxnCursor" msgpack:"rxnCursor"`
}

type PostReactionRemovedEvent struct {
	ReactorUser string `redis:"reactorUser" msgpack:"reactorUser"`
	PostId      string `redis:"postId" msgpack:"postId"`
}

type PostCommentEvent struct {
	CommenterUser string                `redis:"commenterUser" msgpack:"commenterUser"`
	PostId        string                `redis:"postId" msgpack:"postId"`
	PostOwner     string                `redis:"postOwner" msgpack:"postOwner"`
	CommentId     string                `redis:"commentId" msgpack:"commentId"`
	CommentData   string                `redis:"commentData" msgpack:"commentData"`
	Mentions      appTypes.BinableSlice `redis:"mentions" msgpack:"mentions"`
	At            int64                 `redis:"at" msgpack:"at"`
	CommentCursor int64                 `redis:"commentCursor" msgpack:"commentCursor"`
}

type PostCommentRemovedEvent struct {
	CommenterUser string `redis:"commenterUser" msgpack:"commenterUser"`
	PostId        string `redis:"postId" msgpack:"postId"`
	CommentId     string `redis:"commentId" msgpack:"commentId"`
}

type CommentReactionEvent struct {
	ReactorUser  string `redis:"reactorUser" msgpack:"reactorUser"`
	CommentId    string `redis:"commentId" msgpack:"commentId"`
	CommentOwner string `redis:"commentOwner" msgpack:"commentOwner"`
	Emoji        string `redis:"emoji" msgpack:"emoji"`
	At           int64  `redis:"at" msgpack:"at"`
	RxnCursor    int64  `redis:"rxnCursor" msgpack:"rxnCursor"`
}

type CommentReactionRemovedEvent struct {
	ReactorUser string `redis:"reactorUser" msgpack:"reactorUser"`
	CommentId   string `redis:"commentId" msgpack:"commentId"`
}

type CommentCommentEvent struct {
	CommenterUser      string                `redis:"commenterUser" msgpack:"commenterUser"`
	ParentCommentId    string                `redis:"parentCommentId" msgpack:"parentCommentId"`
	ParentCommentOwner string                `redis:"parentCommentOwner" msgpack:"parentCommentOwner"`
	CommentId          string                `redis:"commentId" msgpack:"commentId"`
	CommentData        string                `redis:"commentData" msgpack:"commentData"`
	Mentions           appTypes.BinableSlice `redis:"mentions" msgpack:"mentions"`
	At                 int64                 `redis:"at" msgpack:"at"`
	CommentCursor      int64                 `redis:"commentCursor" msgpack:"commentCursor"`
}

type CommentCommentRemovedEvent struct {
	CommenterUser   string `redis:"commenterUser" msgpack:"commenterUser"`
	ParentCommentId string `redis:"parentCommentId" msgpack:"parentCommentId"`
	CommentId       string `redis:"commentId" msgpack:"commentId"`
}

type RepostEvent struct {
	ReposterUser string  `redis:"reposterUser" msgpack:"reposterUser"`
	PostId       string  `redis:"postId" msgpack:"postId"`
	PostOwner    string  `redis:"postOwner" msgpack:"postOwner"`
	RepostId     string  `redis:"repostId" msgpack:"repostId"`
	RepostData   string  `redis:"repostData" msgpack:"repostData"`
	Score        float64 `redis:"score" msgpack:"score"`
	At           int64   `redis:"at" msgpack:"at"`
	RepostCursor int64   `redis:"repostCursor" msgpack:"repostCursor"`
}

type PostSaveEvent struct {
	SaverUser  string `redis:"saverUser" msgpack:"saverUser"`
	PostId     string `redis:"postId" msgpack:"postId"`
	SaveCursor int64  `redis:"saveCursor" msgpack:"saveCursor"`
}

type PostUnsaveEvent struct {
	SaverUser string `redis:"saverUser" msgpack:"saverUser"`
	PostId    string `redis:"postId" msgpack:"postId"`
}

type NewMessageEvent struct {
	FirstFromUser bool   `redis:"ffu" msgpack:"ffu"`
	FirstToUser   bool   `redis:"ftu" msgpack:"ftu"`
	FromUser      string `redis:"fromUser" msgpack:"fromUser"`
	ToUser        string `redis:"toUser" msgpack:"toUser"`
	CHEId         string `redis:"CHEId" msgpack:"CHEId"`
	MsgData       string `redis:"msgData" msgpack:"msgData"`
	CHECursor     int64  `redis:"cheCursor" msgpack:"cheCursor"`
}

type NewMsgReactionEvent struct {
	FromUser  string `redis:"fromUser" msgpack:"fromUser"`
	ToUser    string `redis:"toUser" msgpack:"toUser"`
	CHEId     string `redis:"CHEId" msgpack:"CHEId"`
	RxnData   string `redis:"rxnData" msgpack:"rxnData"`
	ToMsgId   string `redis:"toMsgId" msgpack:"toMsgId"`
	Emoji     string `redis:"emoji" msgpack:"emoji"`
	CHECursor int64  `redis:"cheCursor" msgpack:"cheCursor"`
}

type MsgsAckEvent struct {
	FromUser   string                `redis:"fromUser" msgpack:"fromUser"`
	ToUser     string                `redis:"toUser" msgpack:"toUser"`
	CHEIdList  appTypes.BinableSlice `redis:"cheIdList" msgpack:"cheIdList"`
	Ack        string                `redis:"ack" msgpack:"ack"`
	At         int64                 `redis:"at" msgpack:"at"`
	ChatCursor int64                 `redis:"chatCursor" msgpack:"chatCursor"`
}

type MsgDeletionEvent struct {
	CHEId string `redis:"CHEId" msgpack:"CHEId"`
	For   string `redis:"for" msgpack:"for"`
}

type MsgReactionRemovedEvent struct {
	FromUser string `redis:"fromUser" msgpack:"fromUser"`
	ToUser   string `redis:"toUser" msgpack:"toUser"`
	ToMsgId  string `redis:"toMsgId" msgpack:"toMsgId"`
	CHEId    string `redis:"CHEId" msgpack:"CHEId"`
}
