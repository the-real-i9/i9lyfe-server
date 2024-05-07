import { AppService } from "../services/app.service.js"

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const getHomeFeedController = async (req, res) => {
  try {
    const { limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const homeFeedPosts = await AppService.getFeedPosts({
      client_user_id,
      limit,
      offset,
    })

    res.status(200).send({ homeFeedPosts })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const getExplorePostsController = async (req, res) => {
  try {
    const explorePosts = await AppService.getExplorePosts(
      req.auth?.client_user_id ?? null
    )

    res.status(200).send({ explorePosts })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const searchAndFilterController = async (req, res) => {
  try {
    const { search = "", filter = "all" } = req.query

    const result = await AppService.searchAndFilter({
      search,
      filter,
      client_user_id: req.auth?.client_user_id ?? null,
    })

    res.status(200).send({ result })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const getHashtagPostsController = async (req, res) => {
  try {
    const { hashtag_name } = req.params

    const hashtagPosts = await AppService.getHashtagPosts(
      hashtag_name,
      req.auth?.client_user_id
    )

    res.status(200).send({ hashtagPosts })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}
