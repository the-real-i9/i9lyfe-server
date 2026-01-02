package modelHelpers

import (
	"context"
	"fmt"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/appTypes/UITypes"
	"i9lyfe/src/cache"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

func BuildPostUIFromCache(ctx context.Context, postId, clientUsername string) (postUI UITypes.Post, err error) {
	nilVal := UITypes.Post{}

	postUI, err = cache.GetPost[UITypes.Post](ctx, postId)
	if err != nil {
		return nilVal, err
	}

	postUI.OwnerUser, err = cache.GetUser[UITypes.ContentOwnerUser](ctx, postUI.OwnerUser.(string))
	if err != nil {
		return nilVal, err
	}

	postUI.ReactionsCount, err = cache.GetPostReactionsCount(ctx, postId)
	if err != nil {
		return nilVal, err
	}

	postUI.CommentsCount, err = cache.GetPostCommentsCount(ctx, postId)
	if err != nil {
		return nilVal, err
	}

	postUI.RepostsCount, err = cache.GetPostRepostsCount(ctx, postId)
	if err != nil {
		return nilVal, err
	}

	postUI.SavesCount, err = cache.GetPostSavesCount(ctx, postId)
	if err != nil {
		return nilVal, err
	}

	postUI.MeReaction, err = cache.GetUserPostReaction(ctx, clientUsername, postId)
	if err != nil {
		return nilVal, err
	}

	postUI.MeSaved, err = cache.UserSavedPost(ctx, clientUsername, postId)
	if err != nil {
		return nilVal, err
	}

	postUI.MeReposted, err = cache.UserRepostedPost(ctx, clientUsername, postId)
	if err != nil {
		return nilVal, err
	}

	var mediaUrls []string

	for _, blurActualMcn := range postUI.MediaCloudNames {
		var blurActualMediaUrl string

		for mcn := range strings.SplitSeq(blurActualMcn, " | ") {
			url, err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).SignedURL(mcn, &storage.SignedURLOptions{
				Scheme:  storage.SigningSchemeV4,
				Method:  "GET",
				Expires: time.Now().Add((6 * 24) * time.Hour),
			})
			if err != nil {
				return nilVal, err
			}

			if blurActualMediaUrl != "" {
				blurActualMediaUrl += " | "
			}

			blurActualMediaUrl += url
		}

		mediaUrls = append(mediaUrls, blurActualMediaUrl)
	}

	postUI.MediaUrls = mediaUrls
	postUI.MediaCloudNames = nil

	return postUI, nil
}

func BuildCommentUIFromCache(ctx context.Context, commentId, clientUsername string) (commentUI UITypes.Comment, err error) {
	nilVal := UITypes.Comment{}

	commentUI, err = cache.GetComment[UITypes.Comment](ctx, commentId)
	if err != nil {
		return nilVal, err
	}

	commentUI.OwnerUser, err = cache.GetUser[UITypes.ContentOwnerUser](ctx, commentUI.OwnerUser.(string))
	if err != nil {
		return nilVal, err
	}

	commentUI.ReactionsCount, err = cache.GetCommentReactionsCount(ctx, commentId)
	if err != nil {
		return nilVal, err
	}

	commentUI.CommentsCount, err = cache.GetCommentCommentsCount(ctx, commentId)
	if err != nil {
		return nilVal, err
	}

	commentUI.MeReaction, err = cache.GetUserCommentReaction(ctx, clientUsername, commentId)
	if err != nil {
		return nilVal, err
	}

	return commentUI, nil
}

func buildUserSnippetUIFromCache(ctx context.Context, username, clientUsername string) (userSnippetUI UITypes.UserSnippet, err error) {
	nilVal := UITypes.UserSnippet{}

	userSnippetUI, err = cache.GetUser[UITypes.UserSnippet](ctx, username)
	if err != nil {
		return nilVal, err
	}

	userSnippetUI.MeFollow, err = cache.MeFollowUser(ctx, clientUsername, username)
	if err != nil {
		return nilVal, err
	}

	userSnippetUI.FollowsMe, err = cache.UserFollowsMe(ctx, clientUsername, username)
	if err != nil {
		return nilVal, err
	}

	return userSnippetUI, nil
}

func buildReactorSnippetUIFromCache(ctx context.Context, username, postOrComment, entityId string) (reactorSnippetUI UITypes.ReactorSnippet, err error) {
	nilVal := UITypes.ReactorSnippet{}

	reactorSnippetUI, err = cache.GetUser[UITypes.ReactorSnippet](ctx, username)
	if err != nil {
		return nilVal, err
	}

	switch postOrComment {
	case "post":
		reactorSnippetUI.Emoji, err = cache.GetUserPostReaction(ctx, entityId, username)
		if err != nil {
			return nilVal, err
		}
	case "comment":
		reactorSnippetUI.Emoji, err = cache.GetUserCommentReaction(ctx, entityId, username)
		if err != nil {
			return nilVal, err
		}
	default:
		return nilVal, fmt.Errorf(`postOrComment wants value "post" or "comment"`)
	}

	return reactorSnippetUI, nil
}

