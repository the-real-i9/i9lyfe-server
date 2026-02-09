package chatModel

import (
	"context"
	"fmt"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/appTypes/UITypes"
	"i9lyfe/src/helpers"
	"i9lyfe/src/helpers/pgDB"
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

type NewMessageT struct {
	Id                   string         `json:"id" db:"id_"`
	ChatHistoryEntryType string         `json:"che_type" db:"che_type"`
	Content              map[string]any `json:"content" db:"content_"`
	DeliveryStatus       string         `json:"delivery_status" db:"delivery_status"`
	CreatedAt            int64          `json:"created_at" db:"created_at"`
	Sender               any            `json:"sender" db:"sender"`
	ReplyTargetMsg       map[string]any `json:"reply_target_msg,omitempty" db:"reply_target_msg"`
	FirstFromUser        bool           `json:"-" db:"ffu"`
	FirstToUser          bool           `json:"-" db:"ftu"`
}

func SendMessage(ctx context.Context, clientUsername, partnerUsername, msgContent string, at int64) (NewMessageT, error) {
	newMessage, err := pgDB.QueryRowType[NewMessageT](
		ctx,
		/* sql */ `
		SELECT id_, che_type, content_, delivery_status, created_at, sender, reply_target_msg, ffu, ftu FROM send_message($1, $2, $3, $4);
		`, clientUsername, partnerUsername, msgContent, at,
	)
	if err != nil {
		helpers.LogError(err)
		return NewMessageT{}, helpers.HandleDBError(err)
	}

	return *newMessage, nil
}

func AckMsgDelivered(ctx context.Context, clientUsername, partnerUsername, msgId string, deliveredAt int64) (bool, error) {
	done, err := pgDB.QueryRowField[bool](
		ctx,
		/* sql */ `
		SELECT * FROM ack_msg($1, $2, $3, $4, $5)
		`, clientUsername, partnerUsername, msgId, "delivered", deliveredAt,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

func AckMsgRead(ctx context.Context, clientUsername, partnerUsername, msgId string, readAt int64) (bool, error) {
	done, err := pgDB.QueryRowField[bool](
		ctx,
		/* sql */ `
		SELECT * FROM ack_msg($1, $2, $3, $4, $5)
		`, clientUsername, partnerUsername, msgId, "read", readAt,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}

func ReplyMessage(ctx context.Context, clientUsername, partnerUsername, targetMsgId, msgContent string, at int64) (NewMessageT, error) {
	newMessage, err := pgDB.QueryRowType[NewMessageT](
		ctx,
		/* sql */ `
		SELECT id_, che_type, content_, delivery_status, created_at, sender, reply_target_msg, ffu, ftu FROM reply_to_msg($1, $2, $3, $4, $5);
		`, clientUsername, partnerUsername, msgContent, at, targetMsgId,
	)
	if err != nil {
		helpers.LogError(err)
		return NewMessageT{}, fiber.ErrInternalServerError
	}

	return *newMessage, nil
}

type RxnToMessageT struct {
	CHEId                string `json:"-" db:"che_id"`
	ChatHistoryEntryType string `json:"che_type" db:"che_type"`
	Emoji                string `json:"emoji" db:"emoji"`
	Reactor              any    `json:"reactor" db:"reactor"`
	ToMsgId              string `json:"-" db:"to_msg_id"`
}

func ReactToMsg(ctx context.Context, clientUsername, partnerUsername, msgId, emoji string, at int64) (RxnToMessageT, error) {
	rxnToMessage, err := pgDB.QueryRowType[RxnToMessageT](
		ctx,
		/* sql */ `
		SELECT che_id, che_type, emoji, reactor, to_msg_id FROM react_to_msg($1, $2, $3, $4, $5)
		`, clientUsername, partnerUsername, msgId, emoji, at,
	)
	if err != nil {
		helpers.LogError(err)
		return RxnToMessageT{}, fiber.ErrInternalServerError
	}

	return *rxnToMessage, nil
}

func RemoveMsgReaction(ctx context.Context, clientUsername, partnerUsername, msgId string) (string, error) {
	CHEId, err := pgDB.QueryRowField[string](
		ctx,
		/* sql */ `
		SELECT * FROM remove_msg_reaction($1, $2, $3)
		`, clientUsername, partnerUsername, msgId,
	)
	if err != nil {
		helpers.LogError(err)
		return "", fiber.ErrInternalServerError
	}
	return *CHEId, nil
}

func DeleteMsg(ctx context.Context, clientUsername, partnerUsername, msgId, deleteFor string, at int64) (bool, error) {
	done, err := pgDB.QueryRowField[bool](
		ctx,
		/* sql */ `
		SELECT * FROM delete_msg($1, $2, $3, $4, $5)
		`, clientUsername, partnerUsername, msgId, deleteFor, at,
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	return *done, nil
}
