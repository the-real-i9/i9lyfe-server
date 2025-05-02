package commentModel

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

func Get(ctx context.Context, clientUsername, commentId string) (any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (comment:Comment{ id: $comment_id })<-[:WRITES_COMMENT]-(ownerUser:User), (clientUser:User{ username: $client_username })
		OPTIONAL MATCH (clientUser)-[crxn:REACTS_TO_COMMENT]->(comment)
		WITH comment, 
			toString(comment.created_at) AS created_at, 
			ownerUser { .username, .profile_pic_url } AS owner_user,
			CASE crxn 
				WHEN IS NULL THEN "" 
				ELSE crxn.reaction 
			END AS client_reaction
		RETURN comment { .*, owner_user, created_at, client_reaction } AS found_comment
		`,
		map[string]any{
			"comment_id":      commentId,
			"client_username": clientUsername,
		},
	)
	if err != nil {
		log.Println("commentModel.go: Get:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	foundComment, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "found_comment")

	return foundComment, nil
}

type ReactToResT struct {
	LatestReactionsCount any            `json:"latest_reactions_count"`
	ReactionNotif        map[string]any `json:"reaction_notif"`
}

func ReactTo(ctx context.Context, clientUsername, commentId, reaction string) (ReactToResT, error) {
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
			MATCH (clientUser:User{ username: $client_username }), (comment:Comment{ id: $comment_id })<-[:WRITES_COMMENT]-(commentOwner)

			MERGE (clientUser)-[crxn:REACTS_TO_COMMENT]->(comment)
			ON CREATE
				SET crxn.reaction = $reaction,
					crxn.at = $at,
					comment.reactions_count = comment.reactions_count + 1

			RETURN comment.reactions_count AS latest_reactions_count, 
				commentOwner.username AS comment_owner_username
			`,
			map[string]any{
				"client_username": clientUsername,
				"comment_id":      commentId,
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

		commentOwnerUsername := resMap["comment_owner_username"].(string)

		// handle mentions
		if commentOwnerUsername != clientUsername {
			res, err = tx.Run(
				ctx,
				`
				MATCH (clientUser:User{ username: $client_username }), (comment:Comment{ id: $comment_id })<-[:WRITES_COMMENT]-(commentOwner)
				
				CREATE (commentOwner)-[:RECEIVES_NOTIFICATION]->(reactionNotif:Notification:ReactionNotification{ id: randomUUID(), type: "reaction_to_comment", is_read: false, created_at: $at, details: ["reaction", $reaction, "to_comment_id", $comment_id], reactor_user: ["username", clientUser.username, "profile_pic_url", clientUser.profile_pic_url] })

				WITH reactionNotif, toString(reactionNotif.created_at) AS created_at, commentOwner.username AS receiver_username
				RETURN reactionNotif { .*, created_at, receiver_username } AS reaction_notif
				`,
				map[string]any{
					"comment_id":      commentId,
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
		log.Println("commentModel.go: ReactTo:", err)
		return resData, fiber.ErrInternalServerError
	}

	helpers.ToStruct(res, &resData)

	return resData, nil
}

func GetReactors(ctx context.Context, clientUsername, commentId string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (:Comment{ id: $comment_id })<-[rxn:REACTS_TO_COMMENT]-(reactor:User)
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
			"comment_id":      commentId,
			"client_username": clientUsername,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		log.Println("commentModel.go: GetReactors:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	reactors, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "reactors")

	return reactors, nil
}

func GetReactorsWithReaction(ctx context.Context, clientUsername, commentId, reaction string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (:Comment{ id: $comment_id })<-[rxn:REACTS_TO_COMMENT { reaction: $reaction }]-(reactor:User)
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
			"comment_id":      commentId,
			"client_username": clientUsername,
			"reaction":        reaction,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		log.Println("commentModel.go: GetReactorsWithReaction:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	reactors, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "reactors_wrxn")

	return reactors, nil
}

func UndoReaction(ctx context.Context, clientUsername, commentId string) (any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (:User{ username: $client_username })-[crxn:REACTS_TO_COMMENT]->(comment:Comment{ id: $comment_id })
		DELETE crxn

		WITH comment
		SET comment.reactions_count = comment.reactions_count - 1

		RETURN comment.reactions_count AS latest_reactions_count
		`,
		map[string]any{
			"comment_id":      commentId,
			"client_username": clientUsername,
		},
	)
	if err != nil {
		log.Println("commentModel.go: UndoReaction:", err)
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

func CommentOn(ctx context.Context, clientUsername, commentId, commentText, attachmentUrl string, mentions []string) (CommentOnResT, error) {
	var resData CommentOnResT

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 7)

		var (
			res neo4j.ResultWithContext
			err error
			at  = time.Now().UTC()
		)

		res, err = tx.Run(
			ctx,
			`
			MATCH (clientUser:User{ username: $client_username }), (parentComment:Comment{ id: $comment_id })<-[:WRITES_COMMENT]-(parentCommentOwner)

			CREATE (clientUser)-[:WRITES_COMMENT]->(childComment:Comment{ id: randomUUID(), comment_text: $comment_text, attachment_url: $attachment_url,  reactions_count: 0, comments_count: 0, created_at: $at })-[:COMMENT_ON_COMMENT]->(parentComment)

			WITH parentComment, childComment, toString(childComment.created_at) AS created_at, clientUser { .username, .profile_pic_url } AS owner_user, parentCommentOwner.username AS parent_comment_owner_username

			SET parentComment.comments_count = parentComment.comments_count + 1

			RETURN parentComment.comments_count AS latest_comments_count,
				childComment { .*, created_at, owner_user, client_reaction: "" } AS new_comment_data,
				parent_comment_owner_username
			`,
			map[string]any{
				"client_username": clientUsername,
				"comment_id":      commentId,
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
				MATCH (mentionUser:User WHERE mentionUser.username IN $mentions), (childComment:Comment{ id: $childCommentId })
				CREATE (childComment)-[:MENTIONS_USER]->(mentionUser)
				`,
				map[string]any{
					"mentions":     mentions,
					"childComment": newCommentId,
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
					MATCH (mentionUser:User WHERE mentionUser.username IN $mentionsExcClient), (childComment:Comment{ id: $childCommentId }), (clientUser:User{ username: $client_username })
					
					CREATE (mentionUser)-[:RECEIVES_NOTIFICATION]->(mentionNotif:Notification:MentionNotification{ id: randomUUID(), type: "mention_in_comment", is_read: false, created_at: $at, details: ["in_comment_id", childComment.id], mentioning_user: ["username", clientUser.username, "profile_pic_url", clientUser.profile_pic_url] })

					WITH mentionNotif, toString(mentionNotif.created_at) AS created_at, mentionUser.username AS receiver_username
					RETURN collect(mentionNotif { .*, created_at, receiver_username }) AS mention_notifs
					`,
					map[string]any{
						"mentionsExcClient": mentionsExcClient,
						"childCommentId":    newCommentId,
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

		parentCommentOwnerUsername := resMap["parent_comment_owner_username"].(string)

		if parentCommentOwnerUsername != clientUsername {
			res, err = tx.Run(
				ctx,
				`
				MATCH (clientUser:User{ username: $client_username }), (parentComment:Comment{ id: $comment_id })<-[:WRITES_COMMENT]-(parentCommentOwner)
				
				CREATE (parentCommentOwner)-[:RECEIVES_NOTIFICATION]->(commentNotif:Notification:CommentNotification{ id: randomUUID(), type: "comment_on_comment", is_read: false, created_at: $at, details: ["on_comment_id", $comment_id, "child_comment_id", $childCommentId, "comment_text", $comment_text, "attachment_url", $attachment_url], commenter_user: ["username", clientUser.username, "profile_pic_url", clientUser.profile_pic_url] })
				
				WITH commentNotif, toString(commentNotif.created_at) AS created_at, parentCommentOwner.username AS receiver_username
				RETURN commentNotif { .*, created_at, receiver_username } AS comment_notif
				`,
				map[string]any{
					"client_username": clientUsername,
					"comment_id":      commentId,
					"childCommentId":  newCommentId,
					"comment_text":    commentText,
					"attachment_url":  attachmentUrl,
					"at":              at,
				},
			)
			if err != nil {
				return nil, err
			}

			if !res.Next(ctx) {
				log.Println("no record")
				return nil, nil
			}

			maps.Copy(resMap, res.Record().AsMap())
		}

		return resMap, nil
	})
	if err != nil {
		log.Println("commentModel.go: CommentOn:", err)
		return resData, fiber.ErrInternalServerError
	}

	helpers.ToStruct(res, &resData)

	return resData, nil
}

func GetComments(ctx context.Context, clientUsername, commentId string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (parentComment:Comment{ id: $comment_id })<-[:COMMENT_ON_COMMENT]-(childComment:Comment WHERE childComment.created_at < $offset)<-[:WRITES_COMMENT]-(ownerUser:User)

		OPTIONAL MATCH (childComment)<-[crxn:REACTS_TO_COMMENT]-(:User{ username: $client_username })
		WITH childComment, 
			toString(childComment.created_at) AS created_at, 
			ownerUser { .username, .profile_pic_url } AS owner_user,
			CASE crxn 
				WHEN IS NULL THEN "" 
				ELSE crxn.reaction 
			END AS client_reaction
		ORDER BY childComment.created_at DESC, childComment.reactions_count DESC, childComment.comments_count DESC
		LIMIT $limit
		RETURN collect(childComment {.*, created_at, owner_user, client_reaction }) AS comments
		`,
		map[string]any{
			"comment_id":      commentId,
			"client_username": clientUsername,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		log.Println("commentModel.go: GetComments:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	comments, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "comments")

	return comments, nil
}

func RemoveChildComment(ctx context.Context, clientUsername, parentCommentId, childCommentId string) (any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (clientUser:User{ username: $client_username })-[:WRITES_COMMENT]->(childComment:Comment{ id: $child_comment_id })-[:COMMENT_ON_COMMENT]->(parentComment:Comment{ id: $parent_comment_id })
		DETACH DELETE childComment

		WITH parentComment
		SET parentComment.comments_count = parentComment.comments_count - 1

		RETURN parentComment.comments_count AS latest_comments_count
		`,
		map[string]any{
			"parent_comment_id": parentCommentId,
			"child_comment_id":  childCommentId,
			"client_username":   clientUsername,
		},
	)
	if err != nil {
		log.Println("commentModel.go: RemoveComment:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	lcc, _ := res.Records[0].Get("latest_comments_count")

	return lcc, nil
}
