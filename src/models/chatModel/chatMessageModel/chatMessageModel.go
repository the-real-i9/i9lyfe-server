package chatMessageModel

import (
	"context"
	"i9lyfe/src/helpers"
	"i9lyfe/src/models/db"
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
	newMessage, err := db.QueryRowType[NewMessageT](
		ctx,
		/* sql */ `
		SELECT id_, che_type, content_, delivery_status, created_at, sender, reply_target_msg FROM send_message($1, $2, $3, $4, $5, $6);
		`, clientUsername, partnerUsername, msgContent, at, false, nil,
	)
	if err != nil {
		helpers.LogError(err)
		return NewMessageT{}, fiber.ErrInternalServerError
	}

	return *newMessage, nil
}

func AckDelivered(ctx context.Context, clientUsername, partnerUsername, msgId string, deliveredAt time.Time) (bool, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username }),
      (clientChat)<-[:IN_CHAT]-(message:DMMessage{ id: $message_id, delivery_status: "sent" })<-[:RECEIVES_MESSAGE]-()
    SET message.delivery_status = "delivered", message.delivered_at = $delivered_at, clientChat.unread_messages_count = coalesce(clientChat.unread_messages_count, 0) + 1

		RETURN true AS done
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_id":       msgId,
			"delivered_at":     deliveredAt,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return false, nil
	}

	done, _, _ := neo4j.GetRecordValue[bool](res.Records[0], "done")

	return done, nil
}

func AckRead(ctx context.Context, clientUsername, partnerUsername, msgId string, readAt time.Time) (bool, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username }),
      (clientChat)<-[:IN_CHAT]-(message:DMMessage{ id: $message_id } WHERE message.delivery_status IN ["sent", "delivered"])<-[:RECEIVES_MESSAGE]-()

    WITH clientChat, message, CASE coalesce(clientChat.unread_messages_count, 0) WHEN <> 0 THEN clientChat.unread_messages_count - 1 ELSE 0 END AS unread_messages_count
    SET message.delivery_status = "read", message.read_at = $read_at, clientChat.unread_messages_count = unread_messages_count

		RETURN true AS done
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_id":       msgId,
			"read_at":          readAt,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return false, nil
	}

	done, _, _ := neo4j.GetRecordValue[bool](res.Records[0], "done")

	return done, nil
}

func Reply(ctx context.Context, clientUsername, partnerUsername, targetMsgId, msgContent string, at int64) (NewMessageT, error) {
	newMessage, err := db.QueryRowType[NewMessageT](
		ctx,
		/* sql */ `
		SELECT id_, che_type, content_, delivery_status, created_at, sender, reply_target_msg FROM send_message($1, $2, $3, $4, $5, $6);
		`, clientUsername, partnerUsername, msgContent, at, true, targetMsgId,
	)
	if err != nil {
		helpers.LogError(err)
		return NewMessageT{}, fiber.ErrInternalServerError
	}

	return *newMessage, nil
}

type RxnToMessage struct {
	ClientData  map[string]any `json:"client_resp"`
	PartnerData map[string]any `json:"partner_resp"`
}

func ReactTo(ctx context.Context, clientUsername, partnerUsername, msgId, reaction string, at time.Time) (RxnToMessage, error) {
	var rxnToMessage RxnToMessage

	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientUser)-[:HAS_CHAT]->(clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username })-[:WITH_USER]->(partnerUser),
			(clientChat)<-[:IN_CHAT]-(message:DMMessage{ id: $message_id }),
			(partnerUser)-[:HAS_CHAT]->(partnerChat)-[:WITH_USER]->(clientUser)
		
		WITH clientUser, message, partnerUser, partnerChat
		MERGE (msgrxn:DMMessageReaction:ChatEntry{ reactor_username: clientUser.username, message_id: message.id })
		SET msgrxn.reaction = $reaction, msgrxn.chat_hist_entry_type = "reaction", msgrxn.created_at = $at

		MERGE (clientUser)-[crxn:REACTS_TO_MESSAGE]->(message)
		SET crxn.reaction = $reaction, crxn.created_at = $at
		
		MERGE (clientUser)-[:SENDS_REACTION]->(msgrxn)-[:IN_CHAT]->(clientChat)
		MERGE (partnerUser)-[:RECEIVES_REACTION]->(msgrxn)-[:IN_CHAT]->(partnerChat)

		WITH clientUser.username AS partner_username, message.id AS msg_id, 
			clientUser { .username, .profile_pic_url } AS reactor, crxn

		RETURN true AS client_resp,
			{ partner_username, msg_id, reactor, reaction: crxn.reaction, at: crxn.created_at.epochMillis } AS partner_resp

		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_id":       msgId,
			"reaction":         reaction,
			"at":               at,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return rxnToMessage, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return rxnToMessage, nil
	}

	helpers.ToStruct(res.Records[0].AsMap(), &rxnToMessage)

	return rxnToMessage, nil
}

func RemoveReaction(ctx context.Context, clientUsername, partnerUsername, msgId string) (bool, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientUser)-[:HAS_CHAT]->(clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username })-[:WITH_USER]->(partnerUser),
			(partnerUser)-[:HAS_CHAT]->(partnerChat)-[:WITH_USER]->(clientUser),
			(clientChat)<-[:IN_CHAT]-(message:DMMessage{ id: $message_id }),

			(msgrxn:DMMessageReaction:ChatEntry{ reactor_username: $client_username, message_id: $message_id }),
			(clientUser)-[crxn:REACTS_TO_MESSAGE]->(message)

		DETACH DELETE msgrxn, crxn
		
		RETURN true AS done
    `,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_id":       msgId,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return false, nil
	}

	done, _, _ := neo4j.GetRecordValue[bool](res.Records[0], "done")
	return done, nil
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

	res, err := db.Query(
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

func Delete(ctx context.Context, clientUsername, partnerUsername, msgId, deleteFor string) (bool, error) {
	var query string

	if deleteFor == "everyone" {
		query = `
			MATCH (clientUser)-[:HAS_CHAT]->(clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username })-[:WITH_USER]->(partnerUser),
				(partnerUser)-[:HAS_CHAT]->(partnerChat)-[:WITH_USER]->(clientUser)

			MATCH (clientChat)<-[msgicc:IN_CHAT]-(message:Message{ id: $message_id })<-[:SENDS_MESSAGE]-(),
				(partnerChat)<-[msgipc:IN_CHAT]-(message)<-[:RECEIVES_MESSAGE]-()

      SET msgicc.msg_deleted = true, msgipc.msg_deleted = true

			RETURN true AS done
		`
	} else {
		query = `
      MATCH (clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username })<-[msgicc:IN_CHAT]-(message:Message{ id: $message_id })<-[:SENDS_MESSAGE|RECEIVES_MESSAGE]-()
    	
			SET msgicc.msg_deleted = true

			RETURN true AS done
    `
	}

	res, err := db.Query(
		ctx,
		query,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_id":       msgId,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return false, nil
	}

	done, _, _ := neo4j.GetRecordValue[bool](res.Records[0], "done")

	return done, nil
}
