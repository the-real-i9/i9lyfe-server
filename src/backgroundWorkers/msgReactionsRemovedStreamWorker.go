package backgroundWorkers

import (
	"context"
	"i9lyfe/src/cache"
	"i9lyfe/src/helpers"
	"i9lyfe/src/services/eventStreamService/eventTypes"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

func msgReactionsRemovedStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "msg_reactions_removed"
		groupName    = "msg_reaction_removed_listeners"
		consumerName = "worker-1"
	)

	ctx := context.Background()

	err := rdb.XGroupCreateMkStream(ctx, streamName, groupName, "$").Err()
	if err != nil && (err.Error() != "BUSYGROUP Consumer Group name already exists") {
		helpers.LogError(err)
		log.Fatal()
	}

	go func() {
		for {
			streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    groupName,
				Consumer: consumerName,
				Streams:  []string{streamName, ">"},
				Count:    500,
				Block:    0,
			}).Result()

			if err != nil {
				helpers.LogError(err)
				continue
			}

			var stmsgIds []string
			var msgs []eventTypes.MsgReactionRemovedEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.MsgReactionRemovedEvent

				msg.FromUser = stmsg.Values["fromUser"].(string)
				msg.ToUser = stmsg.Values["toUser"].(string)
				msg.ToMsgId = stmsg.Values["toMsgId"].(string)
				msg.CHEId = stmsg.Values["CHEId"].(string)

				msgs = append(msgs, msg)

			}

			msgReactionEntriesRemoved := []string{}

			chatMsgReactionsRemoved := make(map[[2]string][]any)

			msgReactionsRemoved := make(map[string][]string)

			// batch data for batch processing
			for _, msg := range msgs {
				msgReactionEntriesRemoved = append(msgReactionEntriesRemoved, msg.CHEId)

				chatMsgReactionsRemoved[[2]string{msg.FromUser, msg.ToUser}] = append(chatMsgReactionsRemoved[[2]string{msg.FromUser, msg.ToUser}], msg.CHEId)

				msgReactionsRemoved[msg.ToMsgId] = append(msgReactionsRemoved[msg.ToMsgId], msg.FromUser)
			}

			// batch processing
			if err := cache.RemoveChatHistoryEntries(ctx, msgReactionEntriesRemoved); err != nil {
				return
			}

			eg, sharedCtx := errgroup.WithContext(ctx)

			for ownerUserPartnerUser, CHEIds := range chatMsgReactionsRemoved {
				eg.Go(func() error {
					ownerUserPartnerUser, CHEIds := ownerUserPartnerUser, CHEIds

					return cache.RemoveUserChatHistory(sharedCtx, ownerUserPartnerUser, CHEIds)
				})
			}

			for msgId, reactorUsers := range msgReactionsRemoved {
				eg.Go(func() error {
					msgId, reactorUsers := msgId, reactorUsers

					return cache.RemoveMsgReactions(sharedCtx, msgId, reactorUsers)
				})
			}

			if eg.Wait() != nil {
				return
			}

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
