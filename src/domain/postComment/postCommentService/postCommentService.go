package postCommentService

import (
	"context"
	"fmt"

	"i9lyfe/src/appGlobals"
	comment "i9lyfe/src/domain/postComment/commentDBM"
	post "i9lyfe/src/domain/postComment/postDBM"
	"i9lyfe/src/domain/user/userService"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService"
	"i9lyfe/src/services/mediaStorageService"
	"i9lyfe/src/services/sseService"
	"i9lyfe/src/types/UITypes"
	"i9lyfe/src/types/appTypes"
	"i9lyfe/src/types/eventTypes"

	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"

	"github.com/jackc/pgx/v5/pgxpool"
)

func dbPool() *pgxpool.Pool {
	return appGlobals.DBPool
}

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

func CreateNewPost(ctx context.Context, clientUsername string, mediaCloudNames []string, postType, description string, at int64) (UITypes.NewPost, error) {
	pgTx, err := dbPool().Begin(ctx)
	if err != nil {
		helpers.LogError(err)
		return UITypes.NewPost{}, fiber.ErrInternalServerError
	}

	defer func() {
		if err != nil {
			go helpers.LogError(pgTx.Rollback(ctx))
		}
	}()

	hashtags := extractHashtags(description)
	mentions := extractMentions(description)

	newPost, err := post.New(ctx, pgTx, clientUsername, mediaCloudNames, postType, description, at, mentions, hashtags)
	if err != nil {
		return UITypes.NewPost{}, err
	}

	if newPost.Id == "" {
		return UITypes.NewPost{}, nil
	}

	newPost.Cursor += time.Now().UnixMicro()

	err = eventStreamService.QueueNewPostEvent(ctx, eventTypes.NewPostEvent{
		OwnerUsername: clientUsername,
		PostId:        newPost.Id,
		PostCursor:    newPost.Cursor,
	})
	if err != nil {
		return UITypes.NewPost{}, fiber.ErrInternalServerError
	}

	err = pgTx.Commit(ctx)
	if err != nil {
		helpers.LogError(err)
		return UITypes.NewPost{}, fiber.ErrInternalServerError
	}

	newPost.OwnerUser["profile_pic_url"] = mediaStorageService.ProfilePicCloudNameToUrl(newPost.OwnerUser["profile_pic_url"].(string))

	newPost.MediaUrls = mediaStorageService.PostMediaCloudNamesToUrl(newPost.MediaUrls)

	go func(mentNotifIds []string) {
		notifs, err := userService.GetManyNotifs(ctx, mentNotifIds)
		if err != nil {
			return
		}

		for _, n := range notifs {
			sseService.SendEventMsg(n.OwnerUsername, appTypes.ServerEventMsg{
				Event: "new notification",
				Data:  n,
			})
		}

	}(newPost.MentNotifIds)

	return newPost, nil
}

func GetPost(ctx context.Context, clientUsername, postId string) (UITypes.Post, error) {
	thePost, err := post.Get(ctx, clientUsername, postId)
	if err != nil {
		return UITypes.Post{}, err
	}

	thePost.OwnerUser["profile_pic_url"] = mediaStorageService.ProfilePicCloudNameToUrl(thePost.OwnerUser["profile_pic_url"].(string))

	thePost.MediaUrls = mediaStorageService.PostMediaCloudNamesToUrl(thePost.MediaUrls)

	return thePost, nil
}

func DeletePost(ctx context.Context, clientUsername, postId string) (bool, error) {
	done, err := post.Delete(ctx, clientUsername, postId)
	if err != nil {
		return false, err
	}

	return done, nil
}

