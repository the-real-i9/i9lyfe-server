package modelHelpers

import (
	"context"
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

	reposterUsername, _ := postUI.ReposterUser.(string)

	pud, err := cache.GetPostUIData(ctx, postId, postUI.OwnerUser.(string), reposterUsername, clientUsername)
	if err != nil {
		return nilVal, err
	}

	pud.OwnerUser.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(pud.OwnerUser.ProfilePicUrl)

	postUI.OwnerUser = pud.OwnerUser

	if postUI.ReposterUser != nil {
		pud.ReposterUser.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(pud.ReposterUser.ProfilePicUrl)

		postUI.ReposterUser = pud.ReposterUser
	}

	postUI.ReactionsCount = pud.ReactionsCount
	postUI.CommentsCount = pud.CommentsCount
	postUI.RepostsCount = pud.RepostsCount
	postUI.SavesCount = pud.SavesCount
	postUI.MeReaction = pud.MeReaction
	postUI.MeSaved = pud.MeSaved
	postUI.MeReposted = pud.MeReposted

	return postUI, nil
}

func BuildCommentUIFromCache(ctx context.Context, commentId, clientUsername string) (commentUI UITypes.Comment, err error) {
	nilVal := UITypes.Comment{}

	commentUI, err = cache.GetComment[UITypes.Comment](ctx, commentId)
	if err != nil {
		return nilVal, err
	}

	commentUI.AttachmentUrl = cloudStorageService.CommentAttachCloudNameToUrl(commentUI.AttachmentUrl)

	cud, err := cache.GetCommentUIData(ctx, commentId, commentUI.OwnerUser.(string), clientUsername)
	if err != nil {
		return nilVal, err
	}

	cud.OwnerUser.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(cud.OwnerUser.ProfilePicUrl)

	commentUI.OwnerUser = cud.OwnerUser
	commentUI.ReactionsCount = cud.ReactionsCount
	commentUI.CommentsCount = cud.CommentsCount
	commentUI.MeReaction = cud.MeReaction

	return commentUI, nil
}

func buildUserSnippetUIFromCache(ctx context.Context, username, clientUsername string) (userSnippetUI UITypes.UserSnippet, err error) {
	nilVal := UITypes.UserSnippet{}

	usud, err := cache.GetUserSnippetUIData(ctx, username, clientUsername)
	if err != nil {
		return nilVal, err
	}

	usud.User.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(usud.User.ProfilePicUrl)

	userSnippetUI = usud.User
	userSnippetUI.MeFollow = usud.MeFollow
	userSnippetUI.FollowsMe = usud.FollowsMe

	return userSnippetUI, nil
}

func buildReactorSnippetUIFromCache(ctx context.Context, username, postOrComment, entityId string) (reactorSnippetUI UITypes.ReactorSnippet, err error) {
	nilVal := UITypes.ReactorSnippet{}

	rsud, err := cache.GetReactorSnippetUIData(ctx, username, postOrComment, entityId)
	if err != nil {
		return nilVal, err
	}
	rsud.User.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(rsud.User.ProfilePicUrl)

	reactorSnippetUI = rsud.User
	reactorSnippetUI.Emoji = rsud.UserReaction

	return reactorSnippetUI, nil
}

func BuildUserProfileUIFromCache(ctx context.Context, username, clientUsername string) (userProfileUI UITypes.UserProfile, err error) {
	nilVal := UITypes.UserProfile{}

	upud, err := cache.GetUserProfileUIData(ctx, username, clientUsername)
	if err != nil {
		return nilVal, err
	}
	upud.User.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(upud.User.ProfilePicUrl)

	userProfileUI = upud.User
	userProfileUI.MeFollow = upud.MeFollow
	userProfileUI.FollowsMe = upud.FollowsMe
	userProfileUI.PostsCount = upud.PostsCount
	userProfileUI.FollowersCount = upud.FollowersCount
	userProfileUI.FollowingsCount = upud.FollowingsCount

	return userProfileUI, nil
}

func BuildNotifSnippetUIFromCache(ctx context.Context, notifId string) (notifSnippetUI UITypes.NotifSnippet, err error) {
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

func BuildCHEUIFromCache(ctx context.Context, CHEId string) (CHEUI UITypes.ChatHistoryEntry, err error) {
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
	}

	return CHEUI, nil
}
