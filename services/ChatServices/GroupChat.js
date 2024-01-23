import { getDBClient } from "../../models/db.js"
import * as ChatModel from "../../models/ChatModel.js"
import { ChatRealtimeService } from "./ChatRealtimeService.js"

export class GroupChatService {
  /**
   * @param {object} param0
   * @param {object[]} param0.participants
   * @param {number} param0.participants.user_id
   * @param {string} param0.participants.username
   * @returns The data needed to display the group chat page history
   */
  async createGroup({
    participants,
    client_username,
    title,
    description,
    cover_image_url,
  }) {
    /* Nothing returned yet */
    const groupConversationData = await ChatModel.createGroupConversation({
      conversationInfo: {
        type: "group",
        created_by: client_username,
        title,
        description,
        cover_image_url,
      },
      participantsUserIds: participants.map(({ user_id }) => user_id),
      activity_info: {
        type: "participants_added",
        added_by: client_username,
        added_participants: participants.map(({ username }) => username),
      },
    })

    const { group_conversation_id } = groupConversationData

    // Implement realtime todos where appropriate
    // here we have to create a socket room for this conversation and add these participants to it
    await ChatRealtimeService.createGroupConversation(
      participants.map(({ user_id }) => user_id),
      group_conversation_id
    )

    // next we send the group creation and additon of participants to all participants
    new ChatRealtimeService().sendNewGroupConversation(
      group_conversation_id,
      groupConversationData
    )
  }

  /**
   * @param {object} param0
   * @param {object} param0.client
   * @param {number} param0.client.user_id
   * @param {string} param0.client.username
   * @param {object[]} param0.participants
   * @param {number} param0.participants.user_id
   * @param {string} param0.participants.username
   * @param {number} param0.group_conversation_id
   */
  async addParticipants({ client, participants, group_conversation_id }) {
    const activity_info = {
      type: "participants_added",
      added_by: client.username,
      added_participants: participants.map(({ username }) => username),
    }

    const participantsAdded = await ChatModel.addParticipantsToGroup({
      client_user_id: client.user_id,
      participantsUserIds: participants.map(({ user_id }) => user_id),
      activity_info,
    })


    /* Realtime action */
    // send activity log to participants
    if (participantsAdded) {
      new ChatRealtimeService().sendGroupActivityLog(group_conversation_id, {
        group_conversation_id,
        activity_info,
      })
    }
  }

  /**
   * @param {object} param0
   * @param {object} param0.client
   * @param {number} param0.client.user_id
   * @param {string} param0.client.username
   * @param {object} param0.participant
   * @param {number} param0.participant.user_id
   * @param {string} param0.participant.username
   * @param {number} param0.group_conversation_id
   */
  async removeParticipant({ client, participant, group_conversation_id }) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      if (
        !(await ChatModel.isGroupAdmin(
          {
            participant_user_id: client.user_id,
            group_conversation_id,
          },
          dbClient
        ))
      ) {
        throw new Error("You must be a group admin to remove participant!")
      }

      await ChatModel.updateUserConversation({
        user_id: participant.user_id,
        conversation_id: group_conversation_id,
        updateKVPairs: new Map().set("deleted", true),
      })

      await ChatModel.createGroupConversationActivityLog(
        {
          group_conversation_id,
          activity_info: {
            type: "participant_removed",
            removed_by: client.username,
            removed_participant: participant.username,
          },
        },
        dbClient
      )

      dbClient.query("COMMIT")

      /* Realtime action */
      // send activity log to participants
      new ChatRealtimeService().sendGroupActivityLog(group_conversation_id, {
        group_conversation_id,
        activity_info: {
          type: "participant_removed",
          removed_by: client.username,
          removed_participant: participant.username,
        },
      })
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

      /* Realtime action */
      // send activity log to participants
      new ChatRealtimeService().sendGroupActivityLog(group_conversation_id, {
        group_conversation_id,
        activity_info: {
          type: "group_joined",
          who_joined: participant.username,
        },
      })
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

      await ChatModel.updateUserConversation(
        {
          user_id: participant.user_id,
          conversation_id: group_conversation_id,
          updateKVPairs: new Map().set("deleted", true),
        },
        dbClient
      )

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

