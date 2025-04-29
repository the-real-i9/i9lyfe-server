package contentRecommendationService

import (
	"context"
	"fmt"
	"i9lyfe/src/models/db"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

/*
- Bear with me, I'm still working on a content recommendation algorithm.
- This is far from how I envisage it. I just want to have a dummy implementation for now
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

func FetchPosts(ctx context.Context, clientUsername string, types, hashtags []string) ([]any, error) {
	var hashtagCondPatt = ""
	var query = ""

	if len(hashtags) > 0 {
		hashtagCondPatt = " AND EXISTS { (post)-[:INCLUDES_HASHTAG]->(ht:Hashtag WHERE ht.name IN $hashtags) }"
	}

	if clientUsername != "" {
		query = fmt.Sprintf(`
			MATCH (clientUser:User{ username: $client_username })
			MATCH (ownerUser:User)-[:CREATES_POST]->(post:Post)
			WHERE post.type IN $types%s
	
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
		`, hashtagCondPatt)
	} else {
		query = fmt.Sprintf(`
			MATCH (ownerUser:User)-[:CREATES_POST]->(post:Post)
			WHERE post.type IN $types%s
	
			WITH post, toString(post.created_at) AS created_at, ownerUser { .username, .profile_pic_url } AS owner_user
			ORDER BY post.created_at DESC
			LIMIT 500
			RETURN collect(post { .*, owner_user, created_at, client_reaction: "", client_saved: false, client_reposted: false }) AS res_posts
		`, hashtagCondPatt)
	}

	res, err := db.Query(
		ctx,
		query,
		map[string]any{
			"types":           types,
			"hashtags":        hashtags,
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
