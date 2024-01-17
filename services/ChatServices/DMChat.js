import * as ChatModel from "../../models/ChatModel.js"
import { getDBClient } from "../../models/db.js"

export class DMChat {
  constructor(conversation_id) {
    this.conversation_id = conversation_id
  }

  async createDM(participantsUserIds) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query('BEGIN')
      
      const conversation_id = await ChatModel.createConversation({ type: "direct" })
  
      await ChatModel.createUserConversation({ participantsUserIds, conversation_id })

      dbClient.query('COMMIT')

      return conversation_id
    } catch (error) {
      dbClient.query('ROLLBACK')
      throw error
    } finally {
      dbClient.release()
    }

  }

  sendMessage() {

  }

  getMessages() {

  }

  reactToMessage() {

  }
}
