package modelHelpers

import (
	"context"
	"fmt"
	"i9lyfe/src/appTypes/UITypes"
	"i9lyfe/src/cache"
	"i9lyfe/src/services/cloudStorageService"
)

func BuildPostUIFromCache(ctx context.Context, postId, clientUsername string) (postUI UITypes.Post, err error) {
	nilVal := UITypes.Post{}

	postUI, err = cache.GetPost[UITypes.Post](ctx, postId)
	if err != nil {
		return nilVal, err
	}

	postUI.MediaUrls = cloudStorageService.PostMediaCloudNamesToUrl(postUI.MediaUrls)

	puiou, err := cache.GetUser[UITypes.ContentOwnerUser](ctx, postUI.OwnerUser.(string))
	if err != nil {
		return nilVal, err
	}

	puiou.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(puiou.ProfilePicUrl)

	postUI.OwnerUser = puiou

	if postUI.ReposterUser != nil {
		puiru, err := cache.GetUser[UITypes.ClientUser](ctx, postUI.ReposterUser.(string))
		if err != nil {
			return nilVal, err
		}

		puiru.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(puiru.ProfilePicUrl)

		postUI.ReposterUser = puiru
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

	return postUI, nil
}

func BuildCommentUIFromCache(ctx context.Context, commentId, clientUsername string) (commentUI UITypes.Comment, err error) {
	nilVal := UITypes.Comment{}

	commentUI, err = cache.GetComment[UITypes.Comment](ctx, commentId)
	if err != nil {
		return nilVal, err
	}

	commentUI.AttachmentUrl = cloudStorageService.CommentAttachCloudNameToUrl(commentUI.AttachmentUrl)

	cuiou, err := cache.GetUser[UITypes.ContentOwnerUser](ctx, commentUI.OwnerUser.(string))
	if err != nil {
		return nilVal, err
	}

	cuiou.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(cuiou.ProfilePicUrl)

	commentUI.OwnerUser = cuiou

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

	userSnippetUI.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(userSnippetUI.ProfilePicUrl)

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

	reactorSnippetUI.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(reactorSnippetUI.ProfilePicUrl)

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

	userProfileUI.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(userProfileUI.ProfilePicUrl)

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

		user.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(user.ProfilePicUrl)

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

	csuipu, err := cache.GetUser[UITypes.ChatPartnerUser](ctx, chatSnippetUI.PartnerUser.(string))
	if err != nil {
		return nilVal, err
	}

	csuipu.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(csuipu.ProfilePicUrl)

	chatSnippetUI.PartnerUser = csuipu

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
		cuis, err := cache.GetUser[UITypes.MsgSender](ctx, CHEUI.Sender.(string))
		if err != nil {
			return nilVal, err
		}

		cuis.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(cuis.ProfilePicUrl)

		CHEUI.Sender = cuis

		cloudStorageService.MessageMediaCloudNameToUrl(CHEUI.Content)

		userEmojiMap, err := cache.GetMsgReactions(ctx, CHEId)
		if err != nil {
			return nilVal, err
		}

		msgReactions := []UITypes.MsgReaction{}
		reactionsCount := make(map[string]int, 2)

		for user, emoji := range userEmojiMap {
			var msgr UITypes.MsgReaction

			msgr.Emoji = emoji

			uimsgr, err := cache.GetUser[UITypes.MsgReactor](ctx, user)
			if err != nil {
				return nilVal, err
			}

			uimsgr.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(uimsgr.ProfilePicUrl)

			msgr.Reactor = uimsgr

			msgReactions = append(msgReactions, msgr)
			reactionsCount[emoji]++
		}

		CHEUI.Reactions = msgReactions
		CHEUI.ReactionsCount = reactionsCount

		// TODO: change media_cloud_name for
	}

	return CHEUI, nil
}
