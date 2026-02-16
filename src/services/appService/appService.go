package appService

/* import (
	"context"
	"fmt"
	"i9lyfe/src/models/appModel"
	"i9lyfe/src/services/contentRecommendationService"
	"slices"

	"github.com/gofiber/fiber/v3"
)

func FetchPosts(ctx context.Context, clientUsername string, types, hashtags []string) (any, error) {
	for _, t := range types {
		if !slices.Contains([]string{"photo", "video", "reel"}, t) {
			return nil, fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid type, %s. valid types are photo, video and reel", t))
		}
	}

	if len(types) == 0 {
		types = []string{"photo", "video", "reel"}
	}

	posts, err := contentRecommendationService.FetchPosts(ctx, clientUsername, types, hashtags)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func FetchUsers(ctx context.Context) ([]any, error) {
	users, err := appModel.TopUsers(ctx)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func FetchHashtags(ctx context.Context) ([]any, error) {
	ths, err := appModel.TopHashtags(ctx)
	if err != nil {
		return nil, err
	}

	return ths, nil
}

func Search(ctx context.Context, clientUsername, term, filter string, limit int, offset int64) ([]any, error) {
	if term == "" {
		return nil, fiber.NewError(fiber.StatusBadRequest, "no search term provided")
	}

	if filter != "" && !slices.Contains([]string{"user", "photo", "video", "reel", "hashtag"}, filter) {
		return nil, fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid filter, %s. valid filters are user, photo, video and reel", filter))
	}

	var (
		results []any
		err     error
	)

	switch filter {
	case "photo", "video", "reel":
		results, err = appModel.SearchPost(ctx, clientUsername, filter, term)
	case "user":
		results, err = appModel.SearchUsers(ctx, term)
	case "hashtag":
		results, err = appModel.SearchHashtags(ctx, term)
	}

	if err != nil {
		return nil, err
	}

	return results, nil
}
*/
