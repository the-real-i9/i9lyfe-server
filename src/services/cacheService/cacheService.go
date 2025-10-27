package cacheService

func CacheUserFollowers() {
	// key: username
	// value: { followers:[]map, followers_count:int }
}

func CacheUserFollowing() {
	// key: username
	// value: { following:[]map, following_count:int }
}

func CacheNewPost() {

}

func CachePostReactions() {
	// key: post_id
	// value: {reactions:[]map = []map{reactor:map, emoji:string}, total_reactions:int}
}

func CachePostComments() {
	// key: post_id
	// {reactions:[]map = []map{reactor:map, emoji:string}, total_reactions:int}
}
