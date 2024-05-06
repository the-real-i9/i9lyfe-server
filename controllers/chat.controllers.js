/**
 * @typedef {import("express").Request} ExpressRequest
 * @typedef {import("express").Response} ExpressResponse
 */

import { ChatService } from "../services/chat/chat.service.js"

/**
 *
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const getUsersToChatController = async (req, res) => {
  try {
    const { search = "" } = req.query

    const { client_user_id } = req.auth

    const users = await ChatService.getUsersToChat(client_user_id, search)

    res.status(200).send({ users })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const createConversationController = async (req, res) => {
  try {
    const { partner, init_message } = req.body

    const { client_user_id, client_username } = req.auth

    const dm_conversation_id = await ChatService.createConversation(
      { user_id: client_user_id, username: client_username },
      partner,
      init_message
    )

    res.status(201).send({ dm_conversation_id })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const getConversationsController = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const conversations = await ChatService.getMyConversations(client_user_id)

    res.status(200).send({ conversations })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const deleteConversationController = async (req, res) => {
  try {
    const { conversation_id } = req.params

    const { client_user_id } = req.auth

    await ChatService.deleteMyConversation(client_user_id, conversation_id)

    res.sendStatus(200)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const getConversationHistoryController = async (req, res) => {
  try {
    const { conversation_id } = req.params

    const { limit = 50, offset = 0 } = req.query

    const conversationHistory = await ChatService.getConversationHistory({
      conversation_id,
      limit,
      offset,
    })

    res.status(200).send({ conversationHistory })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const sendMessageController = async (req, res) => {
  try {
    const { conversation_id, partner_user_id } = req.params
    const { msg_content } = req.body

    const { client_user_id } = req.auth

    const client_res = await ChatService.sendMessage({
      client_user_id,
      partner_user_id,
      conversation_id,
      msg_content,
    })

    res.send(201).json(client_res)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const ackMessageDeliveredController = async (req, res) => {
  try {
    const { conversation_id, partner_user_id, message_id } = req.params

    const { delivery_time } = req.body

    const { client_user_id } = req.auth

    await ChatService.acknowledgeMessageDelivered({
      client_user_id,
      partner_user_id,
      conversation_id,
      message_id,
      delivery_time
    })

    res.sendStatus(204)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const ackMessageReadController = async (req, res) => {
  try {
    const { conversation_id, partner_user_id, message_id } = req.params

    const { client_user_id } = req.auth

    await ChatService.acknowledgeMessageRead({
      client_user_id,
      partner_user_id,
      conversation_id,
      message_id,
    })

    res.sendStatus(204)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const reactToMessageController = async (req, res) => {
  try {
    const { conversation_id, partner_user_id, message_id } = req.params
    const { reaction } = req.body

    const { client_user_id, client_username } = req.auth

    await ChatService.reactToMessage({
      conversation_id,
      reactor: {
        user_id: client_user_id,
        username: client_username,
      },
      partner_user_id,
      message_id,
      reaction_code_point: reaction.codePointAt(),
    })

    res.sendStatus(201)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const removeReactionToMessageController = async (req, res) => {
  try {
    const { client_user_id, partner_user_id, client_username } = req.auth

    const { conversation_id, message_id } = req.params

    await ChatService.removeReactionToMessage({
      conversation_id,
      reactor: {
        user_id: client_user_id,
        username: client_username,
      },
      partner_user_id,
      message_id,
    })

    res.sendStatus(200)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const deleteMessageController = async (req, res) => {
  try {
    const { conversation_id, partner_user_id, message_id } = req.params
    
    const { delete_for } = req.query
    
    const { client_user_id, client_username } = req.auth
    
    await ChatService.deleteMessage({
      conversation_id,
      deleter: {
        user_id: client_user_id,
        username: client_username,
      },
      partner_user_id,
      message_id,
      delete_for,
    })

    res.sendStatus(200)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}
