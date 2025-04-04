package contentRecommendationService

import (
	"context"
	"i9lyfe/src/models/db"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

/*
- Obviously, the Content Recommendation Algorithm here isn't what you'll call sophisticated
- But for now, this is where it's at. It promises to improve with subsequent iterations
*/

func RecommendPost(clientUsername, postId string) any {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res, err := db.Query(
		ctx,
		`
    MATCH (post:Post{ id: $post_id })<-[:CREATES_POST]-(ownerUser:User)
    WHERE EXISTS {
      MATCH (:User{ username: $client_username })-[:FOLLOWS_USER]->(ownerUser)
      UNION
      MATCH (:User{ username: $client_username })-[:FOLLOWS_USER]->(:User)-[:FOLLOWS_USER]->(ownerUser)
    }
    WITH post, toString(post.created_at) AS created_at, ownerUser { .username, .profile_pic_url } AS owner_user
    RETURN post { .*, created_at, owner_user } AS the_post
    `,
		map[string]any{
			"post_id":         postId,
			"client_username": clientUsername,
		},
	)
	if err != nil {
		log.Println("contentRecommendationService.go: GetPost:", err)
	}

	if len(res.Records) == 0 {
		return nil
	}

	thePost, _ := res.Records[0].Get("the_post")

	return thePost
}

func GetHomePosts(ctx context.Context, clientUsername string, types []string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
    MATCH (clientUser:User{ username: $client_username })
    MATCH (ownerUser:User)-[:CREATES_POST]->(post:Post WHERE post.type IN $types AND post.created_at < $offset)
    WHERE EXISTS {
      MATCH (clientUser)-[:FOLLOWS_USER]->(ownerUser)
      UNION
      MATCH (clientUser)-[:FOLLOWS_USER]->(:User)-[:FOLLOWS_USER]->(ownerUser)
    }

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
    ORDER BY post.created_at
    LIMIT $limit
    RETURN collect(post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted }) AS feed_posts
    `,
		map[string]any{
			"types":           types,
			"client_username": clientUsername,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		log.Println("contentRecommendationService.go: GetHomePosts:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	feedPosts, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "feed_posts")

	return feedPosts, nil
}

func GetExplorePosts(ctx context.Context, clienUsername string, types []string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
    MATCH (ownerUser:User)-[:CREATES_POST]->(post:Post WHERE post.type IN $types AND post.created_at < $offset), (clientUser:User{ username: $client_username })

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
    ORDER BY post.created_at DESC, post.reactions_count DESC, post.comments_count DESC
    LIMIT $limit
    RETURN collect(post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted }) AS explore_posts
    `,
		map[string]any{
			"types":           types,
			"client_username": clienUsername,
			"limit":           limit,
			"offset":          offset,
		},
	)

	if err != nil {
		log.Println("contentRecommendationService.go: GetExplorePosts:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	explorePosts, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "explore_posts")

	return explorePosts, nil
}
