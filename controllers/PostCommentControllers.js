import {
  Comment,
  Post,
  PostCommentService,
} from "../services/PostCommentService.js"
import { PostService } from "../services/appServices.js"

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const createNewPostController = async (req, res) => {
  // Note: You have to accept binary data(s) in the request body, upload them to a CDN, and receive their corresponding URLS in order
  try {
    const { media_urls, type, description } = req.body

    const { client_user_id } = req.auth

    const response = await new PostService(client_user_id).create({
      media_urls,
      type,
      description,
    })

    // asychronously notify mentioned users with the notificationService (WebSockets)

    res.status(200).send({ postData: response.data })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const createPostReactionController = async (req, res) => {
  try {
    const { post_id, post_owner_user_id, reaction } = req.body
    // Should I accept the code point directly?
    const reaction_code_point = reaction.codePointAt()

    const { client_user_id: reactor_user_id } = req.auth

    await new PostCommentService(
      new Post(post_id, post_owner_user_id)
    ).addReaction({
      reactor_user_id,
      reaction_code_point,
    })

    // asynchronously send a reaction notification with the NotificationService via WebSockets

    res.sendStatus(200)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const createPostCommentController = async (req, res) => {
  try {
    const {
      post_id,
      post_owner_user_id,
      comment_text,
      // attachment is a GIF, an Image, a Sticker etc. provided by frontend services via URLs
      attachment_url = null,
    } = req.body

    const { client_user_id: commenter_user_id } = req.auth

    const response = await new PostCommentService(
      new Post(post_id, post_owner_user_id)
    ).addComment({ commenter_user_id, comment_text, attachment_url })

    // asynchronously send a comment notification with the NotificationService via WebSockets

    res.status(201).send({ commentData: response.data })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const createCommentReactionController = async (req, res) => {
  try {
    const { comment_id, comment_owner_user_id, reaction } = req.body
    // Should I accept the code point directly?
    const reaction_code_point = reaction.codePointAt()

    const { client_user_id: reactor_user_id } = req.auth

    await new PostCommentService(
      new Comment(comment_id, comment_owner_user_id)
    ).addReaction({
      reactor_user_id,
      reaction_code_point,
    })

    // asynchronously send a reaction notification with the NotificationService via WebSockets

    res.sendStatus(200)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const createCommentReplyController = async (req, res) => {
  try {
    const {
      comment_id,
      comment_owner_user_id,
      reply_text,
      // attachment is a GIF, an Image, a Sticker etc. provided by frontend services via URLs
      attachment_url = null,
    } = req.body

    const { client_user_id: replier_user_id } = req.auth

    // Observe that, a reply is a comment on a comment,
    // or, technically put, Comments are nested data structures
    // All Replies are Comments and behave like Comments
    // But, not all Comments are Replies, as Comments belong to Posts and Replies do not.

    const response = await new PostCommentService(
      new Comment(comment_id, comment_owner_user_id)
    ).addComment({
      commenter_user_id: replier_user_id,
      comment_text: reply_text,
      attachment_url,
    })

    // asynchronously send a reply notification with the NotificationService via WebSockets

    res.status(201).send({ replyData: response.data })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const createRepostController = async (/* req, res */) => {
  try {
  } catch (error) {}
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const postSaveController = async (req, res) => {
  try {
    const { post_id } = req.body

    const { client_user_id } = req.auth

    await new PostService(client_user_id, post_id).save()

    res.sendStatus(200)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/* The GETs */

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const getPostController = async (req, res) => {
  try {
    const { post_id } = req.params

    const { client_user_id } = req.auth

    const post = await new PostService(client_user_id, post_id).get()

    res.status(200).send({ post })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const getAllCommentsOnPostController = async (req, res) => {
  try {
    const { post_id } = req.params

    const { client_user_id } = req.auth

    const postComments = await new PostCommentService(
      new Post(post_id)
    ).getAllCommentsORReplies(client_user_id)

    res.status(200).send({ postComments })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const getCommentController = async (req, res) => {
  try {
    const { comment_id } = req.params

    const { client_user_id } = req.auth

    const postComment = await new PostCommentService(
      new Post()
    ).getCommentORReply(comment_id, client_user_id)

    res.status(200).send({ postComment })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const getAllReactorsToPostController = async (req, res) => {
  try {
    const { post_id } = req.params

    const { client_user_id } = req.auth

    const postReactors = await new PostCommentService(
      new Post(post_id)
    ).getAllReactors(client_user_id)

    res.status(200).send({ postReactors })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const getAllReactorsWithReactionToPostController = async (req, res) => {
  try {
    const { post_id, reaction_code_point } = req.params

    const { client_user_id } = req.auth

    const postReactorsWithReaction = await new PostCommentService(
      new Post(post_id)
    ).getAllReactorsWithReaction(reaction_code_point, client_user_id)

    res.status(200).send({ postReactorsWithReaction })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const getAllRepliesToCommentController = async (req, res) => {
  try {
    const { comment_id } = req.params

    const { client_user_id } = req.auth

    const commentReplies = await new PostCommentService(
      new Comment(comment_id)
    ).getAllCommentsORReplies(client_user_id)

    res.status(200).send({ commentReplies })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const getReplyController = async (req, res) => {
  try {
    const { reply_id } = req.params

    const { client_user_id } = req.auth

    const commentReply = await new PostCommentService(
      new Comment()
    ).getCommentORReply(reply_id, client_user_id)

    res.status(200).send({ commentReply })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const getAllReactorsToCommentController = async (req, res) => {
  try {
    const { comment_id } = req.params

    const { client_user_id } = req.auth

    const commentReactors = await new PostCommentService(
      new Comment(comment_id)
    ).getAllReactors(client_user_id)

    res.status(200).send({ commentReactors })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const getAllReactorsWithReactionToCommentController = async (
  req,
  res
) => {
  try {
    const { comment_id, reaction_code_point } = req.params

    const { client_user_id } = req.auth

    const commentReactorsWithReaction = await new PostCommentService(
      new Comment(comment_id)
    ).getAllReactorsWithReaction(reaction_code_point, client_user_id)

    res.status(200).send({ commentReactorsWithReaction })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/* DELETEs */

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const deletePostController = async (req, res) => {
  try {
    const { post_id } = req.params
    const { client_user_id } = req.auth

    await new PostService(client_user_id, post_id).delete()

    res.sendStatus(200)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const removePostReactionController = async (req, res) => {
  try {
    const { post_id } = req.params
    const { client_user_id } = req.auth

    await new PostCommentService(
      new Post(post_id, client_user_id)
    ).removeReaction()

    res.sendStatus(200)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const deletePostCommentController = async (req, res) => {
  try {
    const { comment_id } = req.params
    const { client_user_id } = req.auth

    await new PostCommentService(
      new Comment(comment_id, client_user_id)
    ).deleteCommentORReply()

    res.sendStatus(200)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const removeCommentReactionController = async (req, res) => {
  try {
    const { comment_id } = req.params
    const { client_user_id } = req.auth

    await new PostCommentService(
      new Comment(comment_id, client_user_id)
    ).removeReaction()

    res.sendStatus(200)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const deleteCommentReplyController = async (req, res) => {
  try {
    const { comment_id } = req.params
    const { client_user_id } = req.auth

    await new PostCommentService(new Comment(comment_id, client_user_id)).deleteCommentORReply()

    res.sendStatus(200)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const deleteRepostController = async (req, res) => {
  try {
    const { reposted_post_id } = req.params
    const { client_user_id } = req.auth

    await new PostService(client_user_id, reposted_post_id).deleteRepost()

    res.sendStatus(200)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const postUnsaveController = async (req, res) => {
  try {
    const { post_id } = req.params
    const { client_user_id } = req.auth

    await new PostService(client_user_id, post_id).unsave()

    res.sendStatus(200)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}