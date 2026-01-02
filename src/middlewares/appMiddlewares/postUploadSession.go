package appMiddlewares

import (
	"errors"
	"fmt"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	"os"
	"slices"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/gofiber/fiber/v2"
)

func PostUploadSession(c *fiber.Ctx) error {
	pusData := helpers.FromJson[map[string]any](c.Cookies("session"))["post_upload"]

	if pusData == nil {
		return c.Status(fiber.StatusUnauthorized).SendString("upload post media first")
	}

	var body struct {
		Type            string   `json:"type"`
		MediaCloudNames []string `json:"media_cloud_names"`
	}

	err := c.BodyParser(&body)
	if err != nil {
		return err
	}

	if pusData.(map[string]any)["postType"].(string) != body.Type {
		return c.Status(fiber.StatusBadRequest).SendString("'type' field differs from uploaded media's post type")
	}

	if !slices.Equal(pusData.(map[string]any)["mediaCloudNames"].([]string), body.MediaCloudNames) {
		return c.Status(fiber.StatusBadRequest).SendString("'media_cloud_names' field differs from uploaded media list")
	}

	for _, blurActualMcn := range body.MediaCloudNames {
		for which, mcn := range strings.Split(blurActualMcn, " | ") {
			metadata, err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(mcn).Attrs(c.Context())
			if errors.Is(err, storage.ErrObjectNotExist) {
				return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("upload error: media (%s) does not exist in cloud", mcn))
			}

			const (
				BLUR   int = 0
				ACTUAL int = 1
			)

			if which == BLUR && (metadata.Size < 1*1024 || metadata.Size > 10*1024) {
				return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("upload error: blur media (%s) size is out of range", mcn))
			}

			if which == ACTUAL && (metadata.Size < 1*1024*1024 || metadata.Size > 8*1024*1024) {
				return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("upload error: actual media (%s) size is out of range", mcn))
			}

		}
	}

	return c.Next()
}
