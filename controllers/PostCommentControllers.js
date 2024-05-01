import {
  Comment,
  Post,
  PostCommentService,
} from "../services/PostCommentService.js"
import { PostService } from "../services/PostService.js"

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

    const postData = await new PostService().createPost({
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
    const { post_id, user_id: post_owner_user_id } = req.params
    const { reaction } = req.body
    // Should I accept the code point directly?
    const reaction_code_point = reaction.codePointAt()

    const { client_user_id: reactor_user_id } = req.auth

    await new PostCommentService(
      new Post(post_id, post_owner_user_id)
    ).addReaction(reactor_user_id, reaction_code_point)

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
    const { post_id, post_owner_user_id } = req.params
    const {
      comment_text,
      // attachment is a GIF, an Image, a Sticker etc. provided by frontend services via URLs
      attachment_url = "",
    } = req.body

    const { client_user_id: commenter_user_id } = req.auth

    const commentData = await new PostCommentService(
      new Post(post_id, post_owner_user_id)
    ).addComment({ commenter_user_id, comment_text, attachment_url })

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
    const { comment_id, comment_owner_user_id, reaction } = req.body
    // Should I accept the code point directly?
    const reaction_code_point = reaction.codePointAt()

    const { client_user_id: reactor_user_id } = req.auth

    await new PostCommentService(
      new Comment(comment_id, comment_owner_user_id)
    ).addReaction(reactor_user_id, reaction_code_point)

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
export const commentOnCommentController = async (req, res) => {
  try {
    const { comment_id, parent_comment_owner_user_id } = req.params
    const {
      comment_text,
      // attachment is a GIF, an Image, a Sticker etc. provided by frontend services via URLs
      attachment_url = null,
    } = req.body

    const { client_user_id } = req.auth

    // Observe that, a reply is a comment on a comment,
    // or, technically put, Comments are nested data structures
    // All Replies are Comments and behave like Comments
    // But, not all Comments are Replies, as Comments belong to Posts and Replies do not.

    const commentData = await new PostCommentService(
      new Comment(comment_id, parent_comment_owner_user_id)
    ).addComment({
      commenter_user_id: client_user_id,
      comment_text,
      attachment_url,
    })

    // asynchronously send a reply notification with the NotificationService via WebSockets

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
    const { post_id } = req.body
    const { client_user_id } = req.auth

    await new PostService().repostPost(client_user_id, post_id)

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
    const { post_id } = req.body

    const { client_user_id } = req.auth

    await new PostService().savePost(post_id, client_user_id)

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

    await new PostService().unsavePost(post_id, client_user_id)

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
export const getHomeFeedController = async (req, res) => {
  try {
    const { limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const homeFeedPosts = await new PostService().getFeedPosts({
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

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const getPostController = async (req, res) => {
  try {
    const { post_id } = req.params

    const { client_user_id } = req.auth

    const post = await new PostService().getPost(post_id, client_user_id)

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

    const commentsOnPost = await new PostCommentService(
      new Post(post_id)
    ).getComments({ client_user_id, limit, offset })

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

    const comment = await new PostCommentService(
      new Post()
    ).getComment(comment_id, client_user_id)

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

    const postReactors = await new PostCommentService(
      new Post(post_id)
    ).getReactors({ client_user_id, limit, offset })

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

    const reactorsWithReaction = await new PostCommentService(
      new Post(post_id)
    ).getReactorsWithReaction({
      reaction_code_point: reaction.codePointAt(),
      client_user_id,
      limit,
      offset,
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
    const { parent_comment_id } = req.params

    const { limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const commentsOnComment = await new PostCommentService(
      new Comment(parent_comment_id)
    ).getComments({ client_user_id, limit, offset })

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

    const commentReactors = await new PostCommentService(
      new Comment(comment_id)
    ).getReactors({ client_user_id, limit, offset })

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
    const { comment_id, reaction_code_point } = req.params

    const { limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const commentReactorsWithReaction = await new PostCommentService(
      new Comment(comment_id)
    ).getReactorsWithReaction({
      reaction_code_point,
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

    await new PostService().deletePost(post_id, client_user_id)

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
    const { post_id } = req.params
    const { client_user_id } = req.auth

    await new PostCommentService(
      new Post(post_id, client_user_id)
    ).removeReaction()

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
export const deleteCommentOnPostController = async (req, res) => {
  try {
    const { post_id, comment_id } = req.params
    const { client_user_id } = req.auth

    await new PostCommentService(
      new Post(post_id, client_user_id)
    ).deleteComment(comment_id)

    res.sendStatus(200)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const deleteCommentOnCommentController = async (req, res) => {
  try {
    const { parent_comment_id, comment_id } = req.params
    const { client_user_id } = req.auth

    await new PostCommentService(
      new Comment(parent_comment_id, client_user_id)
    ).deleteComment(comment_id)

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
    const { comment_id } = req.params
    const { client_user_id } = req.auth

    await new PostCommentService(
      new Comment(comment_id, client_user_id)
    ).removeReaction()

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
    const { reposted_post_id } = req.params
    const { client_user_id } = req.auth

    await new PostService().deleteRepost(reposted_post_id, client_user_id)

    res.sendStatus(200)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}
