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

