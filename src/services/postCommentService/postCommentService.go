package postCommentService

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/helpers"
	comment "i9lyfe/src/models/commentModel"
	post "i9lyfe/src/models/postModel"
	"i9lyfe/src/services/cloudStorageService"
	"i9lyfe/src/services/eventStreamService"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/realtimeService"
	"i9lyfe/src/services/utilServices"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v2"
)

func CreateNewPost(ctx context.Context, clientUsername string, mediaDataList [][]byte, postType, description string) (map[string]any, error) {

	mediaUrls := make([]string, len(mediaDataList))

	for i, mediaData := range mediaDataList {
		mediaMIME := mimetype.Detect(mediaData)
		mediaType, mediaExt := mediaMIME.String(), mediaMIME.Extension()

		if ((postType == "reel" || postType == "video") && !strings.HasPrefix(mediaType, "video")) || (postType == "photo" && !strings.HasPrefix(mediaType, "image")) {
			return nil, fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid media type %s, for the post type %s", mediaType, postType))
		}

		murl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("post_medias/user-%s/media-%d%s", clientUsername, time.Now().UnixNano(), mediaExt), mediaData)
		if err != nil {
			return nil, err
		}

		mediaUrls[i] = murl
	}

	hashtags := utilServices.ExtractHashtags(description)
	mentions := utilServices.ExtractMentions(description)

	at := time.Now().UTC()

	newPost, err := post.New(ctx, clientUsername, mediaUrls, postType, description, at)
	if err != nil {
		return nil, err
	}

	eventStreamService.QueueNewPost(ctx, eventTypes.NewPostEvent{
		ClientUsername: clientUsername,
		PostId:         newPost["id"].(string),
		PostData:       newPost,
		Hashtags:       hashtags,
		Mentions:       mentions,
		At:             at,
	})

	/* go func() {
		if len(res.MentionNotifs) > 0 {
			for _, mn := range res.MentionNotifs {
				mn := mn
				receiverUsername := mn["receiver_username"].(string)

				delete(mn, "receiver_username")

				realtimeService.SendEventMsg(receiverUsername, appTypes.ServerEventMsg{
					Event: "new notification",
					Data:  mn,
				})
			}
		}

		contentRecommendationService.FanOutPost(res.NewPostData["id"].(string))
	}() */

	return newPost, nil
}

func GetPost(ctx context.Context, clientUsername, postId string) (any, error) {
	thePost, err := post.Get(ctx, clientUsername, postId)
	if err != nil {
		return nil, err
	}

	return thePost, nil
}

func DeletePost(ctx context.Context, clientUsername, postId string) (any, error) {
	if err := post.Delete(ctx, clientUsername, postId); err != nil {
		return nil, err
	}

	return true, nil
}

func ReactToPost(ctx context.Context, clientUsername, postId, reaction string) (any, error) {
	res, err := post.ReactTo(ctx, clientUsername, postId, reaction)
	if err != nil {
		return nil, err
	}

	go func() {
		if res.ReactionNotif != nil {
			rn := res.ReactionNotif
			receiverUsername := rn["receiver_username"].(string)

			delete(rn, "receiver_username")

			realtimeService.SendEventMsg(receiverUsername, appTypes.ServerEventMsg{
				Event: "new notification",
				Data:  rn,
			})

		}

		realtimeService.PublishPostMetric(ctx, map[string]any{
			"post_id":                postId,
			"latest_reactions_count": res.LatestReactionsCount,
		})
	}()

	return true, nil
}

func GetReactorsToPost(ctx context.Context, clientUsername, postId string, limit int, offset int64) (any, error) {
	reactors, err := post.GetReactors(ctx, clientUsername, postId, limit, helpers.OffsetTime(offset))
	if err != nil {
		return nil, err
	}

	return reactors, nil
}

func GetReactorsWithReactionToPost(ctx context.Context, clientUsername, postId, reaction string, limit int, offset int64) (any, error) {
	reactors, err := post.GetReactorsWithReaction(ctx, clientUsername, postId, reaction, limit, helpers.OffsetTime(offset))
	if err != nil {
		return nil, err
	}

	return reactors, nil
}

func UndoReactionToPost(ctx context.Context, clientUsername, postId string) (any, error) {
	latestReactionsCount, err := post.UndoReaction(ctx, clientUsername, postId)
	if err != nil {
		return nil, err
	}

	go realtimeService.PublishPostMetric(ctx, map[string]any{
		"post_id":                postId,
		"latest_reactions_count": latestReactionsCount,
	})

	return true, nil
}

