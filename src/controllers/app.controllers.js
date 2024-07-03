import { AppService } from "../services/app.service.js"

/**
 *
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const searchUsersToChatController = async (req, res) => {
  try {
    const { search = "", limit = 20, offset = 0 } = req.query

    const users = await AppService.searchUsersToChat({
      client_user_id: req.auth?.client_user_id,
      search,
      limit,
      offset,
    })

    res.status(200).send(users)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getExplorePostsController = async (req, res) => {
  try {
    const { limit = 20, offset = 0 } = req.query

    const explorePosts = await AppService.getExplorePosts({
      limit,
      offset,
      client_user_id: req.auth?.client_user_id,
    })

    res.status(200).send(explorePosts)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const searchAndFilterController = async (req, res) => {
  try {
    const { search = "", filter = "all", limit = 20, offset = 0 } = req.query

    const results = await AppService.searchAndFilter({
      search,
      filter,
      limit,
      offset,
      client_user_id: req.auth?.client_user_id,
    })

    res.status(200).send(results)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const getHashtagPostsController = async (req, res) => {
  try {
    const { hashtag_name } = req.params

    const { limit = 20, offset = 0 } = req.query

    const hashtagPosts = await AppService.getHashtagPosts({
      hashtag_name,
      limit,
      offset,
      client_user_id: req.auth?.client_user_id,
    })

    res.status(200).send(hashtagPosts)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
