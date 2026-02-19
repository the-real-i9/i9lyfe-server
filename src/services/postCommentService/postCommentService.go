package postCommentService

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes/UITypes"
	"i9lyfe/src/helpers"
	comment "i9lyfe/src/models/commentModel"
	post "i9lyfe/src/models/postModel"
	"i9lyfe/src/services/cloudStorageService"
	"i9lyfe/src/services/eventStreamService"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"regexp"
	"slices"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func extractHashtags(description string) []string {
	re := regexp.MustCompile("#[[:alnum:]][[:alnum:]_]+[[:alnum:]]+")

	matches := re.FindAllString(description, -1)

	exist := make(map[string]bool)

	res := []string{}

	for _, m := range matches {
		ht := m[1:]

		if !exist[ht] {
			res = append(res, ht)
		}

		exist[ht] = true
	}

	return res
}

func extractMentions(description string) []string {
	re := regexp.MustCompile("@[[:alnum:]][[:alnum:]_-]+[[:alnum:]]+")

	matches := re.FindAllString(description, -1)

	exist := make(map[string]bool)

	res := []string{}

	for _, m := range matches {
		mnt := m[1:]

		if !exist[mnt] {
			res = append(res, mnt)
		}

		exist[mnt] = true
	}

	return res
}

func CreateNewPost(ctx context.Context, clientUsername string, mediaCloudNames []string, postType, description string, at int64) (map[string]any, error) {
	hashtags := extractHashtags(description)
	mentions := extractMentions(description)

	newPost, err := post.New(ctx, clientUsername, mediaCloudNames, postType, description, at)
	if err != nil {
		return nil, err
	}

	if newPost.Id == "" {
		return nil, nil
	}

	go func(newPost post.NewPostT, clientUsername string, hashtags, mentions []string, at int64) {
		eventStreamService.QueueNewPostEvent(eventTypes.NewPostEvent{
			OwnerUser:  clientUsername,
			PostId:     newPost.Id,
			PostData:   helpers.ToMsgPack(newPost),
			Hashtags:   hashtags,
			Mentions:   mentions,
			At:         at,
			PostCursor: newPost.Cursor,
		})
	}(newPost, clientUsername, hashtags, mentions, at)

	return map[string]any{"new_post_id": newPost.Id, "post_cursor": newPost.Cursor}, nil
}

func GetPost(ctx context.Context, clientUsername, postId string) (UITypes.Post, error) {
	thePost, err := post.Get(ctx, clientUsername, postId)
	if err != nil {
		return UITypes.Post{}, err
	}

	return thePost, nil
}

func DeletePost(ctx context.Context, clientUsername, postId string) (bool, error) {
	mentionedUsers, err := post.Delete(ctx, clientUsername, postId)
	if err != nil {
		return false, err
	}

	done := mentionedUsers != nil

	if done {
		// run a bg worker that:
		// removes this post and all related data (likes, comments, etc.) from cache
		// mark post and all related data (likes, comments, etc.) as deleted
		go func() {
			eventStreamService.QueuePostDeletionEvent(eventTypes.PostDeletionEvent{
				OwnerUser: clientUsername,
				PostId:    postId,
				Mentions:  mentionedUsers,
			})
		}()
	}

	return done, nil
}

func ReactToPost(ctx context.Context, clientUsername, postId, emoji string, at int64) (bool, error) {
	res, err := post.ReactTo(ctx, clientUsername, postId, emoji, at)
	if err != nil {
		return false, err
	}

	done := res.PostOwnerUser != ""

	if done {
		go eventStreamService.QueuePostReactionEvent(eventTypes.PostReactionEvent{
			ReactorUser: clientUsername,
			PostOwner:   res.PostOwnerUser,
			PostId:      postId,
			Emoji:       emoji,
			At:          at,
			RxnCursor:   res.RxnCursor,
		})
	}

	return done, nil
}

func GetReactorsToPost(ctx context.Context, clientUsername, postId string, limit int64, cursor float64) ([]UITypes.ReactorSnippet, error) {
	reactors, err := post.GetReactors(ctx, clientUsername, postId, limit, cursor)
	if err != nil {
		return nil, err
	}

	return reactors, nil
}

/* func GetReactorsWithReactionToPost(ctx context.Context, clientUsername, postId, reaction string, limit int64, offset int64) (any, error) {
	reactors, err := post.GetReactorsWithReaction(ctx, clientUsername, postId, reaction, limit, helpers.OffsetTime(offset))
	if err != nil {
		return nil, err
	}

	return reactors, nil
} */

func RemoveReactionToPost(ctx context.Context, clientUsername, postId string) (bool, error) {
	done, err := post.RemoveReaction(ctx, clientUsername, postId)
	if err != nil {
		return false, err
	}

	if done {
		go eventStreamService.QueuePostReactionRemovedEvent(eventTypes.PostReactionRemovedEvent{
			ReactorUser: clientUsername,
			PostId:      postId,
		})
	}

	return done, nil
}

