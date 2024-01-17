import * as ChatModel from "../../models/ChatModel.js"
import { getDBClient } from "../../models/db.js"

export class DMChat {
  constructor(client_user_id, conversation_id) {
    this.client_user_id = client_user_id
    this.conversation_id = conversation_id
  }

  async createDM(participantsUserIds) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      const conversation_id = await ChatModel.createConversation(
        { type: "direct" },
        dbClient
      )

      await ChatModel.createUserConversation(
        { participantsUserIds, conversation_id },
        dbClient
      )

      dbClient.query("COMMIT")

      return conversation_id
    } catch (error) {
      dbClient.query("ROLLBACK")
      throw error
    } finally {
      dbClient.release()
    }
  }

  async sendMessage(msg_content) {
    await ChatModel.createMessage({
      sender_id: this.client_user_id,
      conversation_id: this.conversation_id,
      msg_content,
    })

    // Implement realtime todos where appropriate
  }

  async getConversationHistory(limit, offset) {
    return await ChatModel.getConversationHistory({
      conversation_id: this.conversation_id,
      limit,
      offset,
    })
  }

  reactToMessage() {}

  deleteMessage() {}
}
