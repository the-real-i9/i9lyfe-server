package eventTypes

type NewPostEvent struct {
	OwnerUser string
	PostId    string
	PostData  map[string]any
	Hashtags  []string
	Mentions  []string
	At        int64
}

type PostReactionEvent struct {
	ReactorUser  string
	PostOwner    string
	PostId       string
	ReactionData map[string]any
	At           int64
}

type RemovePostReactionEvent struct {
	ReactorUser string
	PostId      string
}

type PostCommentEvent struct {
	CommenterUser string
	PostId        string
	PostOwner     string
	CommentId     string
	CommentData   map[string]any
	Mentions      []string
	At            int64
}

type RemovePostCommentEvent struct {
	CommenterUser string
	PostId        string
	CommentId     string
}

type CommentReactionEvent struct {
	ReactorUser  string
	CommentId    string
	CommentOwner string
	ReactionData map[string]any
	At           int64
}

type RemoveCommentReactionEvent struct {
	ReactorUser string
	CommentId   string
}

type CommentCommentEvent struct {
	CommenterUser      string
	ParentCommentId    string
	ParentCommentOwner string
	CommentId          string
	CommentData        map[string]any
	Mentions           []string
	At                 int64
}

type RemoveCommentCommentEvent struct {
	CommenterUser   string
	ParentCommentId string
	CommentId       string
}

type RepostEvent struct {
	ReposterUser string
	PostId       string
	PostOwner    string
	RepostId     string
	RepostData   map[string]any
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
	MsgId    string
	MsgData  map[string]any
}
