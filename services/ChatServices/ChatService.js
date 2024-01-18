import * as ChatModel from "../../models/ChatModel.js"

export class ChatService {
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

  async sendMessage({ sender_id, conversation_id, msg_content }) {
    await ChatModel.createMessage({ sender_id, conversation_id, msg_content })

    // Implement realtime todos where appropriate
  }

  async getConversationHistory({ conversation_id, limit, offset }) {
    return await ChatModel.getConversationHistory({
      conversation_id,
      limit,
      offset,
    })
  }

  async reactToMessage({ reactor_user_id, message_id, reaction_code_point }) {
    await ChatModel.createMessageReaction({
      reactor_user_id,
      message_id,
      reaction_code_point,
    })

    // Implement realtime todos where appropriate
  }

  async deleteMessage({ message_id, deleted_by_user_id, deleted_for }) {
    await ChatModel.createMessageDeletionLog({
      message_id,
      deleted_by_user_id,
      deleted_for,
    })

    // Implement realtime todos where appropriate
  }
}
