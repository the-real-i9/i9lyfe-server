package chatUploadService

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

func Authorize(ctx context.Context, msgType, mediaMIME string, mediaSize int64) (AuthDataT, error) {
	var res AuthDataT

	mediaCloudName := fmt.Sprintf("uploads/chat/%s/%d%d/%s", msgType, time.Now().Year(), time.Now().Month(), uuid.NewString())

	url, err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).SignedURL(
		mediaCloudName,
		&storage.SignedURLOptions{
			Scheme:      storage.SigningSchemeV4,
			Method:      "PUT",
			ContentType: mediaMIME,
			Expires:     time.Now().Add(15 * time.Minute),
			Headers:     []string{fmt.Sprintf("x-goog-content-length-range: %d,%[1]d", mediaSize)},
		},
	)
	if err != nil {
		helpers.LogError(err)
		return AuthDataT{}, fiber.ErrInternalServerError
	}

	res.UploadUrl = url
	res.MediaCloudName = mediaCloudName

	return res, nil
}

func AuthorizeVisual(ctx context.Context, msgType string, mediaMIME [2]string, mediaSize [2]int64) (AuthDataT, error) {
	var res AuthDataT

	for one, size := range mediaSize {
		which := [2]string{"blur", "real"}

		mediaCloudName := fmt.Sprintf("uploads/chat/%s/%d%d/%s-%s", msgType, time.Now().Year(), time.Now().Month(), uuid.NewString(), which[one])

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
			return AuthDataT{}, fiber.ErrInternalServerError
		}

		if res.UploadUrl != "" {
			res.UploadUrl += " | "
			res.MediaCloudName += " | "
		}

		res.UploadUrl += url
		res.MediaCloudName += mediaCloudName
	}

	return res, nil
}
