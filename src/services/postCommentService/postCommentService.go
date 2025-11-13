package postCommentService

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes"
	"i9lyfe/src/appTypes/UITypes"
	"i9lyfe/src/helpers"
	comment "i9lyfe/src/models/commentModel"
	post "i9lyfe/src/models/postModel"
	"i9lyfe/src/services/cloudStorageService"
	"i9lyfe/src/services/eventStreamService"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"i9lyfe/src/services/utilServices"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v2"
)

func CreateNewPost(ctx context.Context, clientUser appTypes.ClientUser, mediaDataList [][]byte, postType, description string, at int64) (any, error) {
	mediaUrls := make([]string, len(mediaDataList))

	for i, mediaData := range mediaDataList {
		mediaMIME := mimetype.Detect(mediaData)
		mediaType, mediaExt := mediaMIME.String(), mediaMIME.Extension()

		if ((postType == "reel" || postType == "video") && !strings.HasPrefix(mediaType, "video")) || (postType == "photo" && !strings.HasPrefix(mediaType, "image")) {
			return nil, fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid media type %s, for the post type %s", mediaType, postType))
		}

		murl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("post_medias/user-%s/media-%d%s", clientUser.Username, time.Now().UnixNano(), mediaExt), mediaData)
		if err != nil {
			return nil, err
		}

		mediaUrls[i] = murl
	}

	hashtags := utilServices.ExtractHashtags(description)
	mentions := utilServices.ExtractMentions(description)

	newPost, err := post.New(ctx, clientUser.Username, mediaUrls, postType, description, at)
	if err != nil {
		return nil, err
	}

	if newPost.Id != "" {
		go eventStreamService.QueueNewPostEvent(eventTypes.NewPostEvent{
			OwnerUser: clientUser,
			PostId:    newPost.Id,
			PostData:  helpers.ToJson(newPost),
			Hashtags:  hashtags,
			Mentions:  mentions,
			At:        at,
		})
	}

	return newPost, nil
}

func GetPost(ctx context.Context, clientUsername, postId string) (UITypes.Post, error) {
	thePost, err := post.Get(ctx, clientUsername, postId)
	if err != nil {
		return UITypes.Post{}, err
	}

	return thePost, nil
}

func DeletePost(ctx context.Context, clientUsername, postId string) (any, error) {
	mentionedUsers, err := post.Delete(ctx, clientUsername, postId)
	if err != nil {
		return nil, err
	}

	done := mentionedUsers != nil

	if done {
		// run a bg worker that:
		// removes this post and all related data (likes, comments, etc.) from cache
		// mark post and all related data (likes, comments, etc.) as deleted
		go eventStreamService.QueuePostDeletionEvent(eventTypes.PostDeletionEvent{
			OwnerUser: clientUsername,
			PostId:    postId,
			Mentions:  mentionedUsers,
		})
	}

	return done, nil
}

func ReactToPost(ctx context.Context, clientUser appTypes.ClientUser, postId, emoji string, at int64) (any, error) {
	postOwner, err := post.ReactTo(ctx, clientUser.Username, postId, emoji, at)
	if err != nil {
		return nil, err
	}

	done := postOwner != ""

	if done {
		go eventStreamService.QueuePostReactionEvent(eventTypes.PostReactionEvent{
			ReactorUser: clientUser,
			PostOwner:   postOwner,
			PostId:      postId,
			Emoji:       emoji,
			At:          at,
		})
	}

	return done, nil
}

func GetReactorsToPost(ctx context.Context, clientUsername, postId string, limit int, cursor float64) ([]UITypes.ReactorSnippet, error) {
	reactors, err := post.GetReactors(ctx, clientUsername, postId, limit, cursor)
	if err != nil {
		return nil, err
	}

	return reactors, nil
}

/* func GetReactorsWithReactionToPost(ctx context.Context, clientUsername, postId, reaction string, limit int, offset int64) (any, error) {
	reactors, err := post.GetReactorsWithReaction(ctx, clientUsername, postId, reaction, limit, helpers.OffsetTime(offset))
	if err != nil {
		return nil, err
	}

	return reactors, nil
} */

