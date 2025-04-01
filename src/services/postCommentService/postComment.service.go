package postCommentService

import (
	"context"
	"fmt"
	post "i9lyfe/src/models/postModel"
	"i9lyfe/src/services/cloudStorageService"
	"i9lyfe/src/services/messageBrokerService"
	"i9lyfe/src/services/realtimeService"
	"i9lyfe/src/services/utilServices"
)

func CreateNewPost(ctx context.Context, clientUsername string, mediaDataList [][]byte, postType, description string) (map[string]any, error) {
	mediaUrls := make([]string, len(mediaDataList))

	for i, mediaData := range mediaDataList {
		murl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("post_medias/user-%s", clientUsername), mediaData)
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

	for _, mn := range res.MentionNotifs {
		receiverUsername := mn["receiver_username"].(string)

		delete(mn, "receiver_username")

		// send notification with message broker
		messageBrokerService.Send(fmt.Sprintf("user-%s-alerts", receiverUsername), messageBrokerService.Message{
			Event: "new notification",
			Data:  mn,
		})
	}

	return res.NewPostData, nil
}

func GetPost(ctx context.Context, clientUsername, postId string) (any, error) {
	res, err := post.FindOne(ctx, clientUsername, postId)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func ReactToPost(ctx context.Context, clientUsername, postId, reaction string) (any, error) {
	res, err := post.ReactTo(ctx, clientUsername, postId, reaction)
	if err != nil {
		return nil, err
	}

	if rn := res.ReactionNotif; rn != nil {
		receiverUsername := rn["receiver_username"].(string)

		delete(rn, "receiver_username")

		// send notification with message broker
		messageBrokerService.Send(fmt.Sprintf("user-%s-alerts", receiverUsername), messageBrokerService.Message{
			Event: "new notification",
			Data:  rn,
		})
	}
	// return res, nil
}
