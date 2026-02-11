package UITypes

type ClientUser struct {
	Username      string `json:"username"`
	Name          string `json:"name"`
	ProfilePicUrl string `json:"profile_pic_url"`
}

type NotifUser struct {
	Username      string `json:"username"`
	Name          string `json:"name"`
	ProfilePicUrl string `json:"profile_pic_url"`
}

type ContentOwnerUser struct {
	Username      string `json:"username"`
	Name          string `json:"name"`
	ProfilePicUrl string `json:"profile_pic_url"`
}

type Post struct {
	Id             string   `json:"id" db:"id_"`
	Type           string   `json:"type" db:"type_"`
	OwnerUser      any      `json:"owner_user" db:"owner_user"`
	ReposterUser   any      `json:"reposter_user,omitempty" db:"reposted_by_user"`
	MediaUrls      []string `json:"media_urls" db:"media_urls"`
	Description    string   `json:"description"`
	CreatedAt      int64    `json:"created_at" db:"created_at"`
	ReactionsCount int64    `json:"reactions_count"`
	CommentsCount  int64    `json:"comments_count"`
	RepostsCount   int64    `json:"reposts_count"`
	SavesCount     int64    `json:"saves_count"`
	MeReaction     string   `json:"me_reaction"`
	MeSaved        bool     `json:"me_saved"`
	MeReposted     bool     `json:"me_reposted"`
	Cursor         float64  `json:"cursor,omitempty" db:"snum"`
}

type Comment struct {
	Id             string  `json:"id" db:"comment_id"`
	OwnerUser      any     `json:"owner_user" db:"owner_user"`
	AttachmentUrl  string  `json:"attachment_url" db:"attachment_url"`
	CommentText    string  `json:"comment_text" db:"comment_text"`
	At             int64   `json:"at" db:"at_"`
	ReactionsCount int64   `json:"reactions_count"`
	CommentsCount  int64   `json:"comments_count"`
	MeReaction     string  `json:"me_reaction"`
	Cursor         float64 `json:"cursor,omitempty" db:"snum"`
}

type UserSnippet struct {
	Username      string  `json:"username"`
	Name          string  `json:"name"`
	ProfilePicUrl string  `json:"profile_pic_url"`
	Bio           string  `json:"bio"`
	MeFollow      bool    `json:"me_follow"`
	FollowsMe     bool    `json:"follows_me"`
	Cursor        float64 `json:"cursor"`
}

type ReactorSnippet struct {
	Username      string  `json:"username"`
	Name          string  `json:"name"`
	ProfilePicUrl string  `json:"profile_pic_url"`
	Emoji         string  `json:"emoji"`
	Cursor        float64 `json:"cursor"`
}

type UserProfile struct {
	Username        string `json:"username"`
	Name            string `json:"name"`
	ProfilePicUrl   string `json:"profile_pic_url"`
	Bio             string `json:"bio"`
	PostsCount      int64  `json:"posts_count"`
	FollowersCount  int64  `json:"followers_count"`
	FollowingsCount int64  `json:"followings_count"`
	MeFollow        bool   `json:"me_follow"`
	FollowsMe       bool   `json:"follows_me"`
}

type NotifSnippet struct {
	Id      string         `json:"id"`
	Type    string         `json:"type"`
	At      int64          `json:"at"`
	Details map[string]any `json:"details"`
	Unread  bool           `json:"unread"`
	Cursor  float64        `json:"cursor"`
}

type ChatPartnerUser struct {
	Username      string `json:"username"`
	ProfilePicUrl string `json:"profile_pic_url"`
}

type ChatSnippet struct {
	PartnerUser any     `json:"partner_user" db:"partner_user"`
	UnreadMC    int64   `json:"unread_messages_count"`
	Cursor      float64 `json:"cursor"`
}

type MsgReactor struct {
	Username      string `json:"username"`
	ProfilePicUrl string `json:"profile_pic_url"`
}

type MsgSender struct {
	Username      string `json:"username"`
	ProfilePicUrl string `json:"profile_pic_url"`
}

type MsgReaction struct {
	Emoji   string     `json:"emoji"`
	Reactor MsgReactor `json:"reactor"`
}

type ChatHistoryEntry struct {
	// appears always
	CHEType string `json:"che_type" db:"type_"`

	// appears for message che_type
	Id             string         `json:"id,omitempty" db:"id_"`
	Content        map[string]any `json:"content,omitempty" db:"content_"`
	DeliveryStatus string         `json:"delivery_status,omitempty" db:"delivery_status"`
	CreatedAt      int64          `json:"created_at,omitempty" db:"created_at"`
	DeliveredAt    int64          `json:"delivered_at,omitempty" db:"delivered_at"`
	ReadAt         int64          `json:"read_at,omitempty" db:"read_at"`
	Sender         any            `json:"sender,omitempty" db:"sender_username"`
	ReactionsCount map[string]int `json:"reactions_count,omitempty"`
	Reactions      []MsgReaction  `json:"reactions,omitempty"`

	// appears if che_type:message is a reply
	ReplyTargetMsg map[string]any `json:"reply_target_msg,omitempty" db:"reply_target_msg"`

	// appears for reaction che_type
	Reactor any    `json:"reactor,omitempty" db:"reactor_username"`
	Emoji   string `json:"emoji,omitempty" db:"emoji"`

	// cursor for pagination
	Cursor float64 `json:"cursor"`
}