func RemoveReactionToPost(ctx context.Context, clientUsername, postId string) (any, error) {
	done, err := post.RemoveReaction(ctx, clientUsername, postId)
	if err != nil {
		return nil, err
	}

	if done {
		go eventStreamService.QueuePostReactionRemovedEvent(eventTypes.PostReactionRemovedEvent{
			ReactorUser: clientUsername,
			PostId:      postId,
		})
	}

	return done, nil
}

func CommentOnPost(ctx context.Context, clientUser appTypes.ClientUser, postId, commentText string, attachmentData []byte, at int64) (any, error) {
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

		attachmentUrl, err = cloudStorageService.Upload(ctx, fmt.Sprintf("comment_on_post_attachments/user-%s/att-%d%s", clientUser.Username, time.Now().UnixNano(), mediaExt), attachmentData)
		if err != nil {
			return nil, err
		}
	}

	mentions := utilServices.ExtractMentions(commentText)

	newComment, err := post.CommentOn(ctx, clientUser.Username, postId, commentText, attachmentUrl, at)
	if err != nil {
		return nil, err
	}

	if newComment.Id != "" {
		go eventStreamService.QueuePostCommentEvent(eventTypes.PostCommentEvent{
			CommenterUser: clientUser,
			PostId:        postId,
			PostOwner:     newComment.PostOwner,
			CommentId:     newComment.Id,
			CommentData:   helpers.ToJson(newComment),
			Mentions:      mentions,
			At:            at,
		})
	}

	return newComment, nil
}

func GetCommentsOnPost(ctx context.Context, clientUsername, postId string, limit int, cursor float64) ([]UITypes.Comment, error) {
	comments, err := post.GetComments(ctx, clientUsername, postId, limit, cursor)
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func GetComment(ctx context.Context, clientUsername, commentId string) (UITypes.Comment, error) {
	theComment, err := comment.Get(ctx, clientUsername, commentId)
	if err != nil {
		return UITypes.Comment{}, err
	}

	return theComment, nil
}

func RemoveCommentOnPost(ctx context.Context, clientUsername, postId, commentId string) (any, error) {
	done, err := post.RemoveComment(ctx, clientUsername, postId, commentId)
	if err != nil {
		return nil, err
	}

	if done {
		// run a bg worker that:
		// removes this comment and all related data from cache
		// mark comment and all related data (likes, comments, etc.) as deleted
		// publish latest post metric
		go eventStreamService.QueuePostCommentRemovedEvent(eventTypes.PostCommentRemovedEvent{
			CommenterUser: clientUsername,
			PostId:        postId,
			CommentId:     commentId,
		})
	}

	return done, nil
}

func ReactToComment(ctx context.Context, clientUser appTypes.ClientUser, commentId, emoji string, at int64) (any, error) {
	commentOwner, err := comment.ReactTo(ctx, clientUser.Username, commentId, emoji, at)
	if err != nil {
		return nil, err
	}

	done := commentOwner != ""

	if done {
		// look to post reaction bg worker for todos
		go eventStreamService.QueueCommentReactionEvent(eventTypes.CommentReactionEvent{
			ReactorUser:  clientUser,
			CommentId:    commentId,
			CommentOwner: commentOwner,
			Emoji:        emoji,
			At:           at,
		})
	}

	return done, nil
}

func GetReactorsToComment(ctx context.Context, clientUsername, commentId string, limit int, cursor float64) ([]UITypes.ReactorSnippet, error) {
	reactors, err := comment.GetReactors(ctx, clientUsername, commentId, limit, cursor)
	if err != nil {
		return nil, err
	}

	return reactors, nil
}

/* func GetReactorsWithReactionToComment(ctx context.Context, clientUsername, commentId, reaction string, limit int, offset int64) (any, error) {
	reactors, err := comment.GetReactorsWithReaction(ctx, clientUsername, commentId, reaction, limit, helpers.OffsetTime(offset))
	if err != nil {
		return nil, err
	}

	return reactors, nil
} */

func RemoveReactionToComment(ctx context.Context, clientUsername, commentId string) (any, error) {
	done, err := comment.RemoveReaction(ctx, clientUsername, commentId)
	if err != nil {
		return nil, err
	}

	if done {
		// look to post reaction removal worker for todos
		go eventStreamService.QueueCommentReactionRemovedEvent(eventTypes.CommentReactionRemovedEvent{
			ReactorUser: clientUsername,
			CommentId:   commentId,
		})
	}

	return done, nil
}

func CommentOnComment(ctx context.Context, clientUser appTypes.ClientUser, parentCommentId, commentText string, attachmentData []byte, at int64) (any, error) {

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

		attachmentUrl, err = cloudStorageService.Upload(ctx, fmt.Sprintf("comment_on_comment_attachments/user-%s/att-%d%s", clientUser.Username, time.Now().UnixNano(), mediaExt), attachmentData)
		if err != nil {
			return nil, err
		}
	}

	mentions := utilServices.ExtractMentions(commentText)

	newComment, err := comment.CommentOn(ctx, clientUser.Username, parentCommentId, commentText, attachmentUrl, at)
	if err != nil {
		return nil, err
	}

	// create comment, direct
	// add comment id to parentComment comments
	// create and add user mention_in_comment notifications (for mentioned users)
	// and comment_on_comment notifications (for parentCommentOwner user),
	// notifying both users in realtime
	// publish comment metric update
	if newComment.Id != "" {
		go eventStreamService.QueueCommentCommentEvent(eventTypes.CommentCommentEvent{
			CommenterUser:      clientUser,
			ParentCommentId:    parentCommentId,
			ParentCommentOwner: newComment.ParentCommentOwner,
			CommentId:          newComment.Id,
			CommentData:        helpers.ToJson(newComment),
			Mentions:           mentions,
			At:                 at,
		})
	}

	return newComment, nil
}

