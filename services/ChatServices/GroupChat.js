import { getDBClient } from "../../models/db.js"
import * as ChatModel from "../../models/ChatModel.js"

export class GroupChat {
  async #addParticipantsToGroup(
    { client_username, participants, group_conversation_id },
    dbClient
  ) {
    await ChatModel.createUserConversation(
      {
        participantsUserIds: participants.map(({ user_id }) => user_id),
        conversation_id: group_conversation_id,
      },
      dbClient
    )

    await ChatModel.createGroupConversationActivityLog(
      {
        group_conversation_id,
        activity_info: {
          type: "participants_added",
          added_by: client_username,
          added_participants: participants.map(({ username }) => username),
        },
      },
      dbClient
    )
  }

  /**
   * @param {object[]} participants
   * @param {number} participants.user_id
   * @param {string} participants.username
   * @returns The data needed to display the group chat page history
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
      await this.#addParticipantsToGroup(
        { client_username, participants, group_conversation_id },
        dbClient
      )

      dbClient.query("COMMIT")

      // Implement realtime todos where appropriate

      return group_conversation_id // more
    } catch (error) {
      dbClient.query("ROLLBACK")
      throw error
    } finally {
      dbClient.release()
    }
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

  /**
   * @param {object} participant
   * @param {number} participant.user_id
   * @param {string} participant.username
   * @param {number} group_conversation_id
   */
  async joinGroup(participant, group_conversation_id) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      await ChatModel.createUserConversation(
        {
          participantsUserIds: [participant.user_id],
          conversation_id: group_conversation_id,
        },
        dbClient
      )

      await ChatModel.createGroupConversationActivityLog(
        {
          group_conversation_id,
          activity_info: {
            type: "group_joined",
            who_joined: participant.username,
          },
        },
        dbClient
      )

      dbClient.query("COMMIT")

      // Implement realtime todos where appropriate
    } catch (error) {
      dbClient.query("ROLLBACK")
      throw error
    } finally {
      dbClient.release()
    }
  }

  /**
   * @param {object} param0
   * @param {object} param0.client_user
   * @param {number} param0.client_user.user_id
   * @param {string} param0.client_user.username
   * @param {object[]} param0.participants
   * @param {number} param0.participants.user_id
   * @param {string} param0.participants.username
   * @param {number} param0.group_conversation_id
   */
  async addParticipants({ client_user, participants, group_conversation_id }) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      if (
        !(await ChatModel.isGroupAdmin({
          member_user_id: client_user.user_id,
          group_conversation_id,
        }, dbClient))
      ) {
        throw new Error("You must be a group admin to add participants!")
      }

      // add participants to group
      // group membership will be crated by a TRIGGER, starting with the first user as "admin"
      await this.#addParticipantsToGroup(
        {
          client_username: client_user.username,
          participants,
          group_conversation_id,
        },
        dbClient
      )

      dbClient.query("COMMIT")

      // Implement realtime todos where appropriate
    } catch (error) {
      dbClient.query("ROLLBACK")
      throw error
    } finally {
      dbClient.release()
    }
  }

  /**
   * @param {object} participant
   * @param {number} participant.user_id
   * @param {string} participant.username
   * @param {number} group_conversation_id
   */
  async leaveGroup(participant, group_conversation_id) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      await ChatModel.updateUserConversation({
        user_id: participant.user_id,
        conversation_id: group_conversation_id,
        updateKVPairs: new Map().set("deleted", true),
      })

      await ChatModel.createGroupConversationActivityLog(
        {
          group_conversation_id,
          activity_info: {
            type: "group_left",
            who_left: participant.username,
          },
        },
        dbClient
      )

      dbClient.query("COMMIT")

      // Implement realtime todos where appropriate
    } catch (error) {
      dbClient.query("ROLLBACK")
      throw error
    } finally {
      dbClient.release()
    }
  }

  removeParticipant() {}

  changeGroupPhoto() {}

  changeGroupDescription() {}

  makeAdmin() {}

  removeFromAdmin() {}
}