func CommentOnPost(ctx context.Context, clientUsername, postId, commentText string, attachmentData []byte) (map[string]any, error) {

	var (
		attachmentUrl string
		err           error
	)

	if attachmentData != nil {
		mediaMIME := mimetype.Detect(attachmentData)
		mediaType, mediaExt := mediaMIME.String(), mediaMIME.Extension()

		if !strings.HasPrefix(mediaType, "image") {
			return nil, fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid media type %s, for attachment_data, expected image/*", mediaType))
		}

		attachmentUrl, err = cloudStorageService.Upload(ctx, fmt.Sprintf("comment_on_post_attachments/user-%s/att-%d%s", clientUsername, time.Now().UnixNano(), mediaExt), attachmentData)
		if err != nil {
			return nil, err
		}
	}

	mentions := utilServices.ExtractMentions(commentText)

	res, err := post.CommentOn(ctx, clientUsername, postId, commentText, attachmentUrl, mentions)
	if err != nil {
		return nil, err
	}

	go func() {
		if len(res.MentionNotifs) > 0 {
			for _, mn := range res.MentionNotifs {
				mn := mn
				receiverUsername := mn["receiver_username"].(string)

				delete(mn, "receiver_username")

				realtimeService.SendEventMsg(receiverUsername, appTypes.ServerEventMsg{
					Event: "new notification",
					Data:  mn,
				})
			}
		}

		if res.CommentNotif != nil {
			cn := res.CommentNotif
			receiverUsername := cn["receiver_username"].(string)

			delete(cn, "receiver_username")

			realtimeService.SendEventMsg(receiverUsername, appTypes.ServerEventMsg{
				Event: "new notification",
				Data:  cn,
			})

		}

		realtimeService.PublishPostMetric(ctx, map[string]any{
			"post_id":               postId,
			"latest_comments_count": res.LatestCommentsCount,
		})

	}()

	return res.NewCommentData, nil
}

func GetCommentsOnPost(ctx context.Context, clientUsername, postId string, limit int, offset int64) (any, error) {
	comments, err := post.GetComments(ctx, clientUsername, postId, limit, helpers.OffsetTime(offset))
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func GetComment(ctx context.Context, clientUsername, commentId string) (any, error) {
	theComment, err := comment.Get(ctx, clientUsername, commentId)
	if err != nil {
		return nil, err
	}

	return theComment, nil
}

func RemoveCommentOnPost(ctx context.Context, clientUsername, postId, commentId string) (any, error) {
	latestCommentsCount, err := post.RemoveComment(ctx, clientUsername, postId, commentId)
	if err != nil {
		return nil, err
	}

	go realtimeService.PublishPostMetric(ctx, map[string]any{
		"post_id":               postId,
		"latest_comments_count": latestCommentsCount,
	})

	return true, nil
}

func ReactToComment(ctx context.Context, clientUsername, commentId, reaction string) (any, error) {
	res, err := comment.ReactTo(ctx, clientUsername, commentId, reaction)
	if err != nil {
		return nil, err
	}

	go func() {
		if res.ReactionNotif != nil {
			rn := res.ReactionNotif
			receiverUsername := rn["receiver_username"].(string)

			delete(rn, "receiver_username")

			realtimeService.SendEventMsg(receiverUsername, appTypes.ServerEventMsg{
				Event: "new notification",
				Data:  rn,
			})

		}

		realtimeService.PublishCommentMetric(ctx, map[string]any{
			"comment_id":             commentId,
			"latest_reactions_count": res.LatestReactionsCount,
		})
	}()

	return true, nil
}

func GetReactorsToComment(ctx context.Context, clientUsername, commentId string, limit int, offset int64) (any, error) {
	reactors, err := comment.GetReactors(ctx, clientUsername, commentId, limit, helpers.OffsetTime(offset))
	if err != nil {
		return nil, err
	}

	return reactors, nil
}

func GetReactorsWithReactionToComment(ctx context.Context, clientUsername, commentId, reaction string, limit int, offset int64) (any, error) {
	reactors, err := comment.GetReactorsWithReaction(ctx, clientUsername, commentId, reaction, limit, helpers.OffsetTime(offset))
	if err != nil {
		return nil, err
	}

	return reactors, nil
}

func UndoReactionToComment(ctx context.Context, clientUsername, commentId string) (any, error) {
	latestReactionsCount, err := comment.UndoReaction(ctx, clientUsername, commentId)
	if err != nil {
		return nil, err
	}

	go realtimeService.PublishCommentMetric(ctx, map[string]any{
		"comment_id":             commentId,
		"latest_reactions_count": latestReactionsCount,
	})

	return true, nil
}

func CommentOnComment(ctx context.Context, clientUsername, commentId, commentText string, attachmentData []byte) (map[string]any, error) {

	var (
		attachmentUrl string
		err           error
	)

	if attachmentData != nil {
		mediaMIME := mimetype.Detect(attachmentData)
		mediaType := mediaMIME.String()
		mediaExt := mediaMIME.Extension()

		if !strings.HasPrefix(mediaType, "image") {
			return nil, fiber.NewError(400, fmt.Sprintf("invalid media type %s, for attachment_data, expected image/*", mediaType))
		}

		attachmentUrl, err = cloudStorageService.Upload(ctx, fmt.Sprintf("comment_on_comment_attachments/user-%s/att-%d%s", clientUsername, time.Now().UnixNano(), mediaExt), attachmentData)
		if err != nil {
			return nil, err
		}
	}

	mentions := utilServices.ExtractMentions(commentText)

	res, err := comment.CommentOn(ctx, clientUsername, commentId, commentText, attachmentUrl, mentions)
	if err != nil {
		return nil, err
	}

	go func() {
		if len(res.MentionNotifs) > 0 {
			for _, mn := range res.MentionNotifs {
				mn := mn

				receiverUsername := mn["receiver_username"].(string)

				delete(mn, "receiver_username")

				realtimeService.SendEventMsg(receiverUsername, appTypes.ServerEventMsg{
					Event: "new notification",
					Data:  mn,
				})
			}
		}

		if res.CommentNotif != nil {
			cn := res.CommentNotif

			receiverUsername := cn["receiver_username"].(string)

			delete(cn, "receiver_username")

			realtimeService.SendEventMsg(receiverUsername, appTypes.ServerEventMsg{
				Event: "new notification",
				Data:  cn,
			})
		}

		realtimeService.PublishCommentMetric(ctx, map[string]any{
			"comment_id":            commentId,
			"latest_comments_count": res.LatestCommentsCount,
		})
	}()

	return res.NewCommentData, nil
}

func GetCommentsOnComment(ctx context.Context, clientUsername, commentId string, limit int, offset int64) (any, error) {
	comments, err := comment.GetComments(ctx, clientUsername, commentId, limit, helpers.OffsetTime(offset))
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func RemoveCommentOnComment(ctx context.Context, clientUsername, parentCommentId, childCommentId string) (any, error) {
	latestCommentsCount, err := comment.RemoveChildComment(ctx, clientUsername, parentCommentId, childCommentId)
	if err != nil {
		return nil, err
	}

	go realtimeService.PublishCommentMetric(ctx, map[string]any{
		"comment_id":            parentCommentId,
		"latest_comments_count": latestCommentsCount,
	})

	return true, nil
}

func RepostPost(ctx context.Context, clientUsername, postId string) (any, error) {
	res, err := post.Repost(ctx, clientUsername, postId)
	if err != nil {
		return nil, err
	}

	go func() {
		if res.RepostNotif != nil {
			rn := res.RepostNotif
			receiverUsername := rn["receiver_username"].(string)

			delete(rn, "receiver_username")

			realtimeService.SendEventMsg(receiverUsername, appTypes.ServerEventMsg{
				Event: "new notification",
				Data:  rn,
			})

		}

		realtimeService.PublishPostMetric(ctx, map[string]any{
			"post_id":              postId,
			"latest_reposts_count": res.LatestRepostsCount,
		})
	}()

	return true, nil
}

func SavePost(ctx context.Context, clientUsername, postId string) (any, error) {
	latestSavesCount, err := post.Save(ctx, clientUsername, postId)
	if err != nil {
		return nil, err
	}

	go realtimeService.PublishPostMetric(ctx, map[string]any{
		"post_id":            postId,
		"latest_saves_count": latestSavesCount,
	})

	return true, nil
}

func UndoSavePost(ctx context.Context, clientUsername, postId string) (any, error) {
	latestSavesCount, err := post.UndoSave(ctx, clientUsername, postId)
	if err != nil {
		return nil, err
	}

	go realtimeService.PublishPostMetric(ctx, map[string]any{
		"post_id":            postId,
		"latest_saves_count": latestSavesCount,
	})

	return true, nil
}
