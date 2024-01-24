/**
 * @typedef {import("express").Request} ExpressRequest
 * @typedef {import("express").Response} ExpressResponse
 */

import { ChatService } from "../services/ChatServices/Chat"

/**
 *
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const getUsersForChatController = async (req, res) => {
  try {
    const { search } = req.query

    const { client_user_id } = req.auth

    const users = new ChatService().getUsersForChat(client_user_id, search)

    res.status(200).send({ users })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
