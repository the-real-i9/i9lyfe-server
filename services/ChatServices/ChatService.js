import * as ChatModel from "../../models/ChatModel.js"

export class ChatService {
  constructor(client_user_id) {
    this.client_user_id = client_user_id
  }

  async getMyConversations() {
    return await ChatModel.getAllUserConversations(this.client_user_id)
  }

  async deleteConversation(conversation_id) {
    await ChatModel.updateUserConversation({
      user_id: this.client_user_id,
      conversation_id,
      updateKVPairs: new Map().set("deleted", true),
    })
  }
}
