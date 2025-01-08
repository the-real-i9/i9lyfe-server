import * as appService from "../services/app.service.js"

export const getExploreFeed = async (req, res) => {
  try {
    const { limit = 20, offset = 0 } = req.query

    const resp = await appService.getExploreFeed({
      limit,
      offset,
      client_user_id: req.auth?.client_user_id,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getExploreReels = async (req, res) => {
  try {
    const { limit = 20, offset = 0 } = req.query

    const resp = await appService.getExploreReels({
      limit,
      offset,
      client_user_id: req.auth?.client_user_id,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const searchAndFilter = async (req, res) => {
  try {
    const { term, filter = "all", limit = 20, offset = 0 } = req.query

    const resp = await appService.searchAndFilter({
      client_user_id: req.auth?.client_user_id,
      term,
      filter,
      limit,
      offset,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getHashtagPosts = async (req, res) => {
  try {
    const { hashtag_name } = req.params

    const { filter = "all", limit = 20, offset = 0 } = req.query

    const resp = await appService.getHashtagPosts({
      hashtag_name,
      filter,
      limit,
      offset,
      client_user_id: req.auth?.client_user_id,
    })

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
