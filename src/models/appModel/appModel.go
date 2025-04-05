package appModel

import (
	"context"
	"i9lyfe/src/models/db"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TopUsers(ctx context.Context) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
    MATCH (user:User)<-[:FOLLOWS_USER]-(follower:User)

		WITH user, count(follower) AS followers_count
    ORDER BY followers_count DESC
    LIMIT 500
    RETURN collect(user { .username, .name, .profile_pic_url, .bio }) AS top_users
    `,
		nil,
	)
	if err != nil {
		log.Println("appModel.go: TopUsers:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	topUsers, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "top_users")

	return topUsers, nil
}

func SearchUsers(ctx context.Context, term string) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
    MATCH (user:User WHERE user.username CONTAINS $term OR user.name CONTAINS $term)

    ORDER BY user.username, user.name
    LIMIT 500
    RETURN collect(user { .username, .name, .profile_pic_url, .bio }) AS res_users
    `,
		map[string]any{
			"term": term,
		},
	)
	if err != nil {
		log.Println("appModel.go: SearchUsers:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	resUsers, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "res_users")

	return resUsers, nil
}

func TopHashtags(ctx context.Context) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
    MATCH (ht:Hashtag)<-[:INCLUDES_HASHTAG]-(post:Post)

		WITH ht.name AS hashtag, count(post) AS posts_count
		ORDER BY post_count DESC
		LIMIT 1000
		RETURN collect({ hashtag, posts_count }) AS top_hashtags
    `,
		nil,
	)
	if err != nil {
		log.Println("appModel.go: TopHashtags:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	topUsers, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "top_hashtags")

	return topUsers, nil
}

func SearchHashtags(ctx context.Context, term string) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
    MATCH (ht:Hashtag WHERE ht.name CONTAINS $term)<-[:INCLUDES_HASHTAG]-(post:Post)

		WITH ht.name AS hashtag, count(post) AS posts_count
		ORDER BY post_count DESC
		LIMIT 1000
		RETURN collect({ hashtag, posts_count }) AS res_hashtags
    `,
		map[string]any{
			"term": term,
		},
	)
	if err != nil {
		log.Println("appModel.go: SearchHashtags:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	topUsers, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "top_hashtags")

	return topUsers, nil
}

func SearchPost(ctx context.Context, clientUsername, postType, term string) ([]any, error) {
	var query = ""

	if clientUsername != "" {
		query = `
			MATCH (clientUser:User{ username: $client_username })
			MATCH (ownerUser:User)-[:CREATES_POST]->(post:Post { type: $type } WHERE post.description CONTAINS $term)
	
			OPTIONAL MATCH (clientUser)-[crxn:REACTS_TO_POST]->(post)
			OPTIONAL MATCH (clientUser)-[csaves:SAVES_POST]->(post)
			OPTIONAL MATCH (clientUser)-[creposts:REPOSTS_POST]->(post)
	
			WITH post, 
				toString(post.created_at) AS created_at, 
				ownerUser { .username, .profile_pic_url } AS owner_user,
				CASE crxn 
					WHEN IS NULL THEN "" 
					ELSE crxn.reaction 
				END AS client_reaction, 
				CASE csaves 
					WHEN IS NULL THEN false 
					ELSE true 
				END AS client_saved, 
				CASE creposts 
					WHEN IS NULL THEN false 
					ELSE true 
				END AS client_reposted
			ORDER BY post.created_at DESC
			LIMIT 500
			RETURN collect(post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted }) AS res_posts
		`
	} else {
		query = `
			MATCH (ownerUser:User)-[:CREATES_POST]->(post:Post { type: $type } WHERE post.description CONTAINS $term)
	
			WITH post, toString(post.created_at) AS created_at, ownerUser { .username, .profile_pic_url } AS owner_user
			ORDER BY post.created_at DESC
			LIMIT 500
			RETURN collect(post { .*, owner_user, created_at, client_reaction: "", client_saved: false, client_reposted: false }) AS res_posts
		`
	}

	res, err := db.Query(
		ctx,
		query,
		map[string]any{
			"type":            postType,
			"term":            term,
			"client_username": clientUsername,
		},
	)
	if err != nil {
		log.Println("contentRecommendationService.go: FetchPosts:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	resPosts, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "res_posts")

	return resPosts, nil
}