func CommentOnPost(ctx context.Context, clientUsername, postId, commentText, attachmentCloudName string, at int64) (map[string]any, error) {
	mentions := extractMentions(commentText)

	newComment, err := post.CommentOn(ctx, clientUsername, postId, commentText, attachmentCloudName, at)
	if err != nil {
		return nil, err
	}

	if newComment.Id == "" {
		return nil, nil
	}

	go func(newComment post.NewCommentT, clientUsername, postId string, mentions []string, at int64) {
		// we're not creating mention notifications for the post owner
		mentions = slices.DeleteFunc(mentions, func(u string) bool {
			return u == newComment.PostOwner
		})

		eventStreamService.QueuePostCommentEvent(eventTypes.PostCommentEvent{
			CommenterUser: clientUsername,
			PostId:        postId,
			PostOwner:     newComment.PostOwner,
			CommentId:     newComment.Id,
			CommentData:   helpers.ToMsgPack(newComment),
			Mentions:      mentions,
			At:            at,
			CommentCursor: newComment.Cursor,
		})
	}(newComment, clientUsername, postId, mentions, at)

	return map[string]any{"new_comment_id": newComment.Id, "comment_cursor": newComment.Cursor}, nil
}

func GetCommentsOnPost(ctx context.Context, clientUsername, postId string, limit int64, cursor float64) ([]UITypes.Comment, error) {
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

func RemoveCommentOnPost(ctx context.Context, clientUsername, postId, commentId string) (bool, error) {
	done, err := post.RemoveComment(ctx, clientUsername, postId, commentId)
	if err != nil {
		return false, err
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

func ReactToComment(ctx context.Context, clientUsername, commentId, emoji string, at int64) (bool, error) {
	res, err := comment.ReactTo(ctx, clientUsername, commentId, emoji, at)
	if err != nil {
		return false, err
	}

	done := res.CommentOwnerUser != ""

	if done {
		// look to post reaction bg worker for todos
		go eventStreamService.QueueCommentReactionEvent(eventTypes.CommentReactionEvent{
			ReactorUser:  clientUsername,
			CommentId:    commentId,
			CommentOwner: res.CommentOwnerUser,
			Emoji:        emoji,
			At:           at,
			RxnCursor:    res.RxnCursor,
		})
	}

	return done, nil
}

func GetReactorsToComment(ctx context.Context, clientUsername, commentId string, limit int64, cursor float64) ([]UITypes.ReactorSnippet, error) {
	reactors, err := comment.GetReactors(ctx, clientUsername, commentId, limit, cursor)
	if err != nil {
		return nil, err
	}

	return reactors, nil
}

/* func GetReactorsWithReactionToComment(ctx context.Context, clientUsername, commentId, reaction string, limit int64, offset int64) (any, error) {
	reactors, err := comment.GetReactorsWithReaction(ctx, clientUsername, commentId, reaction, limit, helpers.OffsetTime(offset))
	if err != nil {
		return nil, err
	}

	return reactors, nil
} */

func RemoveReactionToComment(ctx context.Context, clientUsername, commentId string) (bool, error) {
	done, err := comment.RemoveReaction(ctx, clientUsername, commentId)
	if err != nil {
		return false, err
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

func CommentOnComment(ctx context.Context, clientUsername, parentCommentId, commentText, attachmentCloudName string, at int64) (map[string]any, error) {
	mentions := extractMentions(commentText)

	newComment, err := comment.CommentOn(ctx, clientUsername, parentCommentId, commentText, attachmentCloudName, at)
	if err != nil {
		return nil, err
	}

	if newComment.Id == "" {
		return nil, nil
	}

	go func(newComment comment.NewCommentT, clientUsername, parentCommentId string, mentions []string, at int64) {
		eventStreamService.QueueCommentCommentEvent(eventTypes.CommentCommentEvent{
			CommenterUser:      clientUsername,
			ParentCommentId:    parentCommentId,
			ParentCommentOwner: newComment.ParentCommentOwner,
			CommentId:          newComment.Id,
			CommentData:        helpers.ToMsgPack(newComment),
			Mentions:           mentions,
			At:                 at,
			CommentCursor:      newComment.Cursor,
		})
	}(newComment, clientUsername, parentCommentId, mentions, at)

	return map[string]any{"new_comment_id": newComment.Id, "comment_cursor": newComment.Cursor}, nil
}

func GetCommentsOnComment(ctx context.Context, clientUsername, commentId string, limit int64, cursor float64) ([]UITypes.Comment, error) {
	comments, err := comment.GetComments(ctx, clientUsername, commentId, limit, cursor)
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func RemoveCommentOnComment(ctx context.Context, clientUsername, parentCommentId, commentId string) (bool, error) {
	done, err := comment.RemoveComment(ctx, clientUsername, parentCommentId, commentId)
	if err != nil {
		return false, err
	}

	if done {
		// run a bg worker that:
		// removes this comment and all related data from cache
		// mark comment and all related data (likes, comments, etc.) as deleted
		// publish latest comment metric
		go eventStreamService.QueueCommentCommentRemovedEvent(eventTypes.CommentCommentRemovedEvent{
			CommenterUser:   clientUsername,
			ParentCommentId: parentCommentId,
			CommentId:       commentId,
		})
	}

	return done, nil
}

func RepostPost(ctx context.Context, clientUsername, postId string) (bool, error) {
	at := time.Now().UnixMilli()

	repost, err := post.Repost(ctx, clientUsername, postId, at)
	if err != nil {
		return false, err
	}

	done := repost.Id != ""

	if done {
		go func(repost post.RepostT, clientUsername string, at int64) {
			eventStreamService.QueueRepostEvent(eventTypes.RepostEvent{
				ReposterUser: clientUsername,
				PostId:       repost.RepostedPostId,
				PostOwner:    repost.OwnerUser.(string),
				RepostId:     repost.Id,
				RepostData:   helpers.ToMsgPack(repost),
				At:           at,
				RepostCursor: repost.Cursor,
			})
		}(repost, clientUsername, at)
	}

	return done, nil
}

func SavePost(ctx context.Context, clientUsername, postId string) (bool, error) {
	saveCursor, err := post.Save(ctx, clientUsername, postId)
	if err != nil {
		return false, err
	}

	done := saveCursor != 0

	if done {
		// add saves (saver users) to post
		// add postId to saved posts for saver user
		// publish latest post metric
		go eventStreamService.QueuePostSaveEvent(eventTypes.PostSaveEvent{
			SaverUser:  clientUsername,
			PostId:     postId,
			SaveCursor: saveCursor,
		})
	}

	return done, nil
}

func UnsavePost(ctx context.Context, clientUsername, postId string) (bool, error) {
	done, err := post.Unsave(ctx, clientUsername, postId)
	if err != nil {
		return false, err
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

/* ------------- */

type AuthCommAttDataT struct {
	UploadUrl           string `msgpack:"uploadUrl"`
	AttachmentCloudName string `msgpack:"attachmentCloudName"`
}

func AuthorizeCommAttUpload(ctx context.Context, attachmentMIME string) (AuthCommAttDataT, error) {
	var res AuthCommAttDataT

	attachmentCloudName := fmt.Sprintf("uploads/comment/%d%d/%s", time.Now().Year(), time.Now().Month(), uuid.NewString())

	url, err := cloudStorageService.GetUploadUrl(attachmentCloudName, attachmentMIME)
	if err != nil {
		return AuthCommAttDataT{}, fiber.ErrInternalServerError
	}

	res.UploadUrl = url
	res.AttachmentCloudName = attachmentCloudName

	return res, nil
}

type AuthPostMediaDataT struct {
	UploadUrl      string `msgpack:"uploadUrl"`
	MediaCloudName string `msgpack:"mediaCloudName"`
}

func AuthorizePostMediaUpload(ctx context.Context, postType string, mediaMIME [2]string, mediaCount int) ([]AuthPostMediaDataT, error) {
	var res []AuthPostMediaDataT

	for i := range mediaCount {
		var blurPlchActualUrl string
		var blurPlchActualMediaCloudName string

		for blurPlch0_actual1, mime := range mediaMIME {

			which := [2]string{"blur_placeholder", "actual"}

			mediaCloudName := fmt.Sprintf("uploads/post/%s/%d%d/%s-media_%d_%s", postType, time.Now().Year(), time.Now().Month(), uuid.NewString(), i, which[blurPlch0_actual1])

			url, err := cloudStorageService.GetUploadUrl(mediaCloudName, mime)
			if err != nil {
				return nil, fiber.ErrInternalServerError
			}

			if blurPlch0_actual1 == 0 {
				blurPlchActualUrl += "blur_placeholder:"
				blurPlchActualMediaCloudName += "blur_placeholder:"
			} else {
				blurPlchActualUrl += "actual:"
				blurPlchActualMediaCloudName += "actual:"
			}

			blurPlchActualUrl += url
			blurPlchActualMediaCloudName += mediaCloudName

			if blurPlch0_actual1 == 0 {
				blurPlchActualUrl += " "
				blurPlchActualMediaCloudName += " "
			}
		}

		res = append(res, AuthPostMediaDataT{UploadUrl: blurPlchActualUrl, MediaCloudName: blurPlchActualMediaCloudName})
	}

	return res, nil
}
