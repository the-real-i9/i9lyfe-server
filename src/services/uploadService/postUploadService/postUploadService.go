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

func Authorize(ctx context.Context, postType string, mediaMIME [2]string, mediaSizes [][2]int64) ([]AuthDataT, []string, error) {
	var res []AuthDataT
	var mCloudNames []string

	for i, blurSize_realSize := range mediaSizes {
		var blurRealUrl string
		var blurRealMediaCloudName string

		for one, size := range blurSize_realSize {

			which := [2]string{"blur", "real"}

			mediaCloudName := fmt.Sprintf("uploads/post/%s/%d%d/%s-media_%d_%s", postType, time.Now().Year(), time.Now().Month(), uuid.NewString(), i, which[one])

			url, err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).SignedURL(
				mediaCloudName,
				&storage.SignedURLOptions{
					Scheme:      storage.SigningSchemeV4,
					Method:      "PUT",
					ContentType: mediaMIME[one],
					Expires:     time.Now().Add(15 * time.Minute),
					Headers:     []string{fmt.Sprintf("x-goog-content-length-range: %d,%[1]d", size)},
				},
			)
			if err != nil {
				helpers.LogError(err)
				return nil, nil, fiber.ErrInternalServerError
			}

			if blurRealUrl != "" {
				blurRealUrl += " | "
				blurRealMediaCloudName += " | "
			}

			blurRealUrl += url
			blurRealMediaCloudName += mediaCloudName
		}

		res = append(res, AuthDataT{UploadUrl: blurRealUrl, MediaCloudName: blurRealMediaCloudName})
		mCloudNames = append(mCloudNames, blurRealMediaCloudName)
	}

	return res, mCloudNames, nil
}
