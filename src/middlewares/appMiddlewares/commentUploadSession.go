package appMiddlewares

import (
	"errors"
	"fmt"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	"os"

	"cloud.google.com/go/storage"
	"github.com/gofiber/fiber/v2"
)

func CommentUploadSession(c *fiber.Ctx) error {
	cusData := helpers.FromJson[map[string]any](c.Cookies("session"))["comment_upload"]

	if cusData == nil {
		return c.Next()
	}

	var body struct {
		AttachmentCloudName string `json:"attachment_cloud_name"`
	}

	err := c.BodyParser(&body)
	if err != nil {
		return err
	}

	if cusData.(map[string]any)["attachmentCloudName"].(string) != body.AttachmentCloudName {
		return c.Status(fiber.StatusBadRequest).SendString("'attachment_cloud_name' field differs from uploaded media attachment cloud name")
	}

	_, err = appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(body.AttachmentCloudName).Attrs(c.Context())
	if errors.Is(err, storage.ErrObjectNotExist) {
		return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("upload error: media (%s) does not exist in cloud", body.AttachmentCloudName))
	}

	return c.Next()
}