func GetCommentsOnComment(ctx context.Context, clientUsername, commentId string, limit int, cursor float64) ([]UITypes.Comment, error) {
	comments, err := comment.GetComments(ctx, clientUsername, commentId, limit, cursor)
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func RemoveCommentOnComment(ctx context.Context, clientUser appTypes.ClientUser, parentCommentId, commentId string) (any, error) {
	done, err := comment.RemoveComment(ctx, clientUser.Username, parentCommentId, commentId)
	if err != nil {
		return nil, err
	}

	if done {
		// run a bg worker that:
		// removes this comment and all related data from cache
		// mark comment and all related data (likes, comments, etc.) as deleted
		// publish latest comment metric
		go eventStreamService.QueueCommentCommentRemovedEvent(eventTypes.CommentCommentRemovedEvent{
			CommenterUser:   clientUser,
			ParentCommentId: parentCommentId,
			CommentId:       commentId,
		})
	}

	return done, nil
}

func RepostPost(ctx context.Context, clientUser appTypes.ClientUser, postId string) (any, error) {
	at := time.Now().UnixMilli()

	repost, err := post.Repost(ctx, clientUser.Username, postId, at)
	if err != nil {
		return nil, err
	}

	// cache (re)post list a new post created
	// bg worker: add to (re)post to user posts
	// notify post owner
	// publish post metric
	// fan out repost
	if repost.Id != "" {
		go eventStreamService.QueueRepostEvent(eventTypes.RepostEvent{
			ReposterUser: clientUser,
			PostId:       postId,
			PostOwner:    repost.OwnerUser,
			RepostId:     repost.Id,
			RepostData:   helpers.ToJson(repost),
			At:           at,
		})
	}

	return repost.Id != "", nil
}

func SavePost(ctx context.Context, clientUsername, postId string) (any, error) {
	done, err := post.Save(ctx, clientUsername, postId)
	if err != nil {
		return nil, err
	}

	if done {
		// add saves (saver users) to post
		// add postId to saved posts for saver user
		// publish latest post metric
		go eventStreamService.QueuePostSaveEvent(eventTypes.PostSaveEvent{
			SaverUser: clientUsername,
			PostId:    postId,
		})
	}

	return done, nil
}

func UnsavePost(ctx context.Context, clientUsername, postId string) (any, error) {
	done, err := post.Unsave(ctx, clientUsername, postId)
	if err != nil {
		return nil, err
	}

	if done {
		// add saves (saver users) to post
		// add postId to saved posts for saver user
		// publish latest post metric
		go eventStreamService.QueuePostUnsaveEvent(eventTypes.PostUnsaveEvent{
			SaverUser: clientUsername,
			PostId:    postId,
		})
	}

	return done, nil
}
