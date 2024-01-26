/**
 * @typedef {import("express").Request} ExpressRequest
 * @typedef {import("express").Response} ExpressResponse
 */

import { AppService } from "../services/AppServices.js"

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const getExplorePostsController = async (req, res) => {
  try {
    const explorePosts = await new AppService().getExplorePosts(
      req.auth?.client_user_id ?? null
    )

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
export const searchAndFilterController = async (req, res) => {
  try {
    const { search = "", category = "all" } = req.query

    const result = await new AppService().searchAndFilter({
      search,
      category,
      client_user_id: req.auth?.client_user_id ?? null,
    })

    res.status(200).send({ result })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const getHashtagPostsController = async (req, res) => {}
