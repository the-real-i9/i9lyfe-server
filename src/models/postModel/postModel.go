package postModel

import (
	"context"
	"i9lyfe/src/helpers"
	"i9lyfe/src/models/db"
	"maps"
	"slices"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func New(ctx context.Context, clientUsername string, mediaUrls []string, postType, description string, at time.Time) (map[string]any, error) {
	res, err := db.Query(
		ctx,
		`
			MATCH (clientUser:User{ username: $client_username })

      CREATE (clientUser)-[:CREATES_POST]->(post:Post{ id: randomUUID(), type: $type, media_urls: $media_urls, description: $description, created_at: $at })
			SET clientUser.posts_count = coalesce(clientUser.posts_count, 0) + 1

      WITH post, toString(post.created_at) AS created_at, clientUser { .username, .profile_pic_url } AS owner_user
      RETURN post { .*, created_at, owner_user } AS new_post_data
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
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	npd, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "new_post_data")

	return npd, nil
}

func NewPostExtras(ctx context.Context, clientUsername, newPostId string, mentions, hashtags []string, at time.Time) error {
	_, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		var (
			err error
		)

		if len(mentions) > 0 {

			_, err = tx.Run(
				ctx,
				`
				MATCH (mentionUser:User WHERE mentionUser.username IN $mentions), (post:Post{ id: $postId })
        MERGE (post)-[:MENTIONS_USER]->(mentionUser)
				`,
				map[string]any{
					"mentions": mentions,
					"postId":   newPostId,
				},
			)
			if err != nil {
				return nil, err
			}
		}

		_, err = tx.Run(
			ctx,
			`
			MATCH (post:Post{ id: $postId })

			UNWIND $hashtags AS hashtagName
			MERGE (ht:Hashtag{ name: hashtagName })
			MERGE (post)-[:INCLUDES_HASHTAG]->(ht)
			`,
			map[string]any{
				"hashtags": hashtags,
				"postId":   newPostId,
			},
		)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func Get(ctx context.Context, clientUsername, postId string) (map[string]any, error) {
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
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	foundPost, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "found_post")

	return foundPost, nil
}

func Delete(ctx context.Context, clientUsername, postId string) error {
	_, err := db.Query(
		ctx,
		`
		MATCH (clientUser:User{ username: $client_username })-[:CREATES_POST]->(post:Post{ id: $post_id })
		SET clientUser.posts_count = CASE WHEN clientUser.posts_count > 0 THEN clientUser.posts_count - 1 ELSE 0 END

		DETACH DELETE post
		`,
		map[string]any{
			"post_id":         postId,
			"client_username": clientUsername,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return fiber.ErrInternalServerError
	}

	return nil
}

type ReactToResT struct {
	LatestReactionsCount any            `json:"latest_reactions_count"`
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
				SET post.reactions_count = post.reactions_count + 1
      
			SET crxn.reaction = $reaction, crxn.at = $at

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

		if !res.Next(ctx) {
			return nil, nil
		}

		maps.Copy(resMap, res.Record().AsMap())

		postOwnerUsername := resMap["post_owner_username"].(string)

		// handle mentions
		if postOwnerUsername != clientUsername {
			res, err = tx.Run(
				ctx,
				`
          MATCH (clientUser:User{ username: $client_username }), (post:Post{ id: $post_id })<-[:CREATES_POST]-(postOwner)
  
          CREATE (postOwner)-[:RECEIVES_NOTIFICATION]->(reactionNotif:Notification:ReactionNotification{ id: randomUUID(), type: "reaction_to_post", is_read: false, created_at: $at, details: ["reaction", $reaction, "to_post_id", $post_id], reactor_user: ["username", clientUser.username, "profile_pic_url", clientUser.profile_pic_url] })
  
          WITH reactionNotif, toString(reactionNotif.created_at) AS created_at, postOwner.username AS receiver_username
          RETURN reactionNotif { .*, created_at, receiver_username } AS reaction_notif
          `,
				map[string]any{
					"post_id":         postId,
					"client_username": clientUsername,
					"reaction":        reaction,
					"at":              at,
				},
			)
			if err != nil {
				return nil, err
			}

			if !res.Next(ctx) {
				return resMap, nil
			}

			maps.Copy(resMap, res.Record().AsMap())
		}

		return resMap, nil
	})
	if err != nil {
		helpers.LogError(err)
		return resData, fiber.ErrInternalServerError
	}

	helpers.ToStruct(res, &resData)

	return resData, nil
}

func GetReactors(ctx context.Context, clientUsername, postId string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (:Post{ id: $post_id })<-[rxn:REACTS_TO_POST]-(reactor:User)
		WHERE rxn.at < $offset
		OPTIONAL MATCH (reactor)<-[fur:FOLLOWS_USER]-(:User{ username: $client_username })
		WITH reactor, 
			rxn, 
			CASE fur 
				WHEN IS NULL THEN false
				ELSE true 
			END AS client_follows
		ORDER BY rxn.at DESC
		LIMIT $limit
		RETURN collect(reactor { .username, .profile_pic_url, reaction: rxn.reaction }) AS reactors
		`,
		map[string]any{
			"post_id":         postId,
			"client_username": clientUsername,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	reactors, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "reactors")

	return reactors, nil
}

func GetReactorsWithReaction(ctx context.Context, clientUsername, postId, reaction string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (post:Post{ id: $post_id })<-[rxn:REACTS_TO_POST { reaction: toString($reaction) }]-(reactor:User)
		WHERE rxn.at < $offset
		OPTIONAL MATCH (reactor)<-[fur:FOLLOWS_USER]-(:User{ username: $client_username })
		WITH reactor, 
			rxn, 
			CASE fur 
				WHEN IS NULL THEN false
				ELSE true 
			END AS client_follows
		ORDER BY rxn.at DESC
		LIMIT $limit
		RETURN collect(reactor { .username, .profile_pic_url, reaction: rxn.reaction }) AS reactors_wrxn
		`,
		map[string]any{
			"post_id":         postId,
			"client_username": clientUsername,
			"reaction":        reaction,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	reactors, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "reactors_wrxn")

	return reactors, nil
}

func UndoReaction(ctx context.Context, clientUsername, postId string) (any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (:User{ username: $client_username })-[crxn:REACTS_TO_POST]->(post:Post{ id: $post_id })
		DELETE crxn

		WITH post
		SET post.reactions_count = CASE WHEN post.reactions_count > 0 THEN post.reactions_count - 1 ELSE 0 END

		RETURN post.reactions_count AS latest_reactions_count
		`,
		map[string]any{
			"post_id":         postId,
			"client_username": clientUsername,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	lrc, _ := res.Records[0].Get("latest_reactions_count")

	return lrc, nil
}

type CommentOnResT struct {
	NewCommentData      map[string]any   `json:"new_comment_data"`
	MentionNotifs       []map[string]any `json:"mention_notifs"`
	CommentNotif        map[string]any   `json:"comment_notif"`
	LatestCommentsCount any              `json:"latest_comments_count"`
}

func CommentOn(ctx context.Context, clientUsername, postId, commentText, attachmentUrl string, mentions []string) (CommentOnResT, error) {
	var resData CommentOnResT

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 5)

		var (
			res neo4j.ResultWithContext
			err error
			at  = time.Now().UTC()
		)

		res, err = tx.Run(
			ctx,
			`
			MATCH (clientUser:User{ username: $client_username }), (post:Post{ id: $post_id })<-[:CREATES_POST]-(postOwner)
			CREATE (clientUser)-[:WRITES_COMMENT]->(comment:Comment{ id: randomUUID(), comment_text: $comment_text, attachment_url: $attachment_url, reactions_count: 0, comments_count: 0, created_at: $at })-[:COMMENT_ON_POST]->(post)

			WITH post, comment, toString(comment.created_at) AS created_at, clientUser { .username, .profile_pic_url } AS owner_user, postOwner.username AS post_owner_username
			
			SET post.comments_count = post.comments_count + 1

			RETURN post.comments_count AS latest_comments_count,
				comment { .*, created_at, owner_user, client_reaction: "" } AS new_comment_data,
				post_owner_username
			`,
			map[string]any{
				"client_username": clientUsername,
				"post_id":         postId,
				"comment_text":    commentText,
				"attachment_url":  attachmentUrl,
				"at":              at,
			},
		)
		if err != nil {
			return nil, err
		}

		if !res.Next(ctx) {
			return nil, nil
		}

		maps.Copy(resMap, res.Record().AsMap())

		newCommentId := resMap["new_comment_data"].(map[string]any)["id"]

		if len(mentions) > 0 {

			_, err = tx.Run(
				ctx,
				`
				MATCH (mentionUser:User WHERE mentionUser.username IN $mentions), (comment:Comment{ id: $commentId })
        CREATE (comment)-[:MENTIONS_USER]->(mentionUser)
				`,
				map[string]any{
					"mentions":  mentions,
					"commentId": newCommentId,
				},
			)
			if err != nil {
				return nil, err
			}

			mentionsExcClient := slices.DeleteFunc(mentions, func(uname string) bool {
				return uname == clientUsername
			})

			// handle mentions
			if len(mentionsExcClient) > 0 {
				res, err = tx.Run(
					ctx,
					`
					MATCH (mentionUser:User WHERE mentionUser.username IN $mentionsExcClient), (clientUser:User{ username: $client_username })

					CREATE (mentionUser)-[:RECEIVES_NOTIFICATION]->(mentionNotif:Notification:MentionNotification{ id: randomUUID(), type: "mention_in_comment", is_read: false, created_at: $at, details: ["in_comment_id", $commentId], mentioning_user: ["username", clientUser.username, "profile_pic_url", clientUser.profile_pic_url] })

					WITH mentionNotif, toString(mentionNotif.created_at) AS created_at, mentionUser.username AS receiver_username
					RETURN collect(mentionNotif { .*, receiver_username, created_at }) AS mention_notifs
					`,
					map[string]any{
						"mentionsExcClient": mentionsExcClient,
						"commentId":         newCommentId,
						"client_username":   clientUsername,
						"at":                at,
					},
				)
				if err != nil {
					return nil, err
				}

				if !res.Next(ctx) {
					return nil, nil
				}

				maps.Copy(resMap, res.Record().AsMap())
			}
		}

		postOwnerUsername := resMap["post_owner_username"].(string)

		if postOwnerUsername != clientUsername {
			res, err = tx.Run(
				ctx,
				`
				MATCH (clientUser:User{ username: $client_username }), (post:Post{ id: $post_id })<-[:CREATES_POST]-(postOwner)
				
				CREATE (postOwner)-[:RECEIVES_NOTIFICATION]->(commentNotif:Notification:CommentNotification{ id: randomUUID(), type: "comment_on_post", is_read: false, created_at: $at, details: ["on_post_id", $post_id, "comment_id", $commentId, "comment_text", $comment_text, "attachment_url", $attachment_url], commenter_user: ["username", clientUser.username, "profile_pic_url", clientUser.profile_pic_url] })

				WITH commentNotif, 
					toString(commentNotif.created_at) AS created_at, 
					postOwner.username AS receiver_username
				RETURN commentNotif { .*, created_at, receiver_username } AS comment_notif
				`,
				map[string]any{
					"client_username": clientUsername,
					"post_id":         postId,
					"commentId":       newCommentId,
					"comment_text":    commentText,
					"attachment_url":  attachmentUrl,
					"at":              at,
				},
			)
			if err != nil {
				return nil, err
			}

			if !res.Next(ctx) {
				return nil, nil
			}

			maps.Copy(resMap, res.Record().AsMap())
		}

		return resMap, nil
	})
	if err != nil {
		helpers.LogError(err)
		return resData, fiber.ErrInternalServerError
	}

	helpers.ToStruct(res, &resData)

	return resData, nil
}

