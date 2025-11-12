package eventTypes

import "i9lyfe/src/appTypes"

type NewUserEvent struct {
	Username string
	UserData string
}

type EditUserEvent struct {
	Username    string
	UpdateKVMap map[string]any
}

type UserFollowEvent struct {
	FollowerUser  appTypes.ClientUser
	FollowingUser string
	At            int64
}

type UserUnfollowEvent struct {
	FollowerUser  string
	FollowingUser string
}

type NewPostEvent struct {
	OwnerUser appTypes.ClientUser
	PostId    string
	PostData  string
	Hashtags  []string
	Mentions  []string
	At        int64
}

type PostDeletionEvent struct {
	OwnerUser string
	PostId    string
	Mentions  []*string
}

type PostReactionEvent struct {
	ReactorUser appTypes.ClientUser
	PostOwner   string
	PostId      string
	Emoji       string
	At          int64
}

type PostReactionRemovedEvent struct {
	ReactorUser string
	PostId      string
}

type PostCommentEvent struct {
	CommenterUser appTypes.ClientUser
	PostId        string
	PostOwner     string
	CommentId     string
	CommentData   string
	Mentions      []string
	At            int64
}

type PostCommentRemovedEvent struct {
	CommenterUser string
	PostId        string
	CommentId     string
}

type CommentReactionEvent struct {
	ReactorUser  appTypes.ClientUser
	CommentId    string
	CommentOwner string
	Emoji        string
	At           int64
}

type CommentReactionRemovedEvent struct {
	ReactorUser string
	CommentId   string
}

type CommentCommentEvent struct {
	CommenterUser      appTypes.ClientUser
	ParentCommentId    string
	ParentCommentOwner string
	CommentId          string
	CommentData        string
	Mentions           []string
	At                 int64
}

type CommentCommentRemovedEvent struct {
	CommenterUser   appTypes.ClientUser
	ParentCommentId string
	CommentId       string
}

type RepostEvent struct {
	ReposterUser appTypes.ClientUser
	PostId       string
	PostOwner    string
	RepostId     string
	RepostData   string
	At           int64
}

type PostSaveEvent struct {
	SaverUser string
	PostId    string
}

type PostUnsaveEvent struct {
	SaverUser string
	PostId    string
}

type NewMessageEvent struct {
	FromUser string
	ToUser   string
	CHEId    string
	MsgData  string
}

type NewMsgReactionEvent struct {
	FromUser string
	ToUser   string
	CHEId    string
	RxnData  string
	ToMsgId  string
	Emoji    string
}

type MsgAckEvent struct {
	CHEId string
	Ack   string
	At    int64
}

type MsgDeletionEvent struct {
	CHEId string
	For   string
}

type MsgReactionRemovedEvent struct {
	FromUser string
	ToUser   string
	ToMsgId  string
	CHEId    string
}