func BuildUserProfileUIFromCache(ctx context.Context, username, clientUsername string) (userProfileUI UITypes.UserProfile, err error) {
	nilVal := UITypes.UserProfile{}

	userProfileUI, err = cache.GetUser[UITypes.UserProfile](ctx, username)
	if err != nil {
		return nilVal, err
	}

	userProfileUI.MeFollow, err = cache.MeFollowUser(ctx, clientUsername, username)
	if err != nil {
		return nilVal, err
	}

	userProfileUI.FollowsMe, err = cache.UserFollowsMe(ctx, clientUsername, username)
	if err != nil {
		return nilVal, err
	}

	userProfileUI.FollowsMe, err = cache.UserFollowsMe(ctx, clientUsername, username)
	if err != nil {
		return nilVal, err
	}

	userProfileUI.PostsCount, err = cache.GetUserPostsCount(ctx, username)
	if err != nil {
		return nilVal, err
	}

	userProfileUI.FollowersCount, err = cache.GetUserFollowersCount(ctx, username)
	if err != nil {
		return nilVal, err
	}

	userProfileUI.FollowingsCount, err = cache.GetUserFollowingsCount(ctx, username)
	if err != nil {
		return nilVal, err
	}

	return userProfileUI, nil
}

func buildNotifSnippetUIFromCache(ctx context.Context, notifId string) (notifSnippetUI UITypes.NotifSnippet, err error) {
	nilVal := UITypes.NotifSnippet{}

	notifSnippetUI, err = cache.GetNotification[UITypes.NotifSnippet](ctx, notifId)
	if err != nil {
		return nilVal, err
	}

	notifSnippetUI.Unread, err = cache.NotificationIsUnread(ctx, notifId)
	if err != nil {
		return nilVal, err
	}

	setNotifUserDetail := func(userKey string) error {
		uname := notifSnippetUI.Details[userKey].(string)
		user, err := cache.GetUser[UITypes.NotifUser](ctx, uname)
		if err != nil {
			return err
		}

		notifSnippetUI.Details[userKey] = user

		return nil
	}

	switch notifSnippetUI.Type {
	case "user_follow":
		if err := setNotifUserDetail("follower_user"); err != nil {
			return nilVal, err
		}
	case "repost":
		if err := setNotifUserDetail("reposter_user"); err != nil {
			return nilVal, err
		}
	case "reaction_to_post", "reaction_to_comment":
		if err := setNotifUserDetail("reactor_user"); err != nil {
			return nilVal, err
		}
	case "mention_in_post", "mention_in_comment":
		if err := setNotifUserDetail("mentioning_user"); err != nil {
			return nilVal, err
		}
	case "comment_on_post", "comment_on_comment":
		if err := setNotifUserDetail("commenter_user"); err != nil {
			return nilVal, err
		}
	}

	return notifSnippetUI, nil
}

func buildChatSnippetUIFromCache(ctx context.Context, clientUsername, partnerUser string) (chatSnippetUI UITypes.ChatSnippet, err error) {
	nilVal := UITypes.ChatSnippet{}

	chatSnippetUI, err = cache.GetChat[UITypes.ChatSnippet](ctx, clientUsername, partnerUser)
	if err != nil {
		return nilVal, err
	}

	chatSnippetUI.PartnerUser, err = cache.GetUser[UITypes.ChatPartnerUser](ctx, chatSnippetUI.PartnerUser.(string))
	if err != nil {
		return nilVal, err
	}

	chatSnippetUI.UnreadMC, err = cache.GetChatUnreadMsgsCount(ctx, clientUsername, partnerUser)
	if err != nil {
		return nilVal, err
	}

	return chatSnippetUI, nil
}

func buildCHEUIFromCache(ctx context.Context, CHEId string) (CHEUI UITypes.ChatHistoryEntry, err error) {
	nilVal := UITypes.ChatHistoryEntry{}

	CHEUI, err = cache.GetChatHistoryEntry[UITypes.ChatHistoryEntry](ctx, CHEId)
	if err != nil {
		return nilVal, err
	}

	if CHEUI.CHEType == "message" {
		CHEUI.Sender, err = cache.GetUser[UITypes.MsgSender](ctx, CHEUI.Sender.(string))
		if err != nil {
			return nilVal, err
		}

		userEmojiMap, err := cache.GetMsgReactions(ctx, CHEId)
		if err != nil {
			return nilVal, err
		}

		msgReactions := []UITypes.MsgReaction{}
		reactionsCount := make(map[string]int, 2)

		for user, emoji := range userEmojiMap {
			var msgr UITypes.MsgReaction

			msgr.Emoji = emoji
			msgr.Reactor, err = cache.GetUser[UITypes.MsgReactor](ctx, user)
			if err != nil {
				return nilVal, err
			}

			msgReactions = append(msgReactions, msgr)
			reactionsCount[emoji]++
		}

		CHEUI.Reactions = msgReactions
		CHEUI.ReactionsCount = reactionsCount
	}

	return CHEUI, nil
}
