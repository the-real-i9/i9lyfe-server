import { PostService as Post } from "../services/post.service.js"
import { CommentService as Comment } from "../services/comment.service.js"

/**
 * @typedef {import("express").Request} ExpressRequest
 * @typedef {import("express").Response} ExpressResponse
 */

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const createNewPostController = async (req, res) => {
  // Note: You have to accept binary data(s) in the request body, upload them to a CDN, and receive their corresponding URLS in order
  try {
    const { media_urls, type, description } = req.body

    const { client_user_id } = req.auth

    const postData = await Post.create({
      client_user_id,
      media_urls,
      type,
      description,
    })

    // asychronously notify mentioned users with the notificationService (WebSockets)

    res.status(200).send({ postData })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const reactToPostController = async (req, res) => {
  try {
    const { target_post_id, target_post_owner_user_id } = req.params
    const { reaction } = req.body

    const reaction_code_point = reaction.codePointAt()

    const { client_user_id } = req.auth

    await Post.reactTo({
      client_user_id,
      target_post_id,
      target_post_owner_user_id,
      reaction_code_point,
    })

    // asynchronously send a reaction notification with the NotificationService via WebSockets

    res.sendStatus(200)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const commentOnPostController = async (req, res) => {
  try {
    const { target_post_id, target_post_owner_user_id } = req.params
    const {
      comment_text,
      // attachment is a GIF, an Image, a Sticker etc. provided by frontend services via URLs
      attachment_url = "",
    } = req.body

    const { client_user_id } = req.auth

    const commentData = await Post.commentOn({
      client_user_id,
      target_post_id,
      target_post_owner_user_id,
      comment_text,
      attachment_url,
    })

    // asynchronously send a comment notification with the NotificationService via WebSockets

    res.status(201).send({ commentData })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const reactToCommentController = async (req, res) => {
  try {
    const { target_comment_id, target_comment_owner_user_id } = req.params
    const { reaction } = req.body

    const reaction_code_point = reaction.codePointAt()

    const { client_user_id } = req.auth

    await Comment.reactTo({
      client_user_id,
      target_comment_id,
      target_comment_owner_user_id,
      reaction_code_point,
    })

    res.sendStatus(200)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const commentOnCommentController = async (req, res) => {
  try {
    const { target_comment_id, target_comment_owner_user_id } = req.params
    const {
      comment_text,
      // attachment is a GIF, an Image, a Sticker etc. provided by frontend services via URLs
      attachment_url = "",
    } = req.body

    const { client_user_id } = req.auth

    const commentData = await Comment.commentOn({
      client_user_id,
      target_comment_id,
      target_comment_owner_user_id,
      comment_text,
      attachment_url
    })

    res.status(201).send({ commentData })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const createRepostController = async (req, res) => {
  try {
    const { post_id } = req.params
    const { client_user_id } = req.auth

    await Post.repost(post_id, client_user_id)

    res.sendStatus(200)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const postSaveController = async (req, res) => {
  try {
    const { post_id } = req.params

    const { client_user_id } = req.auth

    await Post.save(post_id, client_user_id)

    res.sendStatus(200)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const postUnsaveController = async (req, res) => {
  try {
    const { post_id } = req.params
    const { client_user_id } = req.auth

    await Post.unsave(post_id, client_user_id)

    res.sendStatus(200)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/* The GETs */

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const getPostController = async (req, res) => {
  try {
    const { post_id } = req.params

    const { client_user_id } = req.auth

    const post = await Post.getDetail(post_id, client_user_id)

    res.status(200).send({ post })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const getCommentsOnPostController = async (req, res) => {
  try {
    const { post_id } = req.params

    const { limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const commentsOnPost = await Post.getComments({
      post_id,
      client_user_id,
      limit,
      offset
    })

    res.status(200).send({ commentsOnPost })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const getCommentController = async (req, res) => {
  try {
    const { comment_id } = req.params

    const { client_user_id } = req.auth

    const comment = await Comment.getDetail(comment_id, client_user_id)

    res.status(200).send({ comment })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const getReactorsToPostController = async (req, res) => {
  try {
    const { post_id } = req.params

    const { limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const postReactors = await Post.getReactors({
      post_id,
      client_user_id,
      limit,
      offset
    })

    res.status(200).send({ postReactors })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const getReactorsWithReactionToPostController = async (req, res) => {
  try {
    const { post_id, reaction } = req.params

    const { limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const reactorsWithReaction = await Post.getReactorsWithReaction({
      post_id,
      reaction_code_point: reaction.codePointAt(),
      client_user_id,
      limit,
      offset
    })

    res.status(200).send({ reactorsWithReaction })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const getCommentsOnCommentController = async (req, res) => {
  try {
    const { comment_id } = req.params

    const { limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const commentsOnComment = await Comment.getComments({
      comment_id,
      client_user_id,
      limit,
      offset
    })

    res.status(200).send({ commentsOnComment })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const getReactorsToCommentController = async (req, res) => {
  try {
    const { comment_id } = req.params

    const { limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const commentReactors = await Comment.getReactors({
      comment_id,
      client_user_id,
      limit,
      offset
    })

    res.status(200).send({ commentReactors })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const getReactorsWithReactionToCommentController = async (req, res) => {
  try {
    const { comment_id, reaction } = req.params

    const { limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const commentReactorsWithReaction = await Comment.getReactorsWithReaction({
      comment_id,
      reaction_code_point: reaction.codePointAt(),
      client_user_id,
      limit,
      offset,
    })

    res.status(200).send({ commentReactorsWithReaction })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/* DELETEs */

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const deletePostController = async (req, res) => {
  try {
    const { post_id } = req.params
    const { client_user_id } = req.auth

    await Post.delete(post_id, client_user_id)

    res.sendStatus(200)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const removeReactionToPostController = async (req, res) => {
  try {
    const { target_post_id } = req.params
    const { client_user_id } = req.auth

    await Post.removeReaction(target_post_id, client_user_id)

    res.sendStatus(200)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const removeCommentOnPostController = async (req, res) => {
  try {
    const { post_id, comment_id } = req.params
    const { client_user_id } = req.auth

    await Post.removeComment({
      post_id,
      comment_id,
      client_user_id,
    })

    res.sendStatus(200)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const removeCommentOnCommentController = async (req, res) => {
  try {
    const { parent_comment_id, comment_id } = req.params
    const { client_user_id } = req.auth

    await Comment.removeComment({
      parent_comment_id,
      comment_id,
      client_user_id,
    })

    res.sendStatus(200)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const removeReactionToCommentController = async (req, res) => {
  try {
    const { target_comment_id } = req.params
    const { client_user_id } = req.auth

    await Comment.removeReaction(target_comment_id, client_user_id)

    res.sendStatus(200)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const deleteRepostController = async (req, res) => {
  try {
    const { post_id } = req.params
    const { client_user_id } = req.auth

    await Post.unrepost(post_id, client_user_id)

    res.sendStatus(200)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}
