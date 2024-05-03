import { AppService } from "../services/app.service.js"

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
    const { search = "", category = "all" } = req.query

    const result = await new AppService().searchAndFilter({
      search,
      category,
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

    const hashtagPosts = await new AppService().getHashtagPosts(
      hashtag_name,
      req.auth?.client_user_id
    )

    res.status(200).send({ hashtagPosts })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}