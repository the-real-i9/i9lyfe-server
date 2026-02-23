package UITypes

type ClientUser struct {
	Username      string `msgpack:"username"`
	Name          string `msgpack:"name"`
	ProfilePicUrl string `msgpack:"profile_pic_url"`
}

type NotifUser struct {
	Username      string `msgpack:"username"`
	Name          string `msgpack:"name"`
	ProfilePicUrl string `msgpack:"profile_pic_url"`
}

type ContentOwnerUser struct {
	Username      string `msgpack:"username"`
	Name          string `msgpack:"name"`
	ProfilePicUrl string `msgpack:"profile_pic_url"`
}

type Post struct {
	Id             string   `msgpack:"id"`
	Type           string   `msgpack:"type"`
	OwnerUser      any      `msgpack:"owner_user"`
	ReposterUser   any      `msgpack:"reposter_user,omitempty"`
	MediaUrls      []string `msgpack:"media_urls"`
	Description    string   `msgpack:"description"`
	CreatedAt      int64    `msgpack:"created_at"`
	ReactionsCount int64    `msgpack:"reactions_count"`
	CommentsCount  int64    `msgpack:"comments_count"`
	RepostsCount   int64    `msgpack:"reposts_count"`
	SavesCount     int64    `msgpack:"saves_count"`
	MeReaction     string   `msgpack:"me_reaction"`
	MeSaved        bool     `msgpack:"me_saved"`
	MeReposted     bool     `msgpack:"me_reposted"`
	Cursor         int64    `msgpack:"cursor"`
}

type Comment struct {
	Id             string `msgpack:"id" db:"comment_id"`
	OwnerUser      any    `msgpack:"owner_user" db:"owner_user"`
	AttachmentUrl  string `msgpack:"attachment_url" db:"attachment_url"`
	CommentText    string `msgpack:"comment_text" db:"comment_text"`
	At             int64  `msgpack:"at" db:"at_"`
	ReactionsCount int64  `msgpack:"reactions_count"`
	CommentsCount  int64  `msgpack:"comments_count"`
	MeReaction     string `msgpack:"me_reaction"`
	Cursor         int64  `msgpack:"cursor"`
}

type UserSnippet struct {
	Username      string `msgpack:"username"`
	Name          string `msgpack:"name"`
	ProfilePicUrl string `msgpack:"profile_pic_url"`
	Bio           string `msgpack:"bio"`
	MeFollow      bool   `msgpack:"me_follow"`
	FollowsMe     bool   `msgpack:"follows_me"`
	Cursor        int64  `msgpack:"cursor"`
}

type ReactorSnippet struct {
	Username      string `msgpack:"username"`
	Name          string `msgpack:"name"`
	ProfilePicUrl string `msgpack:"profile_pic_url"`
	Emoji         string `msgpack:"emoji"`
	Cursor        int64  `msgpack:"cursor"`
}

type UserProfile struct {
	Username        string `msgpack:"username"`
	Name            string `msgpack:"name"`
	ProfilePicUrl   string `msgpack:"profile_pic_url"`
	Bio             string `msgpack:"bio"`
	PostsCount      int64  `msgpack:"posts_count"`
	FollowersCount  int64  `msgpack:"followers_count"`
	FollowingsCount int64  `msgpack:"followings_count"`
	MeFollow        bool   `msgpack:"me_follow"`
	FollowsMe       bool   `msgpack:"follows_me"`
}

type NotifSnippet struct {
	Id      string         `msgpack:"id"`
	Type    string         `msgpack:"type"`
	At      int64          `msgpack:"at"`
	Details map[string]any `msgpack:"details"`
	Unread  bool           `msgpack:"unread"`
	Cursor  int64          `msgpack:"cursor"`
}

type ChatPartnerUser struct {
	Username      string `msgpack:"username"`
	ProfilePicUrl string `msgpack:"profile_pic_url"`
}

type ChatSnippet struct {
	PartnerUser any   `msgpack:"partner_user" db:"partner_user"`
	UnreadMC    int64 `msgpack:"unread_messages_count"`
	Cursor      int64 `msgpack:"cursor"`
}

type MsgReactor struct {
	Username      string `msgpack:"username"`
	ProfilePicUrl string `msgpack:"profile_pic_url"`
}

type MsgSender struct {
	Username      string `msgpack:"username"`
	ProfilePicUrl string `msgpack:"profile_pic_url"`
}

type MsgReaction struct {
	Emoji   string     `msgpack:"emoji"`
	Reactor MsgReactor `msgpack:"reactor"`
}

type ChatHistoryEntry struct {
	// appears always
	CHEType string `msgpack:"che_type"`

	// appears for message che_type
	Id             string         `msgpack:"id,omitempty"`
	Content        map[string]any `msgpack:"content,omitempty"`
	DeliveryStatus string         `msgpack:"delivery_status,omitempty"`
	CreatedAt      int64          `msgpack:"created_at,omitempty"`
	DeliveredAt    int64          `msgpack:"delivered_at,omitempty"`
	ReadAt         int64          `msgpack:"read_at,omitempty"`
	Sender         any            `msgpack:"sender,omitempty"`
	ReactionsCount map[string]int `msgpack:"reactions_count,omitempty"`
	Reactions      []MsgReaction  `msgpack:"reactions,omitempty"`

	// appears if che_type:message is a reply
	ReplyTargetMsg map[string]any `msgpack:"reply_target_msg,omitempty"`

	// appears for reaction che_type
	Reactor any    `msgpack:"reactor,omitempty"`
	Emoji   string `msgpack:"emoji,omitempty"`
	ToMsgId string `msgpack:"to_msg_id,omitempty"`

	// cursor for pagination
	Cursor int64 `msgpack:"cursor"`
}
