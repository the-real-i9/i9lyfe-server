import { Post, PostCommentService } from "../services/PostCommentService.js"
import { PostService } from "../services/appServices.js"

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const postCreationController = async (req, res) => {
  try {
    const { user_id, media_urls, type, description } = req.body

    const response = await new PostService().create({
      user_id,
      media_urls,
      type,
      description,
    })

    // asychronously notify mentioned users with the notificationService (WebSockets)

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
    const { reactor_user_id, post_id, owner_user_id, reaction_code_point } = req.body

    await new PostCommentService(new Post(owner_user_id, post_id)).addReaction({
      reactor_user_id,
      reaction_code_point,
    })

    // asychronously notify post owners with the notificationService (WebSockets)

    res.sendStatus(200)
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}
