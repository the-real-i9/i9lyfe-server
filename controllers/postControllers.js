import {
  postCreationService,
  postReactionService,
} from "../services/postServices.js"

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const postCreationController = async (req, res) => {
  try {
    const { user_id, media_urls, type, description } = req.body

    const response = await postCreationService({
      user_id,
      media_urls,
      type,
      description,
    })
    if (!response.ok) {
      return res.status(response.err.code).send({ reason: response.err.reason })
    }

    res.status(200).send({ postData: response.data })
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const postReactionController = async (req, res) => {
  try {
    const { reaction_by, post_reacted_to, reaction_code_point } = req.body
    await postReactionService({
      user_id: reaction_by,
      post_id: post_reacted_to,
      reaction_code_point,
    })

    res.sendStatus(200)
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}
