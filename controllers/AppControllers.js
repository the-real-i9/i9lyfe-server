/**
 * @typedef {import("express").Request} ExpressRequest
 * @typedef {import("express").Response} ExpressResponse
 */

import { AppService } from "../services/AppServices"

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const getExplorePostsController = async (req, res) => {
  try {
    const { client_user_id = null } = req.auth
    const explorePosts = await new AppService().getExplorePosts(client_user_id)

    res.status(200).send({ explorePosts })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const searchFilterController = async (req, res) => {}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const getHashtagPostsController = async (req, res) => {}
