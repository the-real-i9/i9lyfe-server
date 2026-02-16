package cloudStorageService

import (
	"fmt"
	"i9lyfe/src/helpers"
	"maps"
)

func ProfilePicCloudNameToUrl(ppicCloudName string) string {
	if ppicCloudName != "{notset}" {
		var (
			smallPPicn  string
			mediumPPicn string
			largePPicn  string
		)

		_, err := fmt.Sscanf(ppicCloudName, "small:%s medium:%s large:%s", &smallPPicn, &mediumPPicn, &largePPicn)
		if err != nil {
			helpers.LogError(err)
		}

		smallPicUrl := GetMediaurl(smallPPicn)
		mediumPicUrl := GetMediaurl(mediumPPicn)
		largePicUrl := GetMediaurl(largePPicn)

		return fmt.Sprintf("small:%s medium:%s large:%s", smallPicUrl, mediumPicUrl, largePicUrl)
	}

	return ppicCloudName
}

func MessageMediaCloudNameToUrl(msgContent map[string]any) map[string]any {
	msgContent = maps.Clone(msgContent)

	contentProps := msgContent["props"].(map[string]any)

	msgContentType := msgContent["type"].(string)

	if msgContentType != "text" {
		mediaCloudName := contentProps["media_cloud_name"].(string)

		if msgContentType == "photo" || msgContentType == "video" {
			var (
				blurPlchMcn string
				actualMcn   string
			)

			_, err := fmt.Sscanf(mediaCloudName, "blur_placeholder:%s actual:%s", &blurPlchMcn, &actualMcn)
			if err != nil {
				helpers.LogError(err)
			}

			blurPlchUrl := GetMediaurl(blurPlchMcn)
			actualUrl := GetMediaurl(actualMcn)

			contentProps["media_url"] = fmt.Sprintf("blur_placeholder:%s actual:%s", blurPlchUrl, actualUrl)
		} else {
			mediaUrl := GetMediaurl(mediaCloudName)

			contentProps["media_url"] = mediaUrl
		}

		delete(contentProps, "media_cloud_name")
	}

	return msgContent
}

func PostMediaCloudNamesToUrl(mediaCloudNames []string) []string {
	var replacement []string

	for _, mcn := range mediaCloudNames {

		var (
			blurPlchMcn string
			actualMcn   string
		)

		_, err := fmt.Sscanf(mcn, "blur_placeholder:%s actual:%s", &blurPlchMcn, &actualMcn)
		if err != nil {
			helpers.LogError(err)
		}

		blurPlchUrl := GetMediaurl(blurPlchMcn)
		actualUrl := GetMediaurl(actualMcn)

		replacement = append(replacement, fmt.Sprintf("blur_placeholder:%s actual:%s", blurPlchUrl, actualUrl))
	}

	return replacement
}

func CommentAttachCloudNameToUrl(attachmentCloudName string) string {
	var attachmentUrl string

	if attachmentCloudName != "" {
		attachmentUrl = GetMediaurl(attachmentCloudName)
	}

	return attachmentUrl
}
