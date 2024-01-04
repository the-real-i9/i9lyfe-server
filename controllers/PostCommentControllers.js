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
export const createPostController = async (req, res) => {
  // Note: You have to accept binary data(s) in the request body, upload them to a CDN, and receive their corresponding URLS in order
  try {
    const { media_urls, type, description } = req.body

    const { user_id } = req.auth

    const response = await new PostService().create({
      user_id,
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
export const reactToPostController = async (req, res) => {
  try {
    const { post_id, post_owner_user_id, reaction } = req.body
    // Should I accept the code point directly?
    const reaction_code_point = reaction.codePointAt()

    const { user_id: reactor_user_id } = req.auth

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
export const commentOnPostController = async (req, res) => {
  try {
    const {
      post_id,
      post_owner_user_id,
      comment_text,
      // attachment is a GIF, an Image, a Sticker etc. provided by frontend services via URLs
      attachment_url = null,
    } = req.body

    const { user_id: commenter_user_id } = req.auth

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
export const reactToCommentController = async (req, res) => {
  try {
    const { comment_id, comment_owner_user_id, reaction } = req.body
    // Should I accept the code point directly?
    const reaction_code_point = reaction.codePointAt()

    const { user_id: reactor_user_id } = req.auth

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
export const replyToCommentController = async (req, res) => {
  try {
    const {
      comment_id,
      comment_owner_user_id,
      reply_text,
      // attachment is a GIF, an Image, a Sticker etc. provided by frontend services via URLs
      attachment_url = null,
    } = req.body

    const { user_id: replier_user_id } = req.auth

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
export const repostPostController = async (/* req, res */) => {
  try {
  } catch (error) {}
}

/* The GETs */

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const getSinglePostController = async (req, res) => {
  try {
    const { post_id } = req.params

    const { user_id: client_user_id } = req.auth

    const post = await new PostService().get({ post_id, client_user_id })

    res.status(200).send({ post })
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const getAllPostCommentsController = async (req, res) => {
  try {
    const { post_id } = req.params

    const { user_id: client_user_id } = req.auth

    const postComments = await new PostCommentService(
      new Post(post_id)
    ).getAllCommentsORReplies(client_user_id)

    res.status(200).send({ postComments })
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const getSinglePostCommentController = async (req, res) => {
  try {
    const { comment_id } = req.params

    const { user_id: client_user_id } = req.auth

    const postComment = await new PostCommentService(
      new Post()
    ).getSingleCommentORReply({
      comment_or_reply_id: comment_id,
      client_user_id,
    })

    res.status(200).send({ postComment })
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const getAllPostReactorsController = async (req, res) => {
  try {
    const { post_id } = req.params

    const { user_id: client_user_id } = req.auth

    const postReactors = await new PostCommentService(
      new Post(post_id)
    ).getAllReactors(client_user_id)

    res.status(200).send({ postReactors })
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const getAllPostReactorsWithReactionController = async (req, res) => {
  try {
    const { post_id, reaction_code_point } = req.params

    const { user_id: client_user_id } = req.auth

    const postReactorsWithReaction = await new PostCommentService(
      new Post(post_id)
    ).getAllReactorsWithReaction(reaction_code_point, client_user_id)

    res.status(200).send({ postReactorsWithReaction })
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const getAllCommentRepliesController = async (req, res) => {
  try {
    const { comment_id } = req.params

    const { user_id: client_user_id } = req.auth

    const commentReplies = await new PostCommentService(
      new Comment(comment_id)
    ).getAllCommentsORReplies(client_user_id)

    res.status(200).send({ commentReplies })
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const getSingleCommentReplyController = async (req, res) => {
  try {
    const { reply_id } = req.params

    const { user_id: client_user_id } = req.auth

    const commentReply = await new PostCommentService(
      new Comment()
    ).getSingleCommentORReply({
      comment_or_reply_id: reply_id,
      client_user_id,
    })

    res.status(200).send({ commentReply })
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const getAllCommentReactorsController = async (req, res) => {
  try {
    const { comment_id } = req.params

    const { user_id: client_user_id } = req.auth

    const commentReactors = await new PostCommentService(
      new Comment(comment_id)
    ).getAllReactors(client_user_id)

    res.status(200).send({ commentReactors })
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const getAllCommentReactorsWithReactionController = async (req, res) => {
  try {
    const { comment_id, reaction_code_point } = req.params

    const { user_id: client_user_id } = req.auth

    const commentReactorsWithReaction = await new PostCommentService(
      new Comment(comment_id)
    ).getAllReactorsWithReaction(reaction_code_point, client_user_id)

    res.status(200).send({ commentReactorsWithReaction })
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}
