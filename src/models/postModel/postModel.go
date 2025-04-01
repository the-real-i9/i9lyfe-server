package postModel

import (
	"context"
	"i9lyfe/src/helpers"
	"i9lyfe/src/models/db"
	"log"
	"maps"
	"slices"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type NewResT struct {
	NewPostData   map[string]any   `json:"new_post_data"`
	MentionNotifs []map[string]any `json:"mention_notifs"`
}

func New(ctx context.Context, clientUsername string, mediaUrls []string, postType, description string, mentions, hashtags []string) (NewResT, error) {
	var resData NewResT

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 2)

		var (
			res neo4j.ResultWithContext
			err error
			at  = time.Now().UTC()
		)

		res, err = tx.Run(
			ctx,
			`
			MATCH (clientUser:User{ username: $client_username })

      CREATE (clientUser)-[:CREATES_POST]->(post:Post{ id: randomUUID(), type: $type, media_urls: $media_urls, description: $description, created_at: $at, reactions_count: 0, comments_count: 0, reposts_count: 0, saves_count: 0 })
      WITH post, toString(post.created_at) AS created_at, clientUser { .username, .profile_pic_url } AS owner_user
      RETURN post { .*, created_at, owner_user, client_reaction: "", client_reposted: false, client_saved: false } AS new_post_data
			`,
			map[string]any{
				"client_username": clientUsername,
				"media_urls":      mediaUrls,
				"type":            postType,
				"description":     description,
				"at":              at,
			},
		)
		if err != nil {
			return nil, err
		}

		if res.Record() == nil {
			return nil, nil
		}

		maps.Copy(resMap, res.Record().AsMap())

		newPostId := resMap["new_post_data"].(map[string]any)["id"]

		mentionsExcClient := slices.DeleteFunc(mentions, func(uname string) bool {
			return uname == clientUsername
		})

		// handle mentions
		if len(mentionsExcClient) > 0 {
			res, err = tx.Run(
				ctx,
				`
				MATCH (mentionUser:User WHERE mentionUser.username IN $mentionsExcClient), (clientUser:User{ username: $client_username })

        CREATE (mentionUser)-[:RECEIVES_NOTIFICATION]->(mentionNotif:Notification:MentionNotification{ id: randomUUID(), type: "mention_in_post", is_read: false, created_at: datetime(), details: ["in_post_id", $postId], mentioning_user: ["username", clientUser.username, "profile_pic_url", clientUser.profile_pic_url] })

        WITH mentionNotif, toString(mentionNotif.created_at) AS created_at, mentionUser.username AS receiver_username
        RETURN collect(mentionNotif { .*, created_at, receiver_username }) AS mention_notifs
				`,
				map[string]any{
					"mentionsExcClient": mentionsExcClient,
					"postId":            newPostId,
					"client_username":   clientUsername,
				},
			)
			if err != nil {
				return nil, err
			}

			if res.Record() == nil {
				return nil, nil
			}

			maps.Copy(resMap, res.Record().AsMap())
		}

		_, err = tx.Run(
			ctx,
			`
			MATCH (post:Post{ id: $postId })

			UNWIND $hashtags AS hashtagName
			MERGE (ht:Hashtag{name: hashtagName})
			CREATE (post)-[:INCLUDES_HASHTAG]->(ht)
			`,
			map[string]any{
				"hashtags": hashtags,
				"postId":   newPostId,
			},
		)
		if err != nil {
			return nil, err
		}

		return resMap, nil
	})
	if err != nil {
		log.Println("postModel.go: New:", err)
		return resData, fiber.ErrInternalServerError
	}

	helpers.AnyToStruct(res, &resData)

	return resData, nil
}

func FindOne(ctx context.Context, clientUsername, postId string) (any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (post:Post{ id: $post_id })<-[:CREATES_POST]-(ownerUser:User), (clientUser:User{ username: $client_username })

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
		RETURN post { .*, owner_user, created_at, client_reaction, client_saved, client_reposted } AS found_post
    `,
		map[string]any{
			"post_id":         postId,
			"client_username": clientUsername,
		},
	)
	if err != nil {
		log.Println("postModel.go: FindOne:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	foundPost, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "found_post")

	return foundPost, nil
}

type ReactToResT struct {
	LatestReactionsCount int            `json:"latest_reactions_count"`
	ReactionNotif        map[string]any `json:"reaction_notif"`
}

func ReactTo(ctx context.Context, clientUsername, postId, reaction string) (ReactToResT, error) {
	var resData ReactToResT

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 3)

		var (
			res neo4j.ResultWithContext
			err error
			at  = time.Now().UTC()
		)

		res, err = tx.Run(
			ctx,
			`
      MATCH (clientUser:User{ username: $client_username }), (post:Post{ id: $post_id })<-[:CREATES_POST]-(postOwner)

      MERGE (clientUser)-[crxn:REACTS_TO_POST]->(post)
      ON CREATE
        SET crxn.reaction = $reaction,
          crxn.at = $at,
          post.reactions_count = post.reactions_count + 1

      RETURN post.reactions_count AS latest_reactions_count, postOwner.username AS post_owner_username
      `,
			map[string]any{
				"client_username": clientUsername,
				"post_id":         postId,
				"reaction":        reaction,
				"at":              at,
			},
		)
		if err != nil {
			return nil, err
		}

		if res.Record() == nil {
			return nil, nil
		}

		maps.Copy(resMap, res.Record().AsMap())

		postOwnerUsername := resMap["post_owner_username"]

		// handle mentions
		if postOwnerUsername != clientUsername {
			res, err = tx.Run(
				ctx,
				`
          MATCH (clientUser:User{ username: $client_username }), (post:Post{ id: $post_id })<-[:CREATES_POST]-(postOwner)
  
          CREATE (postOwner)-[:RECEIVES_NOTIFICATION]->(reactionNotif:Notification:ReactionNotification{ id: randomUUID(), type: "reaction_to_post", is_read: false, created_at: datetime(), details: ["reaction", $reaction, "to_post_id", $post_id], reactor_user: ["username", clientUser.username, "profile_pic_url", clientUser.profile_pic_url] })
  
          WITH reactionNotif, toString(reactionNotif.created_at) AS created_at, postOwner.username AS receiver_username
          RETURN reactionNotif { .*, created_at, receiver_username } AS reaction_notif
          `,
				map[string]any{
					"post_id":         postId,
					"client_username": clientUsername,
					"reaction":        reaction,
				},
			)
			if err != nil {
				return nil, err
			}

			if res.Record() == nil {
				return resMap, nil
			}

			maps.Copy(resMap, res.Record().AsMap())
		}

		return resMap, nil
	})
	if err != nil {
		log.Println("postModel.go: ReactTo:", err)
		return resData, fiber.ErrInternalServerError
	}

	helpers.AnyToStruct(res, &resData)

	return resData, nil
}
