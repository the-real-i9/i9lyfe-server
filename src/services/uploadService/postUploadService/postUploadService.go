package postUploadService

import (
	"context"
	"fmt"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthDataT struct {
	UploadUrl      string `json:"uploadUrl"`
	MediaCloudName string `json:"mediaCloudName"`
}

func Authorize(ctx context.Context, mediaMIME string, mediaCount int) ([]AuthDataT, []string, error) {
	var res []AuthDataT
	var mCloudNames []string

	for i := range mediaCount {
		var blurActualUrl string
		var blurActualMediaCloudName string

		for _, quality := range []string{"blur", "actual"} {
			mediaCloudName := fmt.Sprintf("uploads/post/%d%d/%s-media-%d-%s", time.Now().Year(), time.Now().Month(), uuid.NewString(), i+1, quality)

			url, err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).SignedURL(
				mediaCloudName,
				&storage.SignedURLOptions{
					Scheme:      storage.SigningSchemeV4,
					Method:      "PUT",
					ContentType: mediaMIME,
					Expires:     time.Now().Add(20 * time.Minute),
				},
			)
			if err != nil {
				helpers.LogError(err)
				return nil, nil, fiber.ErrInternalServerError
			}

			if blurActualUrl != "" {
				blurActualUrl += " | "
				blurActualMediaCloudName += " | "
			}

			blurActualUrl += url
			blurActualMediaCloudName += mediaCloudName
		}

		res = append(res, AuthDataT{UploadUrl: blurActualUrl, MediaCloudName: blurActualMediaCloudName})
		mCloudNames = append(mCloudNames, blurActualMediaCloudName)
	}

	return res, mCloudNames, nil
}
