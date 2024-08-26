import { AppService } from "../services/app.service.js"

export const searchUsersToChat = async (req, res) => {
  try {
    const { term = "", limit = 20, offset = 0 } = req.query

    const users = await AppService.searchUsersToChat({
      client_user_id: req.auth?.client_user_id,
      term,
      limit,
      offset,
    })

    res.status(200).send(users)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getExplorePosts = async (req, res) => {
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

export const searchAndFilter = async (req, res) => {
  try {
    const { term = "", filter = "all", limit = 20, offset = 0 } = req.query

    const results = await AppService.searchAndFilter({
      term,
      filter,
      limit,
      offset,
      client_user_id: req.auth?.client_user_id,
    })

    res.status(200).send(results)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getHashtagPosts = async (req, res) => {
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
