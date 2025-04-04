package postCommentService

import (
	"context"
	"fmt"
	"i9lyfe/src/appGlobals"
	comment "i9lyfe/src/models/commentModel"
	post "i9lyfe/src/models/postModel"
	"i9lyfe/src/services/cloudStorageService"
	"i9lyfe/src/services/messageBrokerService"
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
		mime := mimetype.Detect(mediaData)
		fileType := mime.String()
		fileExt := mime.Extension()

		if ((postType == "reel" || postType == "video") && !strings.HasPrefix(fileType, "video")) || (postType == "photo" && !strings.HasPrefix(fileType, "image")) {
			return nil, fiber.NewError(400, fmt.Sprintf("invalid file type %s, for the post type %s", fileType, postType))
		}

		murl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("post_medias/user-%s", clientUsername), mediaData, fileExt)
		if err != nil {
			return nil, err
		}

		mediaUrls[i] = murl
	}

	hashtags := utilServices.ExtractHashtags(description)
	mentions := utilServices.ExtractMentions(description)

	res, err := post.New(ctx, clientUsername, mediaUrls, postType, description, mentions, hashtags)
	if err != nil {
		return nil, err
	}

	realtimeService.BroadcastNewPost(res.NewPostData["id"].(string), clientUsername)

	go func(mentionNotifs []map[string]any) {
		for _, mn := range mentionNotifs {
			mn := mn
			receiverUsername := mn["receiver_username"].(string)

			delete(mn, "receiver_username")

			// send notification with message broker
			messageBrokerService.Send(fmt.Sprintf("user-%s-alerts", receiverUsername), messageBrokerService.Message{
				Event: "new notification",
				Data:  mn,
			})
		}
	}(res.MentionNotifs)

	return res.NewPostData, nil
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

	return appGlobals.OprSucc, nil
}

func ReactToPost(ctx context.Context, clientUsername, postId string, reaction rune) (any, error) {
	res, err := post.ReactTo(ctx, clientUsername, postId, reaction)
	if err != nil {
		return nil, err
	}

	go func(reactionNotif map[string]any) {
		if rn := reactionNotif; rn != nil {
			receiverUsername := rn["receiver_username"].(string)

			delete(rn, "receiver_username")

			// send notification with message broker
			messageBrokerService.Send(fmt.Sprintf("user-%s-alerts", receiverUsername), messageBrokerService.Message{
				Event: "new notification",
				Data:  rn,
			})
		}
	}(res.ReactionNotif)

	go realtimeService.SendPostUpdate(postId, map[string]any{
		"post_id":                postId,
		"latest_reactions_count": res.LatestReactionsCount,
	})

	return appGlobals.OprSucc, nil
}

func GetReactorsToPost(ctx context.Context, clientUsername, postId string, limit int, offset int64) (any, error) {
	reactors, err := post.GetReactors(ctx, clientUsername, postId, limit, time.UnixMilli(offset).UTC())
	if err != nil {
		return nil, err
	}

	return reactors, nil
}

func GetReactorsWithReactionToPost(ctx context.Context, clientUsername, postId string, reaction rune, limit int, offset int64) (any, error) {
	reactors, err := post.GetReactorsWithReaction(ctx, clientUsername, postId, reaction, limit, time.UnixMilli(offset).UTC())
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

	go realtimeService.SendPostUpdate(postId, map[string]any{
		"post_id":                postId,
		"latest_reactions_count": latestReactionsCount,
	})

	return appGlobals.OprSucc, nil
}

func CommentOnPost(ctx context.Context, clientUsername, postId, commentText string, attachmentData []byte) (map[string]any, error) {

	mime := mimetype.Detect(attachmentData)
	fileType := mime.String()
	fileExt := mime.Extension()

	if !strings.HasPrefix(fileType, "image") {
		return nil, fiber.NewError(400, fmt.Sprintf("invalid file type %s, for attachment_data, expected image/*", fileType))
	}

	attachmentUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("comment_on_post_attachments/user-%s", clientUsername), attachmentData, fileExt)
	if err != nil {
		return nil, err
	}

	mentions := utilServices.ExtractMentions(commentText)

	res, err := post.CommentOn(ctx, clientUsername, postId, commentText, attachmentUrl, mentions)
	if err != nil {
		return nil, err
	}

	go func(mentionNotifs []map[string]any) {
		for _, mn := range mentionNotifs {
			mn := mn
			receiverUsername := mn["receiver_username"].(string)

			delete(mn, "receiver_username")

			// send notification with message broker
			messageBrokerService.Send(fmt.Sprintf("user-%s-alerts", receiverUsername), messageBrokerService.Message{
				Event: "new notification",
				Data:  mn,
			})
		}
	}(res.MentionNotifs)

	go func(commentNotif map[string]any) {
		if cn := commentNotif; cn != nil {
			receiverUsername := cn["receiver_username"].(string)

			delete(cn, "receiver_username")

			// send notification with message broker
			messageBrokerService.Send(fmt.Sprintf("user-%s-alerts", receiverUsername), messageBrokerService.Message{
				Event: "new notification",
				Data:  cn,
			})
		}
	}(res.CommentNotif)

	go realtimeService.SendPostUpdate(postId, map[string]any{
		"post_id":               postId,
		"latest_comments_count": res.LatestCommentsCount,
	})

	return res.NewCommentData, nil
}

func GetCommentsOnPost(ctx context.Context, clientUsername, postId string, limit int, offset int64) (any, error) {
	comments, err := post.GetComments(ctx, clientUsername, postId, limit, time.UnixMilli(offset).UTC())
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

	go realtimeService.SendPostUpdate(postId, map[string]any{
		"post_id":               postId,
		"latest_comments_count": latestCommentsCount,
	})

	return appGlobals.OprSucc, nil
}

