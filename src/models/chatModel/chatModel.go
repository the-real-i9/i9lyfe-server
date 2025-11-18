package chatModel

import (
	"context"
	"fmt"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/appTypes/UITypes"
	"i9lyfe/src/helpers"
	"i9lyfe/src/models/modelHelpers"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func redisDB() *redis.Client {
	return appGlobals.RedisClient
}

func MyChats(ctx context.Context, clientUsername string, limit int, cursor float64) (myChats []UITypes.ChatSnippet, err error) {
	partnerMembers, err := redisDB().ZRevRangeByScoreWithScores(ctx, fmt.Sprintf("user:%s:chats_sorted", clientUsername), &redis.ZRangeBy{
		Max:   helpers.MaxCursor(cursor),
		Min:   "-inf",
		Count: int64(limit),
	}).Result()
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	mychats, err := modelHelpers.ChatPartnerMembersForUIChatSnippets(ctx, partnerMembers, clientUsername)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return mychats, nil
}

func Delete(ctx context.Context, clientUsername, partnerUsername string) (bool, error) {
	return true, nil
}

func History(ctx context.Context, clientUsername, partnerUsername string, limit int, cursor float64) (chatHistory []UITypes.ChatHistoryEntry, err error) {
	CHEMembers, err := redisDB().ZRevRangeByScoreWithScores(ctx, fmt.Sprintf("chat:owner:%s:partner:%s:history", clientUsername, partnerUsername), &redis.ZRangeBy{
		Max:   helpers.MaxCursor(cursor),
		Min:   "-inf",
		Count: int64(limit),
	}).Result()
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	chatHistory, err = modelHelpers.CHEMembersForUICHEs(ctx, CHEMembers)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return chatHistory, nil
}
