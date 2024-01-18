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
  
}
