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
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
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
		return nil, fiber.ErrInternalServerError
	}

	return mychats, nil
}

func Delete(ctx context.Context, clientUsername, partnerUsername string) error {
	return nil
}

func History(ctx context.Context, clientUsername, partnerUsername string, limit int, cursor float64) (chatHistory []UITypes.ChatHistoryEntry, err error) {

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
			"offset":           cursor,
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
