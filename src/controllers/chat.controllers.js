/**
 * @typedef {import("express").Request} ExpressRequest
 * @typedef {import("express").Response} ExpressResponse
 */

import { Conversation, Message } from "../models/chat.model.js"
import * as mediaUploadService from "../services/mediaUploader.service.js"
import { producer } from "../services/messageBroker.service.js"

export const createConversation = async (req, res) => {
  try {
    const { partner, init_message } = req.body

    const { client_user_id, client_username } = req.auth
    const client = { user_id: client_user_id, username: client_username }

    const { media_data, ...init_msg } = init_message

    init_msg.media_url = await mediaUploadService.uploadMessageMediaData(
      media_data,
      init_msg.extension
    )

    const { client_res, partner_res } = await Conversation.create(
      client,
      partner.user_id,
      init_msg
    )

    producer.send([
      {
        topic: `user-${partner.user_id}`,
        messages: JSON.stringify({
          event: "new conversation",
          data: partner_res,
        }),
      },
      (err) => console.log(err),
    ])

    res.status(201).send(client_res)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getMyConversations = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const conversations = await Conversation.getAll(client_user_id)

    res.status(200).send(conversations)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const deleteConversation = async (req, res) => {
  try {
    const { conversation_id } = req.params

    const { client_user_id } = req.auth

    await Conversation.delete(client_user_id, conversation_id)

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getConversationHistory = async (req, res) => {
  try {
    const { conversation_id } = req.params

    const { limit = 50, offset = 0 } = req.query

    const conversationHistory = await Conversation.getHistory({
      conversation_id,
      limit,
      offset,
    })

    res.status(200).send(conversationHistory)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const sendMessage = async (req, res) => {
  try {
    const { conversation_id, partner_user_id } = req.params
    const { msg_content } = req.body

    const { client_user_id } = req.auth

    const { media_data, ...message_content } = msg_content

    message_content.media_url = await mediaUploadService.uploadMessageMediaData(
      media_data,
      message_content.extension
    )

    const { client_res, partner_res } = await Conversation.sendMessage({
      client_user_id,
      conversation_id,
      message_content,
    })

    // replace with
    producer.send([
      {
        topic: `user-${partner_user_id}`,
        messages: JSON.stringify({ event: "new message", data: partner_res }),
      },
      (err) => console.log(err),
    ])

    res.status(201).send(client_res)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const ackMessageDelivered = async (req, res) => {
  try {
    const { conversation_id, partner_user_id, message_id } = req.params

    const { delivery_time } = req.body

    const { client_user_id } = req.auth

    await Message.isDelivered({
      client_user_id,
      conversation_id,
      message_id,
      delivery_time,
    })

    producer.send([
      {
        topic: `user-${partner_user_id}`,
        messages: JSON.stringify({
          event: "message delivered",
          data: {
            conversation_id,
            message_id,
          },
        }),
      },
      (err) => console.log(err),
    ])

    res.status(200).send({ msg: "operation sucessful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const ackMessageRead = async (req, res) => {
  try {
    const { conversation_id, partner_user_id, message_id } = req.params

    const { client_user_id } = req.auth

    await Message.isRead({
      client_user_id,
      conversation_id,
      message_id,
    })

    producer.send([
      {
        topic: `user-${partner_user_id}`,
        messages: JSON.stringify({
          event: "message read",
          data: {
            conversation_id,
            message_id,
          },
        }),
      },
      (err) => console.log(err),
    ])

    res.status(200).send({ msg: "operation sucessful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const reactToMessage = async (req, res) => {
  try {
    const { conversation_id, partner_user_id, message_id } = req.params
    const { reaction } = req.body

    const { client_user_id, client_username } = req.auth

    const reactor = {
      user_id: client_user_id,
      username: client_username,
    }

    const reaction_code_point = reaction.codePointAt()

    await Message.reactTo({
      reactor_user_id: reactor.user_id,
      message_id,
      reaction_code_point,
    })

    producer.send([
      {
        topic: `user-${partner_user_id}`,
        messages: JSON.stringify({
          event: "message reaction",
          data: {
            conversation_id,
            reactor,
            message_id,
            reaction_code_point,
          },
        }),
      },
      (err) => console.log(err),
    ])

    res.status(201).send({ msg: "operation sucessful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const removeReactionToMessage = async (req, res) => {
  try {
    const { client_user_id, client_username } = req.auth

    const { conversation_id, partner_user_id, message_id } = req.params

    const reactor = {
      user_id: client_user_id,
      username: client_username,
    }

    await Message.removeReaction(message_id, reactor.user_id)

    producer.send([
      {
        topic: `user-${partner_user_id}`,
        messages: JSON.stringify({
          event: "message reaction removed",
          data: {
            reactor,
            conversation_id,
            message_id,
          },
        }),
      },
      (err) => console.log(err),
    ])

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const deleteMessage = async (req, res) => {
  try {
    const { conversation_id, partner_user_id, message_id } = req.params

    const { delete_for } = req.query

    const { client_user_id, client_username } = req.auth

    const deleter = {
      user_id: client_user_id,
      username: client_username,
    }

    await Message.delete({
      deleter_user_id: deleter.user_id,
      message_id,
      deleted_for: delete_for,
    })

    if (delete_for === "everyone") {
      producer.send([
        {
          topic: `user-${partner_user_id}`,
          messages: JSON.stringify({
            event: "message deleted",
            data: {
              conversation_id,
              deleter_username: deleter.username,
              message_id,
            },
          }),
        },
        (err) => console.log(err),
      ])
    }

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
