package commentUploadService

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
	UploadUrl           string `json:"uploadUrl"`
	AttachmentCloudName string `json:"attachmentCloudName"`
}

func Authorize(ctx context.Context, attachmentMIME string, attachmentSize int64) (AuthDataT, string, error) {
	var res AuthDataT

	attachmentCloudName := fmt.Sprintf("uploads/comment/%d%d/%s", time.Now().Year(), time.Now().Month(), uuid.NewString())

	url, err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).SignedURL(
		attachmentCloudName,
		&storage.SignedURLOptions{
			Scheme:      storage.SigningSchemeV4,
			Method:      "PUT",
			ContentType: attachmentMIME,
			Expires:     time.Now().Add(15 * time.Minute),
			Headers:     []string{fmt.Sprintf("x-goog-content-length-range: %d,%[1]d", attachmentSize)},
		},
	)
	if err != nil {
		helpers.LogError(err)
		return AuthDataT{}, "", fiber.ErrInternalServerError
	}

	res.UploadUrl = url
	res.AttachmentCloudName = attachmentCloudName

	return res, attachmentCloudName, nil
}