      /* Realtime action */
      // send activity log to participants
      new ChatRealtimeService().sendGroupActivityLog(group_conversation_id, {
        group_conversation_id,
        activity_info: {
          type: "group_left",
          who_left: participant.username,
        },
      })
    } catch (error) {
      dbClient.query("ROLLBACK")
      throw error
    } finally {
      dbClient.release()
    }
  }

  /**
   * @param {object} param0
   * @param {object} param0.client
   * @param {number} param0.client.user_id
   * @param {string} param0.client.username
   * @param {object} param0.participant
   * @param {number} param0.participant.user_id
   * @param {string} param0.participant.username
   * @param {number} param0.group_conversation_id
   */
  async makeAdmin({ client, participant, group_conversation_id }) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      if (
        !(await ChatModel.isGroupAdmin(
          {
            participant_user_id: client.user_id,
            group_conversation_id,
          },
          dbClient
        ))
      ) {
        throw new Error(
          "You must be a group admin to make participant an admin!"
        )
      }

      await ChatModel.updateGroupMembership(
        {
          participant_user_id: participant.user_id,
          group_conversation_id,
          role: "admin",
        },
        dbClient
      )

      await ChatModel.createGroupConversationActivityLog(
        {
          group_conversation_id,
          activity_info: {
            type: "participant_made_admin",
            made_by: client.username,
            new_admin: participant.username,
          },
        },
        dbClient
      )

      dbClient.query("COMMIT")

      // Implement realtime todos where appropriate
      new ChatRealtimeService().sendGroupActivityLog(group_conversation_id, {
        group_conversation_id,
        activity_info: {
          type: "participant_made_admin",
          made_by: client.username,
          new_admin: participant.username,
        },
      })
    } catch (error) {
      dbClient.query("ROLLBACK")
      throw error
    } finally {
      dbClient.release()
    }
  }

  async dropFromAdmin({ client, admin_participant, group_conversation_id }) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      if (
        !(await ChatModel.isGroupAdmin(
          {
            participant_user_id: client.user_id,
            group_conversation_id,
          },
          dbClient
        ))
      ) {
        throw new Error("You must be a group admin to drop admin to member!")
      }

      await ChatModel.updateGroupMembership(
        {
          participant_user_id: admin_participant.user_id,
          group_conversation_id,
          role: "member",
        },
        dbClient
      )

      await ChatModel.createGroupConversationActivityLog(
        {
          group_conversation_id,
          activity_info: {
            type: "admin_dropped_from_admins",
            dropped_by: client.username,
            ex_admin: admin_participant.username,
          },
        },
        dbClient
      )

      dbClient.query("COMMIT")

      // Implement realtime todos where appropriate
      new ChatRealtimeService().sendGroupActivityLog(group_conversation_id, {
        group_conversation_id,
        activity_info: {
          type: "admin_dropped_from_admins",
          dropped_by: client.username,
          ex_admin: admin_participant.username,
        },
      })
    } catch (error) {
      dbClient.query("ROLLBACK")
      throw error
    } finally {
      dbClient.release()
    }
  }

  /**
   * @param {object} param0
   * @param {object} param0.client
   * @param {number} param0.client.user_id
   * @param {string} param0.client.username
   * @param {Object<string, string>} param0.newInfoKVPair
   */
  async changeGroupInfo({ client, group_conversation_id, newInfoKVPair }) {
    const [[infoKey, newInfoValue]] = Object.entries(newInfoKVPair)

    const activity_info = {
      type: `group_${infoKey}_changed`,
      changed_by: client.username,
      [`new_group_${infoKey}`]: newInfoValue,
    }

    const changed = await ChatModel.changeGroupInfo({
      client_user_id: client.user_id,
      group_conversation_id,
      newInfoKVPair,
      activity_info,
    })

    if (changed) {
      new ChatRealtimeService().sendGroupActivityLog(group_conversation_id, {
        group_conversation_id,
        activity_info,
      })
    }
  }

  changeGroupPhoto() {}
}
