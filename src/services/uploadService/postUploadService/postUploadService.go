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

func Authorize(ctx context.Context, postType string, mediaMIME [2]string, mediaSizes [][2]int64) ([]AuthDataT, error) {
	var res []AuthDataT

	for i, blurPlchSize_actualSize := range mediaSizes {
		var blurPlchActualUrl string
		var blurPlchActualMediaCloudName string

		for blurPlch0_actual1, size := range blurPlchSize_actualSize {

			which := [2]string{"blur_placeholder", "actual"}

			mediaCloudName := fmt.Sprintf("uploads/post/%s/%d%d/%s-media_%d_%s", postType, time.Now().Year(), time.Now().Month(), uuid.NewString(), i, which[blurPlch0_actual1])

			url, err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).SignedURL(
				mediaCloudName,
				&storage.SignedURLOptions{
					Scheme:      storage.SigningSchemeV4,
					Method:      "PUT",
					ContentType: mediaMIME[blurPlch0_actual1],
					Expires:     time.Now().Add(15 * time.Minute),
					Headers:     []string{fmt.Sprintf("x-goog-content-length-range: %d,%[1]d", size)},
				},
			)
			if err != nil {
				helpers.LogError(err)
				return nil, fiber.ErrInternalServerError
			}

			if blurPlch0_actual1 == 0 {
				blurPlchActualUrl += "blur_placeholder:"
				blurPlchActualMediaCloudName += "blur_placeholder:"
			} else {
				blurPlchActualUrl += "actual:"
				blurPlchActualMediaCloudName += "actual:"
			}

			blurPlchActualUrl += url
			blurPlchActualMediaCloudName += mediaCloudName

			if blurPlch0_actual1 == 0 {
				blurPlchActualUrl += " "
				blurPlchActualMediaCloudName += " "
			}
		}

		res = append(res, AuthDataT{UploadUrl: blurPlchActualUrl, MediaCloudName: blurPlchActualMediaCloudName})
	}

	return res, nil
}
