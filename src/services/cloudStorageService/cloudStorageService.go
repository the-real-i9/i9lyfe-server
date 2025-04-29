package cloudStorageService

import (
	"context"
	"errors"
	"fmt"
	"i9lyfe/src/appGlobals"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Upload(ctx context.Context, dirPath string, data []byte, ext string) (string, error) {
	mediaUploadCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	dirPath = fmt.Sprintf("%s/_%d_%s", dirPath, time.Now().UnixNano(), ext)

	bucketName := os.Getenv("GCS_BUCKET")
	mediaUrl := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, dirPath)

	stWriter := appGlobals.GCSClient.Bucket(bucketName).Object(dirPath).NewWriter(mediaUploadCtx)

	stWriter.Write(data)

	err := stWriter.Close()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return "", fiber.NewError(fiber.StatusRequestTimeout, "media upload timed out")
		}

		log.Println("cloudStorageService.go: UploadFile:", err)
		// return "", fiber.ErrInternalServerError // don't do this in production
		return "", nil
	}

	return mediaUrl, nil
}
