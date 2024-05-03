import * as ChatModel from "../../models/ChatModel.js"
import { ChatRealtimeService } from "../RealtimeServices/ChatRealtimeService.js"

export class ChatService {
  async getUsersToChat(client_user_id, search) {
    return await ChatModel.getUsersToChat(client_user_id, search)
  }

  async getMyConversations(client_user_id) {
    return await ChatModel.getAllUserConversations(client_user_id)
  }

  async getConversation(conversation_id, client_user_id) {
    return await ChatModel.getConversation(
      conversation_id,
      client_user_id
    )
  }

  async deleteMyConversation(client_user_id, conversation_id) {
    await ChatModel.deleteUserConversation(client_user_id, conversation_id)
  }

  async getConversationHistory({ conversation_id, limit, offset }) {
    return await ChatModel.getConversationHistory({
      conversation_id,
      limit,
      offset,
    })
  }

  async sendMessage({ sender_user_id, conversation_id, msg_content }) {
    const newMessageData = await ChatModel.createMessage({
      sender_user_id,
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
  async removeMyReactionToMessage({ reactor, conversation_id, message_id }) {
    await ChatModel.deleteMessageReaction(message_id, reactor.user_id)

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
  async deleteMessage({ deleter, conversation_id, message_id, delete_for }) {
    await ChatModel.createMessageDeletionLog({
      deleter_user_id: deleter.user_id,
      message_id,
      deleted_for: delete_for,
    })

    if (delete_for === "everyone") {
      /* Realtime actions */
      // delete message for other participants
      new ChatRealtimeService().sendMessageDeleted(conversation_id, {
        deleter_username: deleter.username,
        message_id,
      })
    }
  }
}
