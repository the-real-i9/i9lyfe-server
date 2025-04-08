package chatModel

import (
	"context"
	"i9lyfe/src/models/db"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func All(ctx context.Context, clientUsername string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (clientChat:Chat{ owner_username: $client_username } WHERE clientChat.updated_at < $offset)-[:WITH_USER]->(partnerUser),
			(clientChat)<-[:IN_CHAT]-(lmsg:Message WHERE lmsg.id = clientChat.last_message_id),
			(clientChat)<-[:IN_CHAT]-(:Message)<-[lrxn:REACTS_TO_MESSAGE WHERE lrxn.at = clientChat.last_reaction_at]-(reactor)
		WITH clientChat, toString(clientChat.updated_at) AS updated_at, partnerUser { .username, .profile_pic_url, .connection_status } AS partner,
			CASE clientChat.last_activity_type 
				WHEN "message" THEN lmsg { type: "message", .content, .delivery_status }
				WHEN "reaction" THEN lrxn { type: "reaction", .reaction, reactor: reactor.username }
			END AS last_activity
		ORDER BY clientChat.updated_at DESC
		LIMIT $limit
		RETURN collect(clientChat { partner, .unread_messages_count, updated_at, last_activity }) AS client_chats
		`,
		map[string]any{
			"client_username": clientUsername,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		log.Println("chatModel.go: All:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	clientChats, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "client_chats")

	return clientChats, nil
}

func Delete(ctx context.Context, clientUsername, partnerUsername string) error {
	_, err := db.Query(
		ctx,
		`
		MATCH (clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username })
		DETACH DELETE clientChat
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
		},
	)
	if err != nil {
		log.Println("chatModel.go: Delete:", err)
		return fiber.ErrInternalServerError
	}

	return nil
}

func History(ctx context.Context, clientUsername, partnerUsername string, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username })<-[:IN_CHAT]-(message:Message WHERE message.created_at < $offset)<-[rxn:REACTS_TO_MESSAGE]-(reactor)
		
		WITH message, toString(message.created_at) AS created_at, collect({ user: reactor { .username, .profile_pic_url }, reaction: rxn.reaction }) AS reactions
		ORDER BY message.created_at DESC
		LIMIT 50
		RETURN collect(message { .*, created_at, reactions }) AS chat_history
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"offset":           offset,
		},
	)
	if err != nil {
		log.Println("chatModel.go: History:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	ch, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "chat_history")

	return ch, nil
}
