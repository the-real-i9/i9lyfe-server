import * as ChatModel from "../../models/ChatModel.js"
import { ChatRealtimeService } from "../RealtimeServices/ChatRealtimeService.js"

export class DMChatService {
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
  async createDMConversation(client, partner) {
      const dm_conversation_id = await ChatModel.createDMConversation(client, partner.user_id)

      await ChatRealtimeService.createDMConversation({
        client_user_id: client.user_id,
        partner_user_id: partner.user_id,
        dm_conversation_id,
      })

      return dm_conversation_id
  }
}