func ReactToPost(ctx context.Context, clientUsername, postId, emoji string, at int64) (bool, error) {
	pgTx, err := dbPool().Begin(ctx)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	defer func() {
		if err != nil {
			go helpers.LogError(pgTx.Rollback(ctx))
		}
	}()

	reactNotifId, err := post.ReactTo(ctx, pgTx, clientUsername, postId, emoji, at)
	if err != nil {
		return false, err
	}

	done := reactNotifId != ""

	if done {
		err = eventStreamService.QueuePostReactionEvent(ctx, eventTypes.PostReactionEvent{
			PostId: postId,
		})
		if err != nil {
			return false, fiber.ErrInternalServerError
		}
	}

	err = pgTx.Commit(ctx)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	go func() {
		notif, err := userService.GetOneNotif(ctx, reactNotifId)
		if err != nil {
			return
		}

		sseService.SendEventMsg(notif.OwnerUsername, appTypes.ServerEventMsg{
			Event: "new notification",
			Data:  notif,
		})
	}()

	return done, nil
}

func GetReactorsToPost(ctx context.Context, postId string, limit int64, cursor float64) ([]*UITypes.ReactorSnippet, error) {
	reactors, err := post.GetReactors(ctx, postId, limit, cursor)
	if err != nil {
		return nil, err
	}

	for _, r := range reactors {
		r.ProfilePicUrl = mediaStorageService.ProfilePicCloudNameToUrl(r.ProfilePicUrl)
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
	pgTx, err := dbPool().Begin(ctx)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	defer func() {
		if err != nil {
			go helpers.LogError(pgTx.Rollback(ctx))
		}
	}()

	done, err := post.RemoveReaction(ctx, pgTx, clientUsername, postId)
	if err != nil {
		return false, err
	}

	if done {
		err = eventStreamService.QueuePostReactionRemovedEvent(ctx, eventTypes.PostReactionRemovedEvent{
			PostId: postId,
		})
		if err != nil {
			return false, fiber.ErrInternalServerError
		}
	}

	err = pgTx.Commit(ctx)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return done, nil
}

func CommentOnPost(ctx context.Context, clientUsername, postId, commentText, attachmentCloudName string, at int64) (UITypes.NewComment, error) {
	pgTx, err := dbPool().Begin(ctx)
	if err != nil {
		helpers.LogError(err)
		return UITypes.NewComment{}, fiber.ErrInternalServerError
	}

	defer func() {
		if err != nil {
			go helpers.LogError(pgTx.Rollback(ctx))
		}
	}()

	mentions := extractMentions(commentText)

	newComment, err := post.CommentOn(ctx, pgTx, clientUsername, postId, commentText, attachmentCloudName, at, mentions)
	if err != nil {
		return UITypes.NewComment{}, err
	}

	if newComment.Id == "" {
		return UITypes.NewComment{}, nil
	}

	err = eventStreamService.QueuePostCommentEvent(ctx, eventTypes.PostCommentEvent{
		PostId: postId,
	})
	if err != nil {
		return UITypes.NewComment{}, fiber.ErrInternalServerError
	}

	err = pgTx.Commit(ctx)
	if err != nil {
		helpers.LogError(err)
		return UITypes.NewComment{}, fiber.ErrInternalServerError
	}

	newComment.OwnerUser["profile_pic_url"] = mediaStorageService.ProfilePicCloudNameToUrl(newComment.OwnerUser["profile_pic_url"].(string))

	newComment.AttachmentUrl = mediaStorageService.CommentAttachCloudNameToUrl(newComment.AttachmentUrl)

	go func(newComment UITypes.NewComment) {
		notifs, err := userService.GetManyNotifs(ctx, append(newComment.MentNotifIds, newComment.CommentNotifId))
		if err != nil {
			return
		}

		for _, n := range notifs {
			sseService.SendEventMsg(n.OwnerUsername, appTypes.ServerEventMsg{
				Event: "new notification",
				Data:  n,
			})
		}

	}(newComment)

	return newComment, nil
}

func GetCommentsOnPost(ctx context.Context, clientUsername, postId string, limit int64, cursor float64) ([]*UITypes.Comment, error) {
	comments, err := post.GetComments(ctx, clientUsername, postId, limit, cursor)
	if err != nil {
		return nil, err
	}

	for _, c := range comments {
		c.OwnerUser["profile_pic_url"] = mediaStorageService.ProfilePicCloudNameToUrl(c.OwnerUser["profile_pic_url"].(string))

		c.AttachmentUrl = mediaStorageService.CommentAttachCloudNameToUrl(c.AttachmentUrl)
	}

	return comments, nil
}

func RemoveCommentOnPost(ctx context.Context, clientUsername, postId, commentId string) (bool, error) {
	pgTx, err := dbPool().Begin(ctx)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	defer func() {
		if err != nil {
			go helpers.LogError(pgTx.Rollback(ctx))
		}
	}()

	done, err := post.RemoveComment(ctx, pgTx, clientUsername, postId, commentId)
	if err != nil {
		return false, err
	}

	if done {
		err = eventStreamService.QueuePostCommentRemovedEvent(ctx, eventTypes.PostCommentRemovedEvent{
			PostId: postId,
		})
		if err != nil {
			return false, fiber.ErrInternalServerError
		}
	}

	err = pgTx.Commit(ctx)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return done, nil
}

func ReactToComment(ctx context.Context, clientUsername, commentId, emoji string, at int64) (bool, error) {
	pgTx, err := dbPool().Begin(ctx)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	defer func() {
		if err != nil {
			go helpers.LogError(pgTx.Rollback(ctx))
		}
	}()

	reactNotifId, err := comment.ReactTo(ctx, pgTx, clientUsername, commentId, emoji, at)
	if err != nil {
		return false, err
	}

	done := reactNotifId != ""

	if done {
		err = eventStreamService.QueueCommentReactionEvent(ctx, eventTypes.CommentReactionEvent{
			CommentId: commentId,
		})
		if err != nil {
			return false, fiber.ErrInternalServerError
		}
	}

	err = pgTx.Commit(ctx)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	go func() {
		notif, err := userService.GetOneNotif(ctx, reactNotifId)
		if err != nil {
			return
		}

		sseService.SendEventMsg(notif.OwnerUsername, appTypes.ServerEventMsg{
			Event: "new notification",
			Data:  notif,
		})
	}()

	return done, nil
}

func GetReactorsToComment(ctx context.Context, commentId string, limit int64, cursor float64) ([]*UITypes.ReactorSnippet, error) {
	reactors, err := comment.GetReactors(ctx, commentId, limit, cursor)
	if err != nil {
		return nil, err
	}

	for _, r := range reactors {
		r.ProfilePicUrl = mediaStorageService.ProfilePicCloudNameToUrl(r.ProfilePicUrl)
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
	pgTx, err := dbPool().Begin(ctx)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	defer func() {
		if err != nil {
			go helpers.LogError(pgTx.Rollback(ctx))
		}
	}()

	done, err := comment.RemoveReaction(ctx, pgTx, clientUsername, commentId)
	if err != nil {
		return false, err
	}

	if done {
		err = eventStreamService.QueueCommentReactionRemovedEvent(ctx, eventTypes.CommentReactionRemovedEvent{
			CommentId: commentId,
		})
		if err != nil {
			return false, fiber.ErrInternalServerError
		}
	}

	err = pgTx.Commit(ctx)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return done, nil
}

func CommentOnComment(ctx context.Context, clientUsername, parentCommentId, commentText, attachmentCloudName string, at int64) (UITypes.NewComment, error) {
	pgTx, err := dbPool().Begin(ctx)
	if err != nil {
		helpers.LogError(err)
		return UITypes.NewComment{}, fiber.ErrInternalServerError
	}

	defer func() {
		if err != nil {
			go helpers.LogError(pgTx.Rollback(ctx))
		}
	}()

	mentions := extractMentions(commentText)

	newComment, err := comment.CommentOn(ctx, pgTx, clientUsername, parentCommentId, commentText, attachmentCloudName, at, mentions)
	if err != nil {
		return UITypes.NewComment{}, err
	}

	if newComment.Id == "" {
		return UITypes.NewComment{}, nil
	}

	err = eventStreamService.QueueCommentCommentEvent(ctx, eventTypes.CommentCommentEvent{
		ParentCommentId: parentCommentId,
	})
	if err != nil {
		return UITypes.NewComment{}, fiber.ErrInternalServerError
	}

	err = pgTx.Commit(ctx)
	if err != nil {
		helpers.LogError(err)
		return UITypes.NewComment{}, fiber.ErrInternalServerError
	}

	newComment.OwnerUser["profile_pic_url"] = mediaStorageService.ProfilePicCloudNameToUrl(newComment.OwnerUser["profile_pic_url"].(string))

	newComment.AttachmentUrl = mediaStorageService.CommentAttachCloudNameToUrl(newComment.AttachmentUrl)

	go func(newComment UITypes.NewComment) {
		notifs, err := userService.GetManyNotifs(ctx, append(newComment.MentNotifIds, newComment.CommentNotifId))
		if err != nil {
			return
		}

		for _, n := range notifs {
			sseService.SendEventMsg(n.OwnerUsername, appTypes.ServerEventMsg{
				Event: "new notification",
				Data:  n,
			})
		}

	}(newComment)

	return newComment, nil
}

func GetCommentsOnComment(ctx context.Context, clientUsername, commentId string, limit int64, cursor float64) ([]*UITypes.Comment, error) {
	comments, err := comment.GetComments(ctx, clientUsername, commentId, limit, cursor)
	if err != nil {
		return nil, err
	}

	for _, c := range comments {
		c.OwnerUser["profile_pic_url"] = mediaStorageService.ProfilePicCloudNameToUrl(c.OwnerUser["profile_pic_url"].(string))

		c.AttachmentUrl = mediaStorageService.CommentAttachCloudNameToUrl(c.AttachmentUrl)
	}

	return comments, nil
}

func RemoveCommentOnComment(ctx context.Context, clientUsername, parentCommentId, commentId string) (bool, error) {
	pgTx, err := dbPool().Begin(ctx)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	defer func() {
		if err != nil {
			go helpers.LogError(pgTx.Rollback(ctx))
		}
	}()

	done, err := comment.RemoveComment(ctx, pgTx, clientUsername, parentCommentId, commentId)
	if err != nil {
		return false, err
	}

	if done {
		err = eventStreamService.QueueCommentCommentRemovedEvent(ctx, eventTypes.CommentCommentRemovedEvent{
			ParentCommentId: parentCommentId,
		})
		if err != nil {
			return false, fiber.ErrInternalServerError
		}
	}

	err = pgTx.Commit(ctx)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return done, nil
}

func GetComment(ctx context.Context, clientUsername, commentId string) (UITypes.Comment, error) {
	theComment, err := comment.Get(ctx, clientUsername, commentId)
	if err != nil {
		return UITypes.Comment{}, err
	}

	theComment.OwnerUser["profile_pic_url"] = mediaStorageService.ProfilePicCloudNameToUrl(theComment.OwnerUser["profile_pic_url"].(string))

	theComment.AttachmentUrl = mediaStorageService.CommentAttachCloudNameToUrl(theComment.AttachmentUrl)

	return theComment, nil
}

func RepostPost(ctx context.Context, clientUsername, postId string) (map[string]any, error) {
	pgTx, err := dbPool().Begin(ctx)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	defer func() {
		if err != nil {
			go helpers.LogError(pgTx.Rollback(ctx))
		}
	}()

	repost, err := post.Repost(ctx, pgTx, clientUsername, postId)
	if err != nil {
		return nil, err
	}

	done := repost.Id != ""

	repost.Cursor += time.Now().UnixMicro()

	if done {
		err = eventStreamService.QueueRepostEvent(ctx, eventTypes.RepostEvent{
			PostId:       postId,
			ReposterUser: clientUsername,
			RepostId:     repost.Id,
			RepostCursor: repost.Cursor,
		})
		if err != nil {
			return nil, fiber.ErrInternalServerError
		}
	}

	err = pgTx.Commit(ctx)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	go func(repost post.RepostT) {
		notif, err := userService.GetOneNotif(ctx, repost.NotifId)
		if err != nil {
			return
		}

		sseService.SendEventMsg(notif.OwnerUsername, appTypes.ServerEventMsg{
			Event: "new notification",
			Data:  notif,
		})

	}(repost)

	return map[string]any{"repost_cursor": repost.Cursor}, nil
}

func SavePost(ctx context.Context, clientUsername, postId string) (bool, error) {
	pgTx, err := dbPool().Begin(ctx)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	defer func() {
		if err != nil {
			go helpers.LogError(pgTx.Rollback(ctx))
		}
	}()

	done, err := post.Save(ctx, pgTx, clientUsername, postId)
	if err != nil {
		return false, err
	}

	if done {
		err = eventStreamService.QueuePostSaveEvent(ctx, eventTypes.PostSaveEvent{
			PostId: postId,
		})
		if err != nil {
			return false, fiber.ErrInternalServerError
		}
	}

	err = pgTx.Commit(ctx)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return done, nil
}

func UnsavePost(ctx context.Context, clientUsername, postId string) (bool, error) {
	pgTx, err := dbPool().Begin(ctx)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	defer func() {
		if err != nil {
			go helpers.LogError(pgTx.Rollback(ctx))
		}
	}()

	done, err := post.Unsave(ctx, pgTx, clientUsername, postId)
	if err != nil {
		return false, err
	}

	if done {
		err = eventStreamService.QueuePostUnsaveEvent(ctx, eventTypes.PostUnsaveEvent{
			PostId: postId,
		})
		if err != nil {
			return false, fiber.ErrInternalServerError
		}
	}

	err = pgTx.Commit(ctx)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return done, nil
}

/* ------------- */

type AuthCommAttDataT struct {
	UploadUrl           string `msgpack:"uploadUrl"`
	AttachmentCloudName string `msgpack:"attachmentCloudName"`
}

func AuthorizedCommAttUpload(ctx context.Context, attachmentMIME string) (AuthCommAttDataT, error) {
	var res AuthCommAttDataT

	attachmentCloudName := fmt.Sprintf("uploads/comment/%d%d/%s", time.Now().Year(), time.Now().Month(), utils.UUIDv4())

	url, err := mediaStorageService.GetUploadUrl(attachmentCloudName, attachmentMIME)
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

func AuthorizedPostMediaUpload(ctx context.Context, postType string, mediaMIME [2]string, mediaCount int) ([]AuthPostMediaDataT, error) {
	var res []AuthPostMediaDataT

	for i := range mediaCount {
		var blurPlchActualUrl strings.Builder
		var blurPlchActualMediaCloudName strings.Builder

		for blurPlch0_actual1, mime := range mediaMIME {

			which := [2]string{"blur_placeholder", "actual"}

			mediaCloudName := fmt.Sprintf("uploads/post/%s/%d%d/%s-media_%d_%s", postType, time.Now().Year(), time.Now().Month(), utils.UUIDv4(), i, which[blurPlch0_actual1])

			url, err := mediaStorageService.GetUploadUrl(mediaCloudName, mime)
			if err != nil {
				return nil, fiber.ErrInternalServerError
			}

			if blurPlch0_actual1 == 0 {
				blurPlchActualUrl.WriteString("blur_placeholder:")
				blurPlchActualMediaCloudName.WriteString("blur_placeholder:")
			} else {
				blurPlchActualUrl.WriteString("actual:")
				blurPlchActualMediaCloudName.WriteString("actual:")
			}

			blurPlchActualUrl.WriteString(url)
			blurPlchActualMediaCloudName.WriteString(mediaCloudName)

			if blurPlch0_actual1 == 0 {
				blurPlchActualUrl.WriteString(" ")
				blurPlchActualMediaCloudName.WriteString(" ")
			}
		}

		res = append(res, AuthPostMediaDataT{UploadUrl: blurPlchActualUrl.String(), MediaCloudName: blurPlchActualMediaCloudName.String()})
	}

	return res, nil
}
