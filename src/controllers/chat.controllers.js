import * as chatService from "../services/chat.service.js"


export const getMyChats = async (req, res) => {
  try {
    const { client_user_id } = req.auth

    const resp = await chatService.getMyChats(client_user_id)

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const deleteChat = async (req, res) => {
  try {
    const { partner_user_id } = req.params

    const { client_user_id } = req.auth

    const resp = await chatService.deleteChat(
      client_user_id,
      partner_user_id
    )

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getChatHistory = async (req, res) => {
  try {
    const { chat_id } = req.params

    const { limit = 50, offset = 0 } = req.query

    const resp = await chatService.getChatHistory({
      chat_id,
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
    const { partner_user_id } = req.params
    const { msg_content, at: created_at } = req.body

    const { client_user_id } = req.auth

    const resp = await chatService.sendMessage({
      client_user_id,
      partner_user_id,
      created_at,
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
    const { partner_user_id, message_id } = req.params

    const { delivered_at } = req.body

    const { client_user_id } = req.auth

    const resp = await chatService.ackMessageDelivered({
      client_user_id,
      partner_user_id,
      message_id,
      delivered_at,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const ackMessageRead = async (req, res) => {
  try {
    const { partner_user_id, message_id } = req.params

    const { read_at } = req.body

    const { client_user_id } = req.auth

    const resp = await chatService.ackMessageRead({
      client_user_id,
      partner_user_id,
      message_id,
      read_at,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const reactToMessage = async (req, res) => {
  try {
    const { partner_user_id, message_id } = req.params
    const { reaction } = req.body

    const { client_user_id, client_username } = req.auth

    const resp = await chatService.reactToMessage({
      client: {
        user_id: client_user_id,
        username: client_username,
      },
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

    const { partner_user_id, message_id } = req.params

    const resp = await chatService.removeReactionToMessage({
      client: {
        user_id: client_user_id,
        username: client_username,
      },
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
    const { partner_user_id, message_id } = req.params

    const { delete_for } = req.query

    const { client_user_id, client_username } = req.auth

    const resp = await chatService.deleteMessage({
      client: {
        user_id: client_user_id,
        username: client_username,
      },
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
