package chatDBM

import (
	"context"
	"i9lyfe/src/appGlobals"
	"i9lyfe/src/types/UITypes"

	"i9lyfe/src/helpers"
	"i9lyfe/src/helpers/pgDB"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func redisDB() *redis.Client {
	return appGlobals.RedisClient
}

func MyChats(ctx context.Context, clientUsername string, limit int64, cursor int64) ([]*UITypes.ChatSnippet, error) {
	myChats, err := pgDB.QueryRowsType[UITypes.ChatSnippet](
		ctx,
		/* sql */ `
		SELECT * FROM get_my_chats($1, $2, $3)
		`, clientUsername, limit, cursor,
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return myChats, nil
}

func Delete(ctx context.Context, clientUsername, partnerUsername string) (bool, error) {
	return true, nil
}

func History(ctx context.Context, clientUsername, partnerUsername string, limit int64, cursor float64) (chatHistory []*UITypes.ChatHistoryEntry, err error) {
	history, err := pgDB.QueryRowsType[UITypes.ChatHistoryEntry](
		ctx,
		/* sql */ `
		SELECT * FROM fetch_chat_history($1, $2, $3, $4, $5)
		`, clientUsername, partnerUsername, limit, cursor, "after",
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return history, nil
}

type NewMessageT struct {
	Id             string         `db:"id_"`
	CHEType        string         `db:"che_type"`
	Content        map[string]any `db:"content_"`
	DeliveryStatus *string        `db:"delivery_status"`
	CreatedAt      *int64         `db:"created_at"`
	Sender         map[string]any `db:"sender"`
	ReplyTargetMsg map[string]any `db:"reply_target_msg"`
	Cursor         int64          `db:"cursor_"`
}

func SendMessage(ctx context.Context, clientUsername, partnerUsername, msgContent string, at int64) (NewMessageT, error) {
	newMessage, err := pgDB.QueryRowType[NewMessageT](
		ctx,
		/* sql */ `
		SELECT * FROM send_message($1, $2, $3, $4);
		`, clientUsername, partnerUsername, msgContent, at,
	)
	if err != nil {
		helpers.LogError(err)
		return NewMessageT{}, helpers.HandleDBError(err)
	}

	return *newMessage, nil
}

func AckMsgDelivered(ctx context.Context, clientUsername, partnerUsername string, msgIds []string, deliveredAt int64) (bool, error) {
	done, err := pgDB.QueryRowField[bool](
		ctx,
		/* sql */ `
		SELECT * FROM ack_msg($1, $2, $3, $4, $5)
		`, clientUsername, partnerUsername, msgIds, "delivered", deliveredAt,
	)
	if err != nil {
		helpers.LogError(err)
		return false, helpers.HandleDBError(err)
	}

	return *done, nil
}

func AckMsgRead(ctx context.Context, clientUsername, partnerUsername string, msgIds []string, readAt int64) (bool, error) {
	done, err := pgDB.QueryRowField[bool](
		ctx,
		/* sql */ `
		SELECT * FROM ack_msg($1, $2, $3, $4, $5)
		`, clientUsername, partnerUsername, msgIds, "read", readAt,
	)
	if err != nil {
		helpers.LogError(err)
		return false, helpers.HandleDBError(err)
	}

	return *done, nil
}

func ReplyMessage(ctx context.Context, clientUsername, partnerUsername, targetMsgId, msgContent string, at int64) (NewMessageT, error) {
	newMessage, err := pgDB.QueryRowType[NewMessageT](
		ctx,
		/* sql */ `
		SELECT * FROM reply_to_msg($1, $2, $3, $4, $5);
		`, clientUsername, partnerUsername, msgContent, at, targetMsgId,
	)
	if err != nil {
		helpers.LogError(err)
		return NewMessageT{}, fiber.ErrInternalServerError
	}

	return *newMessage, nil
}

type RxnToMessageT struct {
	CHEId   string         `db:"che_id"`
	CHEType string         `db:"che_type"`
	Emoji   *string        `db:"emoji"`
	Reactor map[string]any `db:"reactor"`
	Cursor  int64          `db:"cursor_"`
	ToMsg   map[string]any `db:"to_msg"`
}

func ReactToMsg(ctx context.Context, clientUsername, partnerUsername, msgId, emoji string, at int64) (RxnToMessageT, error) {
	rxnToMessage, err := pgDB.QueryRowType[RxnToMessageT](
		ctx,
		/* sql */ `
		SELECT * FROM react_to_msg($1, $2, $3, $4, $5)
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
