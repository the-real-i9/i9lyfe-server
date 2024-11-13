import * as chatService from "../services/chat.service.js"

export const createConversation = async (req, res) => {
  try {
    const { partner_user_id, init_message } = req.body

    const { client_user_id } = req.auth

    const resp = chatService.createConversation({
      partner_user_id,
      client_user_id,
      init_message,
    })

    res.status(201).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getMyConversations = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const resp = await chatService.getMyConversations(client_user_id)

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const deleteConversation = async (req, res) => {
  try {
    const { conversation_id } = req.params

    const { client_user_id } = req.auth

    const resp = await chatService.deleteConversation(
      client_user_id,
      conversation_id
    )

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getConversationHistory = async (req, res) => {
  try {
    const { conversation_id } = req.params

    const { limit = 50, offset = 0 } = req.query

    const resp = await chatService.getConversationHistory({
      conversation_id,
      limit,
      offset,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const sendMessage = async (req, res) => {
  try {
    const { conversation_id, partner_user_id } = req.params
    const { msg_content } = req.body

    const { client_user_id } = req.auth

    const resp = await chatService.sendMessage({
      client_user_id,
      conversation_id,
      partner_user_id,
      msg_content,
    })

    res.status(201).send(resp.data)
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

    const resp = await chatService.ackMessageDelivered({
      client_user_id,
      partner_user_id,
      conversation_id,
      message_id,
      delivery_time,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const ackMessageRead = async (req, res) => {
  try {
    const { conversation_id, partner_user_id, message_id } = req.params

    const { client_user_id } = req.auth

    const resp = await chatService.ackMessageRead({
      client_user_id,
      partner_user_id,
      conversation_id,
      message_id,
    })

    res.status(200).send(resp.data)
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

    const resp = await chatService.reactToMessage({
      client: {
        user_id: client_user_id,
        username: client_username,
      },
      conversation_id,
      partner_user_id,
      message_id,
      reaction,
    })

    res.status(201).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const removeReactionToMessage = async (req, res) => {
  try {
    const { client_user_id, client_username } = req.auth

    const { conversation_id, partner_user_id, message_id } = req.params

    const resp = await chatService.removeReactionToMessage({
      client: {
        user_id: client_user_id,
        username: client_username,
      },
      conversation_id,
      partner_user_id,
      message_id,
    })

    res.status(200).send(resp.data)
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

    const resp = await chatService.deleteMessage({
      client: {
        user_id: client_user_id,
        username: client_username,
      },
      conversation_id,
      partner_user_id,
      message_id,
      delete_for,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