func ReactToComment(ctx context.Context, clientUsername, commentId string, reaction rune) (any, error) {
	res, err := comment.ReactTo(ctx, clientUsername, commentId, reaction)
	if err != nil {
		return nil, err
	}

	go func(reactionNotif map[string]any) {
		if rn := reactionNotif; rn != nil {
			receiverUsername := rn["receiver_username"].(string)

			delete(rn, "receiver_username")

			// send notification with message broker
			messageBrokerService.Send(fmt.Sprintf("user-%s-alerts", receiverUsername), messageBrokerService.Message{
				Event: "new notification",
				Data:  rn,
			})
		}
	}(res.ReactionNotif)

	go realtimeService.SendCommentUpdate(commentId, map[string]any{
		"comment_id":             commentId,
		"latest_reactions_count": res.LatestReactionsCount,
	})

	return appGlobals.OprSucc, nil
}

func GetReactorsToComment(ctx context.Context, clientUsername, commentId string, limit int, offset int64) (any, error) {
	reactors, err := comment.GetReactors(ctx, clientUsername, commentId, limit, time.UnixMilli(offset).UTC())
	if err != nil {
		return nil, err
	}

	return reactors, nil
}

func GetReactorsWithReactionToComment(ctx context.Context, clientUsername, commentId string, reaction rune, limit int, offset int64) (any, error) {
	reactors, err := comment.GetReactorsWithReaction(ctx, clientUsername, commentId, reaction, limit, time.UnixMilli(offset).UTC())
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

	go realtimeService.SendCommentUpdate(commentId, map[string]any{
		"comment_id":             commentId,
		"latest_reactions_count": latestReactionsCount,
	})

	return appGlobals.OprSucc, nil
}

func CommentOnComment(ctx context.Context, clientUsername, commentId, commentText string, attachmentData []byte) (map[string]any, error) {

	mime := mimetype.Detect(attachmentData)
	fileType := mime.String()
	fileExt := mime.Extension()

	if !strings.HasPrefix(fileType, "image") {
		return nil, fiber.NewError(400, fmt.Sprintf("invalid file type %s, for attachment_data, expected image/*", fileType))
	}

	attachmentUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("comment_on_comment_attachments/user-%s", clientUsername), attachmentData, fileExt)
	if err != nil {
		return nil, err
	}

	mentions := utilServices.ExtractMentions(commentText)

	res, err := comment.CommentOn(ctx, clientUsername, commentId, commentText, attachmentUrl, mentions)
	if err != nil {
		return nil, err
	}

	go func(mentionNotifs []map[string]any) {
		for _, mn := range mentionNotifs {
			mn := mn
			receiverUsername := mn["receiver_username"].(string)

			delete(mn, "receiver_username")

			// send notification with message broker
			messageBrokerService.Send(fmt.Sprintf("user-%s-alerts", receiverUsername), messageBrokerService.Message{
				Event: "new notification",
				Data:  mn,
			})
		}
	}(res.MentionNotifs)

	go func(commentNotif map[string]any) {
		if cn := commentNotif; cn != nil {
			receiverUsername := cn["receiver_username"].(string)

			delete(cn, "receiver_username")

			// send notification with message broker
			messageBrokerService.Send(fmt.Sprintf("user-%s-alerts", receiverUsername), messageBrokerService.Message{
				Event: "new notification",
				Data:  cn,
			})
		}
	}(res.CommentNotif)

	go realtimeService.SendCommentUpdate(commentId, map[string]any{
		"comment_id":            commentId,
		"latest_comments_count": res.LatestCommentsCount,
	})

	return res.NewCommentData, nil
}

func GetCommentsOnComment(ctx context.Context, clientUsername, commentId string, limit int, offset int64) (any, error) {
	comments, err := comment.GetComments(ctx, clientUsername, commentId, limit, time.UnixMilli(offset).UTC())
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

	go realtimeService.SendCommentUpdate(parentCommentId, map[string]any{
		"comment_id":            parentCommentId,
		"latest_comments_count": latestCommentsCount,
	})

	return appGlobals.OprSucc, nil
}

func CreateRepost(ctx context.Context, clientUsername, postId string) (any, error) {
	res, err := post.Repost(ctx, clientUsername, postId)
	if err != nil {
		return nil, err
	}

	go func(repostNotif map[string]any) {
		if rn := repostNotif; rn != nil {
			receiverUsername := rn["receiver_username"].(string)

			delete(rn, "receiver_username")

			// send notification with message broker
			messageBrokerService.Send(fmt.Sprintf("user-%s-alerts", receiverUsername), messageBrokerService.Message{
				Event: "new notification",
				Data:  rn,
			})
		}
	}(res.RepostNotif)

	go realtimeService.SendPostUpdate(postId, map[string]any{
		"post_id":              postId,
		"latest_reposts_count": res.LatestRepostsCount,
	})

	return res.RepostData, nil
}

func SavePost(ctx context.Context, clientUsername, postId string) (any, error) {
	latestSavesCount, err := post.Save(ctx, clientUsername, postId)
	if err != nil {
		return nil, err
	}

	go realtimeService.SendPostUpdate(postId, map[string]any{
		"post_id":            postId,
		"latest_saves_count": latestSavesCount,
	})

	return appGlobals.OprSucc, nil
}

func UndoSavePost(ctx context.Context, clientUsername, postId string) (any, error) {
	latestSavesCount, err := post.UndoSave(ctx, clientUsername, postId)
	if err != nil {
		return nil, err
	}

	go realtimeService.SendPostUpdate(postId, map[string]any{
		"post_id":            postId,
		"latest_saves_count": latestSavesCount,
	})

	return appGlobals.OprSucc, nil
}
