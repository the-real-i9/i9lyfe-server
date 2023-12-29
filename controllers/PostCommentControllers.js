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
export const reactToPostController = async (req, res) => {
  try {
    const {
      reactor_user_id,
      post_id,
      post_owner_user_id,
      reaction_code_point,
    } = req.body

    await new PostCommentService(
      new Post(post_owner_user_id, post_id)
    ).addReaction({
      reactor_user_id,
      reaction_code_point,
    })

    // asynchronously send a reaction notification with the NotificationService via WebSockets

    res.sendStatus(200)
  } catch (error) {
    console.log(error)
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
      commenter_user_id,
      comment_text,
      // attachment is a GIF, an Image, a Sticker etc. provided by frontend services via URLs
      attachment_url,
    } = req.body

    const response = await new PostCommentService(
      new Post(post_owner_user_id, post_id)
    ).addComment({ commenter_user_id, comment_text, attachment_url })

    // asynchronously send a comment notification with the NotificationService via WebSockets

    res.status(200).send({ commentData: response.data })
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const reactToCommentController = async (req, res) => {
  try {
    const {
      reactor_user_id,
      comment_id,
      comment_owner_user_id,
      reaction_code_point,
    } = req.body

    await new PostCommentService(
      new Comment(comment_owner_user_id, comment_id)
    ).addReaction({
      reactor_user_id,
      reaction_code_point,
    })

    // asynchronously send a reaction notification with the NotificationService via WebSockets

    res.sendStatus(200)
  } catch (error) {
    console.log(error)
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
      replier_user_id,
      reply_text,
      // attachment is a GIF, an Image, a Sticker etc. provided by frontend services via URLs
      attachment_url,
    } = req.body

    // Observe that, a reply is a comment on a comment,
    // or, technically put, Comments are nested data structures
    // All Replies are Comments and behave like Comments
    // But, not all Comments are Replies, as Comments belong to Posts and Replies do not.

    const response = await new PostCommentService(
      new Comment(comment_owner_user_id, comment_id)
    ).addComment({
      commenter_user_id: replier_user_id,
      comment_text: reply_text,
      attachment_url,
    })

    // asynchronously send a reply notification with the NotificationService via WebSockets

    res.status(200).send({ replyData: response.data })
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const repostPostController = async (req, res) => {
  try {
  } catch (error) {}
}