func GetComments(ctx context.Context, clientUsername, postId string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (post:Post{ id: $post_id })<-[:COMMENT_ON_POST]-(comment:Comment WHERE comment.created_at < $offset)<-[:WRITES_COMMENT]-(ownerUser:User)

		OPTIONAL MATCH (comment)<-[crxn:REACTS_TO_COMMENT]-(:User{ username: $client_username })

		WITH comment, 
			toString(comment.created_at) AS created_at, 
			ownerUser { .username, .profile_pic_url } AS owner_user,
			CASE crxn 
				WHEN IS NULL THEN "" 
				ELSE crxn.reaction 
			END AS client_reaction
		ORDER BY comment.created_at DESC
		LIMIT $limit
		RETURN collect(comment {.*, owner_user, created_at, client_reaction }) AS comments
		`,
		map[string]any{
			"post_id":         postId,
			"client_username": clientUsername,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	comments, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "comments")

	return comments, nil
}

func RemoveComment(ctx context.Context, clientUsername, postId, commentId string) (any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (clientUser:User{ username: $client_username })-[:WRITES_COMMENT]->(comment:Comment{ id: $comment_id })-[:COMMENT_ON_POST]->(post:Post{ id: $post_id })
		DETACH DELETE comment

		WITH post
		SET post.comments_count = CASE WHEN post.comments_count > 0 THEN post.comments_count - 1 ELSE 0 END

		RETURN post.comments_count AS latest_comments_count
		`,
		map[string]any{
			"post_id":         postId,
			"comment_id":      commentId,
			"client_username": clientUsername,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	lcc, _ := res.Records[0].Get("latest_comments_count")

	return lcc, nil
}

type RepostResT struct {
	LatestRepostsCount any            `json:"latest_reposts_count"`
	RepostNotif        map[string]any `json:"repost_notif"`
}

func Repost(ctx context.Context, clientUsername, postId string) (RepostResT, error) {
	var resData RepostResT

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

			MERGE (clientUser)-[crep:REPOSTS_POST]->(post)
			ON CREATE
				SET post.reposts_count = post.reposts_count + 1,
					crep.at = $at

			RETURN post.reposts_count AS latest_reposts_count, postOwner.username AS post_owner_username
			`,
			map[string]any{
				"client_username": clientUsername,
				"post_id":         postId,
				"at":              at,
			},
		)
		if err != nil {
			return nil, err
		}

		if !res.Next(ctx) {
			return nil, nil
		}

		maps.Copy(resMap, res.Record().AsMap())

		postOwnerUsername := resMap["post_owner_username"].(string)

		// handle mentions
		if postOwnerUsername != clientUsername {
			res, err = tx.Run(
				ctx,
				`
        MATCH (clientUser:User{ username: $client_username }), (post:Post{ id: $post_id })<-[:CREATES_POST]-(postOwner)

        CREATE (postOwner)-[:RECEIVES_NOTIFICATION]->(repostNotif:Notification:RepostNotification{ id: randomUUID(), type: "repost", is_read: false, created_at: $at, details: ["post_id", $post_id], reposter_user: ["username", clientUser.username, "profile_pic_url", clientUser.profile_pic_url] })
        WITH repostNotif, toString(repostNotif.created_at) AS created_at, postOwner.username AS receiver_username
        RETURN repostNotif { .*, created_at, receiver_username } AS repost_notif
        `,
				map[string]any{
					"post_id":         postId,
					"client_username": clientUsername,
					"at":              at,
				},
			)
			if err != nil {
				return nil, err
			}

			if !res.Next(ctx) {
				return resMap, nil
			}

			maps.Copy(resMap, res.Record().AsMap())
		}

		return resMap, nil
	})
	if err != nil {
		helpers.LogError(err)
		return resData, fiber.ErrInternalServerError
	}

	helpers.ToStruct(res, &resData)

	return resData, nil
}

func Save(ctx context.Context, clientUsername, postId string) (any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (clientUser:User{ username: $client_username }), (post:Post{ id: $post_id })
		MERGE (clientUser)-[:SAVES_POST]->(post)
		ON CREATE
			SET post.saves_count = post.saves_count + 1

		RETURN post.saves_count AS latest_saves_count
		`,
		map[string]any{
			"post_id":         postId,
			"client_username": clientUsername,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	lsc, _ := res.Records[0].Get("latest_saves_count")

	return lsc, nil
}

func UndoSave(ctx context.Context, clientUsername, postId string) (any, error) {
	res, err := db.Query(
		ctx,
		`
      MATCH (:User{ username: $client_username })-[csave:SAVES_POST]->(post:Post{ id: $post_id })
      DELETE csave

			WITH post
      SET post.saves_count = CASE WHEN post.saves_count > 0 THEN post.saves_count - 1 ELSE 0 END

      RETURN post.saves_count AS latest_saves_count
      `,
		map[string]any{
			"post_id":         postId,
			"client_username": clientUsername,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	lsc, _ := res.Records[0].Get("latest_saves_count")

	return lsc, nil
}
