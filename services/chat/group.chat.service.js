import * as ChatModel from "../../models/ChatModel.js"
import { ChatRealtimeService } from "../realtime/chat.realtime.service.js"

export class GroupChatService {
  /**
   * @param {object} param0
   * @param {object[]} param0.participants
   * @param {number} param0.participants.user_id
   * @param {string} param0.participants.username
   * @returns The data needed to display the group chat page history
   */
  async createGroupConversation({
    participants,
    client_username,
    title,
    description,
    cover_image_url,
  }) {
    const group_conversation_id = await ChatModel.createGroupConversation({
      conversationInfo: {
        type: "group",
        created_by: client_username,
        title,
        description,
        cover_image_url,
      },
      participantsUserIds: participants.map(({ user_id }) => user_id),
      activity_info: {
        group_created: {
          type: "group_created",
          created_by: client_username,
        },
        part_added: {
          type: "participants_added",
          added_by: client_username,
          added_participants: participants.map(({ username }) => username),
        },
      },
    })

    ChatRealtimeService.addParticipantsToGroup(
      participants.map(({ user_id }) => user_id),
      group_conversation_id
    )

    return group_conversation_id
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

    await ChatModel.addParticipantsToGroup({
      client_user_id: client.user_id,
      participantsUserIds: participants.map(({ user_id }) => user_id),
      activity_info,
    })

    /* Realtime action */
    // add new participants to group
    ChatRealtimeService.addParticipantsToGroup(
      participants.map(({ user_id }) => user_id),
      group_conversation_id
    )

    // send activity log to participants
    new ChatRealtimeService().sendGroupActivityLog(group_conversation_id, {
      group_conversation_id,
      activity_info,
    })
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
    const activity_info = {
      type: "participant_removed",
      removed_by: client.username,
      removed_participant: participant.username,
    }

    await ChatModel.removeParticipantFromGroup({
      client_user_id: client.user_id,
      participant_user_id: participant.user_id,
      group_conversation_id,
      activity_info,
    })

    /* Realtime action */
    // send activity log to participants
    new ChatRealtimeService().sendGroupActivityLog(group_conversation_id, {
      group_conversation_id,
      activity_info,
    })
  }

  /**
   * @param {object} participant
   * @param {number} participant.user_id
   * @param {string} participant.username
   * @param {number} group_conversation_id
   */
  async joinGroup(participant, group_conversation_id) {
    const activity_info = {
      type: "group_joined",
      who_joined: participant.username,
    }

    await ChatModel.joinGroup({
      participant_user_id: participant.user_id,
      group_conversation_id,
      activity_info,
    })

    /* Realtime action */
    // send activity log to participants
    new ChatRealtimeService().sendGroupActivityLog(group_conversation_id, {
      group_conversation_id,
      activity_info: {
        type: "group_joined",
        who_joined: participant.username,
      },
    })
  }

  /**
   * @param {object} participant
   * @param {number} participant.user_id
   * @param {string} participant.username
   * @param {number} group_conversation_id
   */
  async leaveGroup(participant, group_conversation_id) {
    const activity_info = {
      type: "group_left",
      who_left: participant.username,
    }

    await ChatModel.leaveGroup({
      participant_user_id: participant.user_id,
      group_conversation_id,
      activity_info,
    })

    /* Realtime action */
    // send activity log to participants
    new ChatRealtimeService().sendGroupActivityLog(group_conversation_id, {
      group_conversation_id,
      activity_info,
    })
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
  async makeParticipantAdmin({ client, participant, group_conversation_id }) {
    const activity_info = {
      type: "participant_made_admin",
      made_by: client.username,
      new_admin: participant.username,
    }

    await ChatModel.changeGroupParticipantRole({
      client_user_id: client.user_id,
      participant_user_id: participant.user_id,
      group_conversation_id,
      activity_info,
      role: "admin",
    })

    // Implement realtime todos where appropriate
    new ChatRealtimeService().sendGroupActivityLog(group_conversation_id, {
      group_conversation_id,
      activity_info,
    })
  }

  async removeParticipantFromAdmins({
    client,
    admin_participant,
    group_conversation_id,
  }) {
    const activity_info = {
      type: "admin_removed_from_admins",
      removed_by: client.username,
      ex_admin: admin_participant.username,
    }

    await ChatModel.changeGroupParticipantRole({
      client_user_id: client.user_id,
      participant_user_id: admin_participant.user_id,
      group_conversation_id,
      activity_info,
      role: "member",
    })
    // Implement realtime todos where appropriate
    new ChatRealtimeService().sendGroupActivityLog(group_conversation_id, {
      group_conversation_id,
      activity_info,
    })
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

    await ChatModel.changeGroupInfo({
      client_user_id: client.user_id,
      group_conversation_id,
      newInfoKVPair,
      activity_info,
    })

    new ChatRealtimeService().sendGroupActivityLog(group_conversation_id, {
      group_conversation_id,
      activity_info,
    })
  }

  changeGroupPhoto() {}
}
