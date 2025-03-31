package postCommentService

import (
	"context"
	"fmt"
	"i9lyfe/src/models/postModel"
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

	newPostData, mentionNotifs, err := postModel.New(ctx, clientUsername, mediaUrls, postType, description, mentions, hashtags)
	if err != nil {
		return nil, err
	}

	realtimeService.BroadcastNewPost(newPostData["id"].(string), clientUsername)

	for _, mn := range mentionNotifs {
		receiverUsername := mn["receiver_username"].(string)

		delete(mn, "receiver_username")

		// send notification with message broker
		messageBrokerService.Send(fmt.Sprintf("user-%s-alerts", receiverUsername), messageBrokerService.Message{
			Event: "new notification",
			Data:  mn,
		})
	}

	return newPostData, nil
}
