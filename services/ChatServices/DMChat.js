import * as ChatModel from "../../models/ChatModel.js"
import { getDBClient } from "../../models/db.js"

export class DMChat {
  /**
   *
   * @param {object} client_user
   * @param {number} client_user.user_id
   * @param {string} client_user.username
   * @param {object} partner_user
   * @param {number} partner_user.user_id
   * @param {string} partner_user.username
   * @returns The data needed to display the DM chat page for the client
   */
  async createDM(client_user, partner_user) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      const conversation_id = await ChatModel.createConversation(
        { type: "direct", created_by: client_user.username },
        dbClient
      )

      await ChatModel.createUserConversation(
        {
          participantsUserIds: [client_user.user_id, partner_user.user_id],
          conversation_id,
        },
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
