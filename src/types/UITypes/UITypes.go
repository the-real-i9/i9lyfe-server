package UITypes

type ClientUser struct {
	Username      string `msgpack:"username"`
	Name          string `msgpack:"name"`
	ProfilePicUrl string `msgpack:"profile_pic_url"`
}

type ContentOwnerUser struct {
	Username      string `msgpack:"username" db:"username"`
	Name          string `msgpack:"name" db:"name_"`
	ProfilePicUrl string `msgpack:"profile_pic_url" db:"profile_pic_url"`
}

type Post struct {
	Id               string         `msgpack:"id" db:"id_"`
	Type             string         `msgpack:"type" db:"type_"`
	OwnerUser        map[string]any `msgpack:"owner_user" db:"owner_user"`
	ReposterUsername *string        `msgpack:"reposter_username,omitempty" db:"reposter_username"`
	MediaUrls        []string       `msgpack:"media_urls" db:"media_urls"`
	Description      string         `msgpack:"description" db:"description"`
	CreatedAt        int64          `msgpack:"created_at" db:"created_at"`
	ReactionsCount   int64          `msgpack:"reactions_count" db:"reactions_count"`
	CommentsCount    int64          `msgpack:"comments_count" db:"comments_count"`
	RepostsCount     int64          `msgpack:"reposts_count" db:"reposts_count"`
	SavesCount       int64          `msgpack:"saves_count" db:"saves_count"`
	MeReaction       *string        `msgpack:"me_reaction" db:"me_reaction"`
	MeSaved          bool           `msgpack:"me_saved" db:"me_saved"`
	MeReposted       bool           `msgpack:"me_reposted" db:"me_reposted"`
	Cursor           int64          `msgpack:"cursor" db:"cursor_"`
}

type NewPost struct {
	Id               string         `msgpack:"id" db:"id_"`
	Type             string         `msgpack:"type" db:"type_"`
	OwnerUser        map[string]any `msgpack:"owner_user" db:"owner_user"`
	ReposterUsername string         `msgpack:"reposter_username,omitempty" db:"reposter_username"`
	MediaUrls        []string       `msgpack:"media_urls" db:"media_urls"`
	Description      string         `msgpack:"description" db:"description"`
	CreatedAt        int64          `msgpack:"created_at" db:"created_at"`
	ReactionsCount   int64          `msgpack:"reactions_count" db:"reactions_count"`
	CommentsCount    int64          `msgpack:"comments_count" db:"comments_count"`
	RepostsCount     int64          `msgpack:"reposts_count" db:"reposts_count"`
	SavesCount       int64          `msgpack:"saves_count" db:"saves_count"`
	MeReaction       string         `msgpack:"me_reaction" db:"me_reaction"`
	MeSaved          bool           `msgpack:"me_saved" db:"me_saved"`
	MeReposted       bool           `msgpack:"me_reposted" db:"me_reposted"`
	Cursor           int64          `msgpack:"cursor" db:"cursor_"`
	MentNotifIds     []string       `msgpack:"_" db:"ment_notif_ids"`
}

type Comment struct {
	Id             string         `msgpack:"id" db:"comment_id"`
	OwnerUser      map[string]any `msgpack:"owner_user" db:"owner_user"`
	AttachmentUrl  string         `msgpack:"attachment_url" db:"attachment_url"`
	CommentText    string         `msgpack:"comment_text" db:"comment_text"`
	At             int64          `msgpack:"at" db:"at_"`
	ReactionsCount int64          `msgpack:"reactions_count" db:"reactions_count"`
	CommentsCount  int64          `msgpack:"comments_count" db:"comments_count"`
	MeReaction     *string        `msgpack:"me_reaction" db:"me_reaction"`
	Cursor         int64          `msgpack:"cursor" db:"cursor_"`
}

type NewComment struct {
	Id             string         `msgpack:"id" db:"comment_id"`
	OwnerUser      map[string]any `msgpack:"owner_user" db:"owner_user"`
	AttachmentUrl  string         `msgpack:"attachment_url" db:"attachment_url"`
	CommentText    string         `msgpack:"comment_text" db:"comment_text"`
	At             int64          `msgpack:"at" db:"at_"`
	ReactionsCount int64          `msgpack:"reactions_count" db:"reactions_count"`
	CommentsCount  int64          `msgpack:"comments_count" db:"comments_count"`
	MeReaction     string         `msgpack:"me_reaction" db:"me_reaction"`
	Cursor         int64          `msgpack:"cursor" db:"cursor_"`
	MentNotifIds   []string       `msgpack:"_" db:"ment_notif_ids"`
	CommentNotifId string         `msgpack:"_" db:"comment_notif_id"`
}

