package chatMessageModel

import (
	"context"
	"i9lyfe/src/helpers"
	"i9lyfe/src/helpers/pgDB"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type NewMessageT struct {
	Id                   string         `json:"id" db:"id_"`
	ChatHistoryEntryType string         `json:"chat_history_entry_type" db:"che_type"`
	Content              map[string]any `json:"content" db:"content_"`
	DeliveryStatus       string         `json:"delivery_status" db:"delivery_status"`
	CreatedAt            int64          `json:"created_at" db:"created_at"`
	Sender               map[string]any `json:"sender" db:"sender"`
	ReplyTargetMsg       map[string]any `json:"reply_target_msg" db:"reply_target_msg"`
}

func Send(ctx context.Context, clientUsername, partnerUsername, msgContent string, at int64) (NewMessageT, error) {
	newMessage, err := pgDB.QueryRowType[NewMessageT](
		ctx,
		/* sql */ `
		SELECT id_, che_type, content_, delivery_status, created_at, sender, reply_target_msg FROM send_message($1, $2, $3, $4);
		`, clientUsername, partnerUsername, msgContent, at,
	)
	if err != nil {
		helpers.LogError(err)
		return NewMessageT{}, fiber.ErrInternalServerError
	}

	return *newMessage, nil
}

func AckDelivered(ctx context.Context, clientUsername, partnerUsername, msgId string, deliveredAt int64) (bool, error) {
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

func AckRead(ctx context.Context, clientUsername, partnerUsername, msgId string, readAt int64) (bool, error) {
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

func Reply(ctx context.Context, clientUsername, partnerUsername, targetMsgId, msgContent string, at int64) (NewMessageT, error) {
	newMessage, err := pgDB.QueryRowType[NewMessageT](
		ctx,
		/* sql */ `
		SELECT id_, che_type, content_, delivery_status, created_at, sender, reply_target_msg FROM send_message($1, $2, $3, $4, $5);
		`, clientUsername, partnerUsername, msgContent, at, targetMsgId,
	)
	if err != nil {
		helpers.LogError(err)
		return NewMessageT{}, fiber.ErrInternalServerError
	}

	return *newMessage, nil
}

type rxnToMessageT struct {
	Id                   string         `json:"id" db:"che_id"`
	ChatHistoryEntryType string         `json:"chat_history_entry_type" db:"che_type"`
	Emoji                string         `json:"emoji" db:"emoji"`
	At                   int64          `json:"at" db:"at_"`
	Reactor              map[string]any `json:"reactor" db:"reactor"`
	ToMsgId              string         `json:"-" db:"to_msg_id"`
}

func ReactTo(ctx context.Context, clientUsername, partnerUsername, msgId, emoji string, at int64) (rxnToMessageT, error) {
	rxnToMessage, err := pgDB.QueryRowType[rxnToMessageT](
		ctx,
		/* sql */ `
		SELECT che_id, che_type, emoji, at_, reactor, to_msg_id FROM react_to_msg($1, $2, $3, $4, $5)
		`, clientUsername, partnerUsername, msgId, emoji, at,
	)
	if err != nil {
		helpers.LogError(err)
		return rxnToMessageT{}, fiber.ErrInternalServerError
	}

	return *rxnToMessage, nil
}

func RemoveReaction(ctx context.Context, clientUsername, partnerUsername, msgId string) (string, error) {
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

type ChatHistoryEntry struct {
	EntryType string `json:"chat_hist_entry_type"`
	CreatedAt int64  `json:"created_at"`

	// for message entry
	Id             string           `json:"id,omitempty"`
	Content        map[string]any   `json:"content,omitempty"`
	DeliveryStatus string           `json:"delivery_status,omitempty"`
	Sender         map[string]any   `json:"sender,omitempty"`
	IsOwn          bool             `json:"is_own"`
	Reactions      []map[string]any `json:"reactions,omitempty"`

	// for a reply message entry
	ReplyTargetMsg map[string]any `json:"reply_target_msg,omitempty"`

	// for reaction entry
	Reaction string `json:"reaction,omitempty"`
}

func ChatHistory(ctx context.Context, clientUsername, partnerUsername string, limit int, offset time.Time) ([]ChatHistoryEntry, error) {
	var chatHistory []ChatHistoryEntry

	res, err := pgDB.Query(
		ctx,
		`/*cypher*/
		MATCH (clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username })

		OPTIONAL MATCH (clientChat)<-[:IN_CHAT]-(entry:ChatEntry WHERE entry.created_at < $offset)
		OPTIONAL MATCH (entry)<-[:SENDS_MESSAGE]-(senderUser)
		OPTIONAL MATCH (entry)<-[rxn:REACTS_TO_MESSAGE]-(reactorUser)
		OPTIONAL MATCH (entry)-[:REPLIES_TO]->(replyTargetMsg:DMMessage)
		OPTIONAL MATCH (replyTargetMsg)<-[:SENDS_MESSAGE]-(replyTargetMsgSender)

		WITH entry, senderUser, replyTargetMsg, replyTargetMsgSender,
     collect(CASE WHEN rxn IS NOT NULL 
             THEN { reactor: reactorUser { .username, .profile_pic_url }, reaction: rxn.reaction, at: rxn.created_at.epochMillis }
             ELSE NULL 
             END) AS reaction_list

		WITH entry, entry.created_at.epochMillis AS created_at,
			CASE WHEN senderUser IS NOT NULL
				THEN senderUser { .username, .profile_pic_url } 
				ELSE NULL
			END AS sender,
			CASE WHEN senderUser IS NOT NULL AND senderUser.username = $client_username
				THEN true 
				ELSE false
			END AS is_own,
			CASE WHEN size([r IN reaction_list WHERE r IS NOT NULL]) > 0
         THEN [r IN reaction_list WHERE r IS NOT NULL]
         ELSE NULL
			END AS reactions,
			CASE WHEN replyTargetMsg IS NOT NULL
				THEN replyTargetMsg { .id, content: apoc.convert.fromJsonMap(replyTargetMsg.content), sender_username: replyTargetMsgSender.username, is_own: replyTargetMsgSender.username = $client_username }
				ELSE NULL
			END AS reply_target_msg,
			CASE WHEN entry.chat_hist_entry_type = "message"
				THEN apoc.convert.fromJsonMap(entry.content)
				ELSE NULL
			END AS content
		ORDER BY entry.created_at
		LIMIT $limit
		
		RETURN collect(entry { .*, content, created_at, sender, is_own, reactions, reply_target_msg }) AS chat_history
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"limit":            limit,
			"offset":           offset,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	history, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "chat_history")

	helpers.ToStruct(history, &chatHistory)

	return chatHistory, nil
}

func Delete(ctx context.Context, clientUsername, partnerUsername, msgId, deleteFor string, at int64) (bool, error) {
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
