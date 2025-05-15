package chatMessageModel

import (
	"context"
	"i9lyfe/src/helpers"
	"i9lyfe/src/models/db"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type SendResT struct {
	ClientRes  map[string]any `json:"client_res"`
	PartnerRes map[string]any `json:"partner_res"`
}

func Send(ctx context.Context, clientUsername, toUser, msgType, msgProps string, msgAt time.Time) (SendResT, error) {
	var resData SendResT

	res, err := db.Query(
		ctx,
		`
		MATCH (clientUser:User{ username: $client_username }), (partnerUser:User{ username: $partner_username })
		MERGE (clientUser)-[:HAS_CHAT]->(clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username })-[:WITH_USER]->(partnerUser)
		MERGE (partnerUser)-[:HAS_CHAT]->(partnerChat:Chat{ owner_username: $partner_username, partner_username: $client_username })-[:WITH_USER]->(clientUser)
		SET clientChat.last_activity_type = "message", 
			partnerChat.last_activity_type = "message",
			clientChat.updated_at = $msg_at, 
			partnerChat.updated_at = $msg_at
		WITH clientUser, clientChat, partnerUser, partnerChat
		CREATE (message:Message{ id: randomUUID(), type: $msg_type, props: $msg_props, delivery_status: "sent", created_at: $msg_at }),
			(clientUser)-[:SENDS_MESSAGE]->(message)-[:IN_CHAT]->(clientChat),
			(partnerUser)-[:RECEIVES_MESSAGE]->(message)-[:IN_CHAT]->(partnerChat)
		SET clientChat.last_message_id = message.id,
			partnerChat.last_message_id = message.id
		WITH message, toString(message.created_at) AS created_at, clientUser { .username, .profile_pic_url, .presence } AS sender
		RETURN { new_msg_id: message.id } AS client_res,
			message { .*, created_at, sender } AS partner_res
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": toUser,
			"msg_type":         msgType,
			"msg_props":        msgProps,
			"msg_at":           msgAt,
		},
	)
	if err != nil {
		log.Println("chatModel.go: SendMessage:", err)
		return resData, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return resData, nil
	}

	helpers.ToStruct(res.Records[0].AsMap(), &resData)

	return resData, nil
}

func AckDelivered(ctx context.Context, clientUsername, partnerUsername, msgId string, at time.Time) (bool, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username }),
			(clientChat)<-[:IN_CHAT]-(message:Message{ id: $message_id, delivery_status: "sent" })<-[:RECEIVES_MESSAGE]-()
		SET message.delivery_status = "delivered", message.delivered_at = $delivered_at, clientChat.unread_messages_count = coalesce(clientChat.unread_messages_count, 0) + 1

		RETURN true AS done
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_id":       msgId,
			"delivered_at":     at,
		},
	)
	if err != nil {
		log.Println("chatModel.go: AckMsgDelivered:", err)
		return false, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return false, nil
	}

	done, _, _ := neo4j.GetRecordValue[bool](res.Records[0], "done")

	return done, nil
}

func AckRead(ctx context.Context, clientUsername, partnerUsername, msgId string, at time.Time) (bool, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username }),
			(clientchat)<-[:IN_CHAT]-(message:Message{ id: $message_id } WHERE message.delivery_status IN ["sent", "delivered"])<-[:RECEIVES_MESSAGE]-()
		WITH clientChat, message, CASE coalesce(clientChat.unread_messages_count, 0) WHEN <> 0 THEN clientChat.unread_messages_count - 1 ELSE 0 END AS unread_messages_count
		SET message.delivery_status = "read", message.read_at = $read_at, clientChat.unread_messages_count = unread_messages_count

		RETURN true AS done
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_id":       msgId,
			"read_at":          at,
		},
	)
	if err != nil {
		log.Println("chatModel.go: AckMsgRead:", err)
		return false, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return false, nil
	}

	done, _, _ := neo4j.GetRecordValue[bool](res.Records[0], "done")

	return done, nil
}

func ReactTo(ctx context.Context, clientUsername, partnerUsername, msgId, reaction string, at time.Time) (bool, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (clientUser)-[:HAS_CHAT]->(clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username })<-[:IN_CHAT]-(message:Message{ id: $message_id }),
			(clientChat)-[:WITH_USER]->(partnerChat)
		MERGE (clientUser)-[crxn:REACTS_TO_MESSAGE]->(message)
		ON CREATE
			SET crxn.reaction = $reaction, 
				crxn.at = $reaction_at,
				clientChat.last_activity_type = "reaction", 
				partnerChat.last_activity_type = "reaction",
				clientChat.last_reaction_at = $reaction_at,
				partnerChat.last_reaction_at = $reaction_at

		RETURN true AS done
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_id":       msgId,
			"reaction":         reaction,
			"reaction_at":      at,
		},
	)
	if err != nil {
		log.Println("chatModel.go: ReactToMsg:", err)
		return false, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return false, nil
	}

	done, _, _ := neo4j.GetRecordValue[bool](res.Records[0], "done")

	return done, nil
}

func RemoveReaction(ctx context.Context, clientUsername, partnerUsername, msgId string) (bool, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (:User{ username: $client_username })-[crxn:REACTS_TO_MESSAGE]->(:Message{ id: $message_id })-[:IN_CHAT]->(:Chat{ owner_username: $client_username, partner_username: $partner_username })
		DELETE rr

		RETURN true AS done
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_id":       msgId,
		},
	)
	if err != nil {
		log.Println("chatModel.go: RemoveReactionToMsg:", err)
		return false, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return false, nil
	}

	done, _, _ := neo4j.GetRecordValue[bool](res.Records[0], "done")

	return done, nil
}

func Delete(ctx context.Context, clientUsername, partnerUsername, msgId, deleteFor string) (bool, error) {
	var query string

	if deleteFor == "me" {
		query = `
			MATCH (clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username })<-[incr:IN_CHAT]-(message:Message{ id: $message_id })<-[rsmr:SENDS_MESSAGE|RECEIVES_MESSAGE]-(clientUser)
    	DELETE incr, rsmr

			RETURN true AS done
		`
	} else {
		query = `
      MATCH (clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username })<-[:IN_CHAT]-(message:Message{ id: $message_id })
      DETACH DELETE message

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
		log.Println("chatModel.go: DeleteMsg:", err)
		return false, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return false, nil
	}

	done, _, _ := neo4j.GetRecordValue[bool](res.Records[0], "done")

	return done, nil
}
