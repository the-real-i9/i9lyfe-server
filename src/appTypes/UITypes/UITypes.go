package UITypes

type NotifUser struct {
	Username      string `json:"username"`
	Name          string `json:"name"`
	ProfilePicUrl string `json:"profile_pic_url"`
}

type PostOwnerUser struct {
	Username      string `json:"username"`
	Name          string `json:"name"`
	ProfilePicUrl string `json:"profile_pic_url"`
}

type Post struct {
	Id             string   `json:"id"`
	Type           string   `json:"type"`
	OwnerUser      any      `json:"owner_user"`
	MediaUrls      []string `json:"media_urls"`
	Description    string   `json:"description"`
	CreatedAt      int64    `json:"created_at"`
	ReactionsCount int64    `json:"reactions_count"`
	CommentsCount  int64    `json:"comments_count"`
	RepostsCount   int64    `json:"reposts_count"`
	SavesCount     int64    `json:"saves_count"`
	MeReaction     string   `json:"me_reaction"`
	MeSaved        bool     `json:"me_saved"`
	MeReposted     bool     `json:"me_reposted"`
	Cursor         float64  `json:"cursor"`
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
