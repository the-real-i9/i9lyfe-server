import * as ChatModel from "../../models/ChatModel.js"
import { getDBClient } from "../../models/db.js"

export class DMChat {
  /**
   * 
   * @param {number[]} participantsUserIds The two individual ids
   * @returns The data needed to display the DM chat page for the client
   */
  async createDM(client_username, participantsUserIds) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      const conversation_id = await ChatModel.createConversation(
        { type: "direct", created_by: client_username},
        dbClient
      )

      await ChatModel.createUserConversation(
        { participantsUserIds, conversation_id },
        dbClient
      )

      dbClient.query("COMMIT")

      return conversation_id // more
    } catch (error) {
      dbClient.query("ROLLBACK")
      throw error
    } finally {
      dbClient.release()
    }
  }

  async sendMessage(conversation_id, sender_id, msg_content) {
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
