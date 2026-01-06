package gcsHelpers

import (
	"context"
	"errors"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/helpers"
	"os"

	"cloud.google.com/go/storage"
)

func GetMediaInfo(ctx context.Context, mcn string) *storage.ObjectAttrs {
	mInfo, err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(mcn).Attrs(ctx)
	if err != nil {
		if !errors.Is(err, storage.ErrObjectNotExist) {
			helpers.LogError(err)
		}

		return nil
	}

	return mInfo
}

func DeleteCloudMedia(ctx context.Context, mcn string) {
	err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(mcn).Delete(ctx)
	if err != nil {
		helpers.LogError(err)
	}
}
