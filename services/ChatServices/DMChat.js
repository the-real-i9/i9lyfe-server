import * as ChatModel from "../../models/ChatModel.js"
import { getDBClient } from "../../models/db.js"
import { ChatRealtimeService } from "./ChatRealtimeService.js"

export class DMChat {
  /**
   *
   * @param {object} client
   * @param {number} client.user_id
   * @param {string} client.username
   * @param {object} partner
   * @param {number} partner.user_id
   * @param {string} partner.username
   * @returns The data needed to display the DM chat page for the client
   */
  async createDM(client, partner) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      const conversation_id = await ChatModel.createConversation(
        { type: "direct", created_by: client.username },
        dbClient
      )

      await ChatModel.createUserConversation(
        {
          participantsUserIds: [client.user_id, partner.user_id],
          conversation_id,
        },
        dbClient
      )

      dbClient.query("COMMIT")

      ChatRealtimeService.createDMConversation({
        client_user_id: client.user_id,
        partner_user_id: partner.user_id,
        conversation_id,
      })

      return conversation_id // more
    } catch (error) {
      dbClient.query("ROLLBACK")
      throw error
    } finally {
      dbClient.release()
    }
  }
}