// for displaying a list of followers and followings
type UserSnippet struct {
	Username      string `msgpack:"username" db:"username"`
	Name          string `msgpack:"name" db:"name_"`
	ProfilePicUrl string `msgpack:"profile_pic_url" db:"profile_pic_url"`
	Bio           string `msgpack:"bio" db:"bio"`
	MeFollow      bool   `msgpack:"me_follow" db:"me_follow"`
	FollowsMe     bool   `msgpack:"follows_me" db:"follows_me"`
	Cursor        int64  `msgpack:"cursor" db:"cursor_"`
}

// for displaying post reactors
type ReactorSnippet struct {
	Username      string `msgpack:"username" db:"username"`
	Name          string `msgpack:"name" db:"name_"`
	ProfilePicUrl string `msgpack:"profile_pic_url" db:"profile_pic_url"`
	Emoji         string `msgpack:"emoji" db:"emoji"`
	Cursor        int64  `msgpack:"cursor" db:"cursor_"`
}

// for showing user profile
type UserProfile struct {
	Username        string `msgpack:"username" db:"username"`
	Name            string `msgpack:"name" db:"name_"`
	ProfilePicUrl   string `msgpack:"profile_pic_url" db:"profile_pic_url"`
	Bio             string `msgpack:"bio" db:"bio"`
	PostsCount      int64  `msgpack:"posts_count" db:"posts_count"`
	FollowersCount  int64  `msgpack:"followers_count" db:"followers_count"`
	FollowingsCount int64  `msgpack:"followings_count" db:"followings_count"`
	MeFollow        bool   `msgpack:"me_follow" db:"me_follow"`
	FollowsMe       bool   `msgpack:"follows_me" db:"follows_me"`
}

type NotifSnippet struct {
	Id            string         `msgpack:"id" db:"id_"`
	Type          string         `msgpack:"type" db:"type_"`
	At            int64          `msgpack:"at" db:"at_"`
	Details       map[string]any `msgpack:"details"`
	Unread        bool           `msgpack:"unread"`
	Cursor        int64          `msgpack:"cursor" db:"cursor_"`
	OwnerUsername string         `msgpack:"_" db:"owner_username"`
}

// for displaying chat lists
type ChatPartnerUser struct {
	Username      string `msgpack:"username"`
	ProfilePicUrl string `msgpack:"profile_pic_url"`
}

type ChatSnippet struct {
	PartnerUser map[string]any `msgpack:"partner_user" db:"partner_user"`
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
	Emoji   string         `msgpack:"emoji"`
	Reactor map[string]any `msgpack:"reactor"`
}

type ChatHistoryEntry struct {
	// appears always

	CHEType string `msgpack:"che_type" db:"type_"`
	Id      string `msgpack:"id" db:"id_"`
	Receipt string `msgpack:"receipt" db:"receipt"`

	// appears for message che_type

	Content        map[string]any `msgpack:"content,omitempty" db:"content_"`
	DeliveryStatus *string        `msgpack:"delivery_status,omitempty" db:"delivery_status"`
	CreatedAt      *int64         `msgpack:"created_at,omitempty" db:"created_at"`
	DeliveredAt    *int64         `msgpack:"delivered_at,omitempty" db:"delivered_at"`
	ReadAt         *int64         `msgpack:"read_at,omitempty" db:"read_at"`
	Sender         map[string]any `msgpack:"sender,omitempty" db:"sender"`

	// appears if che_type:message is a reply

	ReplyTargetMsg map[string]any `msgpack:"reply_target_msg,omitempty" db:"reply_target_msg"`

	// appears for reaction che_type

	Reactor map[string]any `msgpack:"reactor,omitempty" db:"reactor"`
	Emoji   *string        `msgpack:"emoji,omitempty" db:"emoji"`
	ToMsg   map[string]any `msgpack:"to_msg,omitempty" db:"rxn_to_msg"`

	// cursor for pagination

	Cursor int64 `msgpack:"cursor" db:"cursor_"`
}
