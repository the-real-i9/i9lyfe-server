import * as chatService from "../services/chat.service.js"


export const getMyChats = async (req, res) => {
  try {
    const { client_username } = req.auth

    const resp = await chatService.getMyChats(client_username)

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const deleteChat = async (req, res) => {
  try {
    const { partner_username } = req.params

    const { client_username } = req.auth

    const resp = await chatService.deleteChat(
      client_username,
      partner_username
    )

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getChatHistory = async (req, res) => {
  try {
    const { partner_username } = req.params

    const { client_username } = req.auth

    const { limit = 50, offset = 0 } = req.query

    const resp = await chatService.getChatHistory({
      client_username,
      partner_username,
      limit,
      offset,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
