import { postCreationService } from '../services/post_managementServices.js'

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const postCreationController = async (req, res)  => {
  try {
    const { user_id, media_urls, type, description } = req.body

    const response = await postCreationService({ user_id, media_urls, type, description })
    if (!response.ok) {
      return res.status(response.err.code).send({ reason: response.err.reason })
    }

    res.status(200).send({ postData: response.data })
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}
