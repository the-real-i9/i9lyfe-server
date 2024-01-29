/**
 * @typedef {import("express").Request} ExpressRequest
 * @typedef {import("express").Response} ExpressResponse
 */

import { ChatService } from "../services/ChatServices/Chat.js"
import { DMChatService } from "../services/ChatServices/DMChat.js"
import { GroupChatService } from "../services/ChatServices/GroupChat.js"

/**
 *
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const getUsersToChatController = async (req, res) => {
  try {
    const { search = "" } = req.query

    const { client_user_id } = req.auth

    const users = await new ChatService().getUsersToChat(client_user_id, search)

    res.status(200).send({ users })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const createDMConversationController = async (req, res) => {
  try {
    const { partner } = req.body

    const { client_user_id, client_username } = req.auth

    const dm_conversation_id = await new DMChatService().createDMConversation(
      { user_id: client_user_id, username: client_username },
      partner
    )

    res.status(201).send({ dm_conversation_id })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const createGroupConversationController = async (req, res) => {
  try {
    const {
      title,
      description = "",
      cover_image_url = "",
      participants,
    } = req.body

    const { client_username } = req.auth

    const groupConversationData =
      await new GroupChatService().createGroupConversation({
        participants,
        client_username,
        title,
        description,
        cover_image_url,
      })

    res.status(201).send({ groupConversationData })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const addParticipantsToGroupController = async (req, res) => {
  try {
    const { participants } = req.body

    const { group_conversation_id } = req.params

    const { client_user_id, client_username } = req.auth

    await new GroupChatService().addParticipants({
      client: {
        user_id: client_user_id,
        username: client_username,
      },
      participants,
      group_conversation_id,
    })

    res.status(201)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const removeParticipantFromGroupController = async (req, res) => {
  try {
    const { participant } = req.body

    const { group_conversation_id } = req.params

    const { client_user_id, client_username } = req.auth

    await new GroupChatService().removeParticipant({
      client: {
        user_id: client_user_id,
        username: client_username,
      },
      participant,
      group_conversation_id,
    })
    res.status(201)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const joinGroupController = async (req, res) => {
  try {
    const { participant } = req.body

    const { group_conversation_id } = req.params

    await new GroupChatService().joinGroup(participant, group_conversation_id)
    res.status(201)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const leaveGroupController = async (req, res) => {
  try {
    const { participant } = req.body

    const { group_conversation_id } = req.params

    await new GroupChatService().leaveGroup(participant, group_conversation_id)
    res.status(201)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const makeParticipantAdminController = async (req, res) => {
  try {
    const { participant } = req.body

    const { group_conversation_id } = req.params

    const { client_user_id, client_username } = req.auth

    await new GroupChatService().makeParticipantAdmin({
      client: {
        user_id: client_user_id,
        username: client_username,
      },
      participant,
      group_conversation_id,
    })
    res.status(201)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const removeParticipantFromAdminsController = async (req, res) => {
  try {
    const { participant } = req.body

    const { group_conversation_id } = req.params

    const { client_user_id, client_username } = req.auth

    await new GroupChatService().removeParticipantFromAdmins({
      client: {
        user_id: client_user_id,
        username: client_username,
      },
      participant,
      group_conversation_id,
    })
    res.status(201)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const changeGroupTitleController = async (req, res) => {
  try {
    const { new_group_title } = req.body

    const { group_conversation_id } = req.params

    const { client_user_id, client_username } = req.auth

    await new GroupChatService().changeGroupInfo({
      client: {
        user_id: client_user_id,
        username: client_username,
      },
      group_conversation_id,
      newInfoKVPair: { title: new_group_title },
    })
    res.status(201)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const changeGroupDescriptionController = async (req, res) => {
  try {
    const { new_group_description } = req.body

    const { group_conversation_id } = req.params

    const { client_user_id, client_username } = req.auth

    await new GroupChatService().changeGroupInfo({
      client: {
        user_id: client_user_id,
        username: client_username,
      },
      group_conversation_id,
      newInfoKVPair: { description: new_group_description },
    })
    res.status(201)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getMyConversationsController = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const myConversations = await new ChatService().getMyConversations(
      client_user_id
    )

    res.status(200).send({ myConversations })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const deleteMyConversationController = async (req, res) => {
  try {
    const { conversation_id } = req.params

    const { client_user_id } = req.auth

    await new ChatService().deleteMyConversation(
      client_user_id,
      conversation_id
    )

    res.status(200)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getConversationHistoryController = async (req, res) => {
  try {
    const { conversation_id } = req.params

    const { limit = 50, offset = 0 } = req.query

    const conversationHistory = await new ChatService().getConversationHistory({
      conversation_id,
      limit,
      offset,
    })

    res.status(200).send({ conversationHistory })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const sendMessageController = async (req, res) => {
  try {
    const { msg_content } = req.body

    const { conversation_id } = req.params

    const { client_user_id: sender_user_id } = req.auth

    await new ChatService().sendMessage({
      sender_user_id,
      conversation_id,
      msg_content,
    })

    res.status(201)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const ackMessageDeliveredController = async (req, res) => {
  try {
    const { user_id } = req.body

    const { conversation_id, message_id } = req.params

    await new ChatService().acknowledgeMessageDelivered({
      user_id,
      conversation_id,
      message_id,
    })

    res.status(201)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const ackMessageReadController = async (req, res) => {
  try {
    const { user_id } = req.body

    const { conversation_id, message_id } = req.params

    await new ChatService().acknowledgeMessageRead({
      user_id,
      conversation_id,
      message_id,
    })

    res.status(201)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const reactToMessageController = async (req, res) => {
  try {
    const { reaction_code_point } = req.body

    const { client_user_id, client_username } = req.auth

    const { conversation_id, message_id } = req.params

    await new ChatService().reactToMessage({
      reactor: {
        user_id: client_user_id,
        username: client_username,
      },
      conversation_id,
      message_id,
      reaction_code_point,
    })

    res.status(201)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const removeMyReactionToMessageController = async (req, res) => {
  try {
    const { client_user_id, client_username } = req.auth

    const { conversation_id, message_id } = req.params

    await new ChatService().removeMyReactionToMessage({
      reactor: {
        user_id: client_user_id,
        username: client_username,
      },
      conversation_id,
      message_id,
    })

    res.status(200)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const deleteMyMessageController = async (req, res) => {
  try {
    const { client_user_id, client_username } = req.auth

    const { delete_for } = req.query

    const { conversation_id, message_id } = req.params

    await new ChatService().deleteMessage({
      deleter: {
        user_id: client_user_id,
        username: client_username,
      },
      conversation_id,
      message_id,
      delete_for,
    })

    res.status(200)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
