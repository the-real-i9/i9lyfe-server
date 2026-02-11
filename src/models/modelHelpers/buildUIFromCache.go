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
	var reposterUsername string = "/nil/"
	if postUI.ReposterUser != nil {
		reposterUsername = postUI.ReposterUser.(string)
	}

	pud, err := cache.GetPostUIData(ctx, postId, postUI.OwnerUser.(string), reposterUsername, clientUsername)
	if err != nil {
		return nilVal, err
	}

	postUI.MediaUrls = cloudStorageService.PostMediaCloudNamesToUrl(postUI.MediaUrls)

	puiou, err := pud.OwnerUser()
	if err != nil {
		return nilVal, err
	}

	puiou.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(puiou.ProfilePicUrl)

	postUI.OwnerUser = puiou

	if postUI.ReposterUser != nil {
		puiru, err := pud.ReposterUser()
		if err != nil {
			return nilVal, err
		}

		puiru.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(puiru.ProfilePicUrl)

		postUI.ReposterUser = puiru
	}

	postUI.ReactionsCount, err = pud.ReactionsCount()
	if err != nil {
		return nilVal, err
	}

	postUI.CommentsCount, err = pud.CommentsCount()
	if err != nil {
		return nilVal, err
	}

	postUI.RepostsCount, err = pud.RepostsCount()
	if err != nil {
		return nilVal, err
	}

	postUI.SavesCount, err = pud.SavesCount()
	if err != nil {
		return nilVal, err
	}

	postUI.MeReaction, err = pud.MeReaction()
	if err != nil {
		return nilVal, err
	}

	postUI.MeSaved, err = pud.MeSaved()
	if err != nil {
		return nilVal, err
	}

	postUI.MeReposted, err = pud.MeReposted()
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

	cud, err := cache.GetCommentUIData(ctx, commentId, commentUI.OwnerUser.(string), clientUsername)
	if err != nil {
		return nilVal, err
	}

	commentUI.AttachmentUrl = cloudStorageService.CommentAttachCloudNameToUrl(commentUI.AttachmentUrl)

	cuiou, err := cud.OwnerUser()
	if err != nil {
		return nilVal, err
	}

	cuiou.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(cuiou.ProfilePicUrl)

	commentUI.OwnerUser = cuiou

	commentUI.ReactionsCount, err = cud.ReactionsCount()
	if err != nil {
		return nilVal, err
	}

	commentUI.CommentsCount, err = cud.CommentsCount()
	if err != nil {
		return nilVal, err
	}

	commentUI.MeReaction, err = cud.MeReaction()
	if err != nil {
		return nilVal, err
	}

	return commentUI, nil
}

func buildUserSnippetUIFromCache(ctx context.Context, username, clientUsername string) (userSnippetUI UITypes.UserSnippet, err error) {
	nilVal := UITypes.UserSnippet{}

	usud, err := cache.GetUserSnippetUIData(ctx, username, clientUsername)
	if err != nil {
		return nilVal, err
	}

	userSnippetUI, err = usud.User()
	if err != nil {
		return nilVal, err
	}

	userSnippetUI.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(userSnippetUI.ProfilePicUrl)

	userSnippetUI.MeFollow, err = usud.MeFollow()
	if err != nil {
		return nilVal, err
	}

	userSnippetUI.FollowsMe, err = usud.FollowsMe()
	if err != nil {
		return nilVal, err
	}

	return userSnippetUI, nil
}

func buildReactorSnippetUIFromCache(ctx context.Context, username, postOrComment, entityId string) (reactorSnippetUI UITypes.ReactorSnippet, err error) {
	nilVal := UITypes.ReactorSnippet{}

	rsud, err := cache.GetReactorSnippetUIData(ctx, username, postOrComment, entityId)
	if err != nil {
		return nilVal, err
	}

	reactorSnippetUI, err = rsud.User()
	if err != nil {
		return nilVal, err
	}

	reactorSnippetUI.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(reactorSnippetUI.ProfilePicUrl)

	reactorSnippetUI.Emoji, err = rsud.UserReaction()
	if err != nil {
		return nilVal, err
	}

	return reactorSnippetUI, nil
}

func BuildUserProfileUIFromCache(ctx context.Context, username, clientUsername string) (userProfileUI UITypes.UserProfile, err error) {
	nilVal := UITypes.UserProfile{}

	upud, err := cache.GetUserProfileUIData(ctx, username, clientUsername)
	if err != nil {
		return nilVal, err
	}

	userProfileUI, err = upud.User()
	if err != nil {
		return nilVal, err
	}

	userProfileUI.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(userProfileUI.ProfilePicUrl)

	userProfileUI.MeFollow, err = upud.MeFollow()
	if err != nil {
		return nilVal, err
	}

	userProfileUI.FollowsMe, err = upud.FollowsMe()
	if err != nil {
		return nilVal, err
	}

	userProfileUI.PostsCount, err = upud.PostsCount()
	if err != nil {
		return nilVal, err
	}

	userProfileUI.FollowersCount, err = upud.FollowersCount()
	if err != nil {
		return nilVal, err
	}

	userProfileUI.FollowingsCount, err = upud.FollowingsCount()
	if err != nil {
		return nilVal, err
	}

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
