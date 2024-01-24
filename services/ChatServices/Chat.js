import * as ChatModel from "../../models/ChatModel.js"
import { ChatRealtimeService } from "./ChatRealtimeService.js"

export class ChatService {
  async getUsersForChat(client_user_id, searchTerm) {
    return await ChatModel.getUsersForChat(client_user_id, searchTerm)
  }

  async getMyConversations(client_user_id) {
    return await ChatModel.getAllUserConversations(client_user_id)
  }

  async deleteConversation(client_user_id, conversation_id) {
    await ChatModel.updateUserConversation({
      user_id: client_user_id,
      conversation_id,
      updateKVPairs: new Map().set("deleted", true),
    })
  }

  async getConversationHistory({ conversation_id, limit, offset }) {
    return await ChatModel.getConversationHistory({
      conversation_id,
      limit,
      offset,
    })
  }

  /**
   * @param {object} param0
   * @param {Object} param0.sender
   * @param {number} param0.sender.user_id
   * @param {string} param0.sender.username
   */
  async sendMessage({ sender, conversation_id, msg_content }) {
    const newMessageData = await ChatModel.createMessage({
      sender_user_id: sender.user_id,
      conversation_id,
      msg_content,
    })

    new ChatRealtimeService().sendNewMessage(conversation_id, newMessageData)
  }

  async acknowledgeMessageDelivered({ user_id, conversation_id, message_id }) {
    const isDelivered = await ChatModel.acknowledgeMessageDelivered(
      user_id,
      message_id
    )

    if (isDelivered) {
      /* Realtime actions */
      // change message delivery_status to delivered for other participants
      new ChatRealtimeService().sendMessageDelivered(conversation_id, {
        conversation_id,
        message_id,
      })
    }
  }

  async acknowledgeMessageRead({ user_id, conversation_id, message_id }) {
    const isRead = await ChatModel.acknowledgeMessageRead(user_id, message_id)

    if (isRead) {
      /* Realtime actions */
      // change message delivery_status to read for other participants
      new ChatRealtimeService().sendMessageRead(conversation_id, {
        conversation_id,
        message_id,
      })
    }
  }

  /**
   * @param {object} param0
   * @param {Object} param0.reactor
   * @param {number} param0.reactor.user_id
   * @param {string} param0.reactor.username
   */
  async reactToMessage({
    reactor,
    conversation_id,
    message_id,
    reaction_code_point,
  }) {
    await ChatModel.createMessageReaction({
      reactor_user_id: reactor.user_id,
      message_id,
      reaction_code_point,
    })

    /* Realtime actions */
    // update message reaction for other participants
    new ChatRealtimeService().sendMessageReaction(conversation_id, {
      reactor,
      message_id,
      reaction_code_point,
    })
  }

  /**
   * @param {object} param0
   * @param {Object} param0.reactor
   * @param {number} param0.reactor.user_id
   * @param {string} param0.reactor.username
   */
  async removeReactionToMessage({ reactor, conversation_id, message_id }) {
    await ChatModel.deleteMessageReaction({
      reactor_user_id: reactor.user_id,
      message_id,
    })

    /* Realtime actions */
    // remove message reaction for other participants
    new ChatRealtimeService().sendMessageReactionRemoval(conversation_id, {
      reactor,
      message_id,
    })
  }

  /**
   * @param {object} param0
   * @param {Object} param0.deleter
   * @param {number} param0.deleter.user_id
   * @param {string} param0.deleter.username
   */
  async deleteMessage({ deleter, conversation_id, message_id, deleted_for }) {
    await ChatModel.createMessageDeletionLog({
      deleter_user_id: deleter.user_id,
      message_id,
      deleted_for,
    })

    if (deleted_for === "everyone") {
      /* Realtime actions */
      // delete message for other participants
      new ChatRealtimeService().sendMessageDeleted(conversation_id, {
        deleter_username: deleter.username,
        message_id,
      })
    }
  }
}
