import { getDBClient } from "../../models/db.js"
import * as ChatModel from "../../models/ChatModel.js"

export class GroupChat {
  /**
   * @param {object[]} participants
   * @param {number} participants.user_id
   * @param {string} participants.username
   */
  async createGroup(client_username, participants) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      const group_conversation_id = await ChatModel.createConversation(
        { type: "group", created_by: client_username },
        dbClient
      )

      // add participants to group
      // group membership will be crated by a TRIGGER, starting with the first user as "admin"
      await ChatModel.createUserConversation(
        {
          participantsUserIds: participants.map(({ user_id }) => user_id),
          conversation_id: group_conversation_id,
        },
        dbClient
      )

      await ChatModel.createGroupConversationActivityLog({
        group_conversation_id,
        activity_info: {
          type: "participants_added",
          added_by: client_username,
          added_participants: participants.map(({ username }) => username),
        },
      }, dbClient)

      dbClient.query("COMMIT")

      return group_conversation_id
    } catch (error) {
      dbClient.query("ROLLBACK")
      throw error
    } finally {
      dbClient.release()
    }
  }

  sendMessage() {}

  joinGroup() {}

  addParticipants() {}

  leaveGroup() {}

  removeMember() {}

  changeGroupPhoto() {}

  changeGroupDescription() {}

  makeAdmin() {}

  removeFromAdmin() {}

  getMessages() {}
}
