import { Conversation, Message } from "../../models/chat.model.js"
import { uploadMessageMediaData } from "../mediaUploader.service.js"
import { ChatRealtimeService } from "../realtime/chat.realtime.service.js"

export class ChatService {
  /**
   * @param {object} client
   * @param {number} client.user_id
   * @param {string} client.username
   * @param {object} partner
   * @param {number} partner.user_id
   * @param {string} partner.username
   * @returns The data needed to display the DM chat page for the client
   */
  static async createConversation(client, partner, init_msg) {
    const { media_data, ...init_message } = init_msg

    init_message.media_url = await uploadMessageMediaData(media_data)

    const { client_res, partner_res } = await Conversation.create(
      client,
      partner.user_id,
      init_message
    )

    ChatRealtimeService.send("new conversation", partner.user_id, partner_res)

    return client_res
  }

  static async getMyConversations(client_user_id) {
    return await Conversation.getAll(client_user_id)
  }

  static async deleteMyConversation(client_user_id, conversation_id) {
    await Conversation.delete(client_user_id, conversation_id)
  }

  static async getConversationHistory({ conversation_id, limit, offset }) {
    return await Conversation.getHistory({
      conversation_id,
      limit,
      offset,
    })
  }

  static async sendMessage({
    client_user_id,
    partner_user_id,
    conversation_id,
    msg_content,
  }) {
    const { client_res, partner_res } = await Conversation.sendMessage({
      client_user_id,
      conversation_id,
      msg_content,
    })

    ChatRealtimeService.send("new message", partner_user_id, partner_res)

    return client_res
  }

  static async acknowledgeMessageDelivered({
    client_user_id,
    partner_user_id,
    conversation_id,
    message_id,
    delivery_time,
  }) {
    await Message.isDelivered({
      client_user_id,
      conversation_id,
      message_id,
      delivery_time,
    })

    ChatRealtimeService.send("message delivered", partner_user_id, {
      conversation_id,
      message_id,
    })
  }

  static async acknowledgeMessageRead({
    conversation_id,
    client_user_id,
    partner_user_id,
    message_id,
  }) {
    await Message.isRead({
      client_user_id,
      conversation_id,
      message_id,
    })

    ChatRealtimeService.send("message read", partner_user_id, {
      conversation_id,
      message_id,
    })
  }

  /**
   * @param {object} param0
   * @param {Object} param0.reactor
   * @param {number} param0.reactor.user_id
   * @param {string} param0.reactor.username
   */
  static async reactToMessage({
    conversation_id,
    reactor,
    partner_user_id,
    message_id,
    reaction_code_point,
  }) {
    await Message.reactTo({
      reactor_user_id: reactor.user_id,
      message_id,
      reaction_code_point,
    })

    ChatRealtimeService.send("message reaction", partner_user_id, {
      conversation_id,
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
  static async removeReactionToMessage({
    conversation_id,
    reactor,
    partner_user_id,
    message_id,
  }) {
    await Message.removeReaction(message_id, reactor.user_id)

    /* Realtime actions */
    // remove message reaction for other participants
    ChatRealtimeService.send("message reaction removed", partner_user_id, {
      reactor,
      conversation_id,
      message_id,
    })
  }

  /**
   * @param {object} param0
   * @param {Object} param0.deleter
   * @param {number} param0.deleter.user_id
   * @param {string} param0.deleter.username
   */
  static async deleteMessage({
    conversation_id,
    deleter,
    partner_user_id,
    message_id,
    delete_for,
  }) {
    await Message.delete({
      deleter_user_id: deleter.user_id,
      message_id,
      deleted_for: delete_for,
    })

    if (delete_for === "everyone") {
      ChatRealtimeService.send("message deleted", partner_user_id, {
        conversation_id,
        deleter_username: deleter.username,
        message_id,
      })
    }
  }
}
