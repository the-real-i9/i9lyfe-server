import * as appServices from "../services/app.services.js"
import * as mediaUploadService from "../services/mediaUploader.service.js"
import { Post } from "../models/post.model.js"
import { PostCommentRealtimeService } from "../services/realtime/postComment.realtime.service.js"
import { NotificationService } from "../services/realtime/notification.service.js"
import { Comment } from "../models/comment.model.js"

export const createNewPost = async (req, res) => {
  try {
    const { media_data_list, type, description } = req.body

    const { client_user_id } = req.auth

    const hashtags = appServices.extractHashtags(description)
    const mentions = appServices.extractMentions(description)

    const media_urls = await mediaUploadService.uploadPostMediaDataList(media_data_list)

    const { new_post_id, mention_notifs } = await Post.create({
      client_user_id,
      media_urls,
      type,
      description,
      mentions,
      hashtags,
    })

    const postData = await Post.find(new_post_id, client_user_id)

    // replace with message broker
    PostCommentRealtimeService.sendNewPost(client_user_id, postData)

    mention_notifs.forEach((notif) => {
      const { receiver_user_id, ...restData } = notif

      // replace with message broker
      NotificationService.sendNotification(receiver_user_id, {
        ...restData,
      })
    })

    res.status(200).send(postData)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const reactToPost = async (req, res) => {
  try {
    const { target_post_id, target_post_owner_user_id } = req.params
    const { reaction } = req.body

    const reaction_code_point = reaction.codePointAt()

    const { client_user_id } = req.auth

    const { reaction_notif, latest_reactions_count } = Post.reactTo({
      client_user_id,
      target_post_id,
      target_post_owner_user_id,
      reaction_code_point,
    })

    // notify post owner of reaction
    if (reaction_notif) {
      const { receiver_user_id, ...restData } = reaction_notif
      
      NotificationService.sendNotification(receiver_user_id, {
        ...restData,
      })
    }

    // update metrics for post for all post watchers
    PostCommentRealtimeService.sendPostMetricsUpdate(target_post_id, {
      post_id: target_post_id,
      latest_reactions_count,
    })

    // asynchronously send a reaction notification with the NotificationService via WebSockets

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const commentOnPost = async (req, res) => {
  try {
    const { target_post_id, target_post_owner_user_id } = req.params
    const { comment_text, attachment_data } = req.body

    const { client_user_id } = req.auth

    const mentions = appServices.extractMentions(comment_text)
    const hashtags = appServices.extractHashtags(comment_text)

    const attachment_url = await mediaUploadService.uploadCommentAttachmentData(attachment_data)

    const {
      new_comment_id,
      comment_notif,
      mention_notifs,
      latest_comments_count,
    } = await Post.commentOn({
      target_post_id,
      target_post_owner_user_id,
      client_user_id,
      comment_text,
      attachment_url,
      mentions,
      hashtags,
    })

    // notify mentioned users
    mention_notifs.forEach((notif) => {
      const { receiver_user_id, ...restData } = notif

      // replace with message broker
      NotificationService.sendNotification(receiver_user_id, {
        ...restData,
      })
    })

    // notify post owner of comment
    if (comment_notif) {
      const { receiver_user_id, ...restData } = comment_notif
      
      // replace with message broker
      NotificationService.sendNotification(receiver_user_id, {
        ...restData,
      })
    }
    
    // replace with message broker
    PostCommentRealtimeService.sendPostMetricsUpdate(target_post_id, {
      post_id: target_post_id,
      latest_comments_count,
    })

    // return comment data back to client
    const commentData = await Comment.find(new_comment_id, client_user_id)

    res.status(201).send(commentData)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const reactToComment = async (req, res) => {
  try {
    const { target_comment_id, target_comment_owner_user_id } = req.params
    const { reaction } = req.body

    const reaction_code_point = reaction.codePointAt()

    const { client_user_id } = req.auth

    const { reaction_notif, latest_reactions_count } =
      Comment.reactTo({
        client_user_id,
        target_comment_id,
        target_comment_owner_user_id,
        reaction_code_point,
      })

    // notify comment owner of reaction
    if (reaction_notif) {
      const { receiver_user_id, ...restData } = reaction_notif
      
      NotificationService.sendNotification(receiver_user_id, {
        ...restData,
      })
    }

    // update metrics for comment for all comment watchers
    PostCommentRealtimeService.sendCommentMetricsUpdate(target_comment_id, {
      comment_id: target_comment_id,
      latest_reactions_count,
    })

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const commentOnComment = async (req, res) => {
  try {
    const { target_comment_id, target_comment_owner_user_id } = req.params
    const { comment_text, attachment_data } = req.body

    const { client_user_id } = req.auth

    const mentions = appServices.extractMentions(comment_text)
    const hashtags = appServices.extractHashtags(comment_text)

    const attachment_url = await mediaUploadService.uploadCommentAttachmentData(attachment_data)

    const {
      new_comment_id,
      comment_notif,
      mention_notifs,
      latest_comments_count,
    } = await Comment.commentOn({
      target_comment_id,
      target_comment_owner_user_id,
      client_user_id,
      comment_text,
      attachment_url,
      mentions,
      hashtags,
    })

    // notify mentioned users
    mention_notifs.forEach((notif) => {
      const { receiver_user_id, ...restData } = notif

      NotificationService.sendNotification(receiver_user_id, {
        ...restData,
      })
    })

    // notify comment owner of comment
    if (comment_notif) {
      const { receiver_user_id, ...restData } = comment_notif

      NotificationService.sendNotification(receiver_user_id, {
        ...restData,
      })
    }

    // update metrics for comment for all comment watchers
    PostCommentRealtimeService.sendCommentMetricsUpdate(target_comment_id, {
      comment_id: target_comment_id,
      latest_comments_count,
    })

    // return comment data back to client
    const commentData = await Comment.find(new_comment_id, client_user_id)

    res.status(201).send(commentData)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const createRepost = async (req, res) => {
  try {
    const { post_id } = req.params
    const { client_user_id } = req.auth

    await Post.repost(post_id, client_user_id)

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const postSave = async (req, res) => {
  try {
    const { post_id } = req.params

    const { client_user_id } = req.auth

    const { latest_saves_count } = await Post.save(post_id, client_user_id)

    /* Realtime: currentSavesCount */
    PostCommentRealtimeService.sendPostMetricsUpdate(post_id, {
      post_id,
      latest_saves_count,
    })

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const postUnsave = async (req, res) => {
  try {
    const { post_id } = req.params
    const { client_user_id } = req.auth

    const { latest_saves_count } = await Post.unsave(post_id, client_user_id)

    /* Realtime: currentSavesCount */
    PostCommentRealtimeService.sendPostMetricsUpdate(post_id, {
      post_id,
      latest_saves_count,
    })

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/* The GETs */

export const getPost = async (req, res) => {
  try {
    const { post_id } = req.params

    const { client_user_id } = req.auth

    const post = await Post.find(post_id, client_user_id)

    res.status(200).send(post)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getCommentsOnPost = async (req, res) => {
  try {
    const { post_id } = req.params

    const { limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const commentsOnPost = await Post.getComments({
      post_id,
      client_user_id,
      limit,
      offset,
    })

    res.status(200).send(commentsOnPost)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getComment = async (req, res) => {
  try {
    const { comment_id } = req.params

    const { client_user_id } = req.auth

    const comment = await Comment.find(comment_id, client_user_id)

    res.status(200).send(comment)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getReactorsToPost = async (req, res) => {
  try {
    const { post_id } = req.params

    const { limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const postReactors = await Post.getReactors({
      post_id,
      client_user_id,
      limit,
      offset,
    })

    res.status(200).send(postReactors)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getReactorsWithReactionToPost = async (req, res) => {
  try {
    const { post_id, reaction } = req.params

    const { limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const reactorsWithReaction = await Post.getReactorsWithReaction({
      post_id,
      reaction_code_point: reaction.codePointAt(),
      client_user_id,
      limit,
      offset,
    })

    res.status(200).send(reactorsWithReaction)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getCommentsOnComment = async (req, res) => {
  try {
    const { comment_id } = req.params

    const { limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const commentsOnComment = await Comment.getComments({
      comment_id,
      client_user_id,
      limit,
      offset,
    })

    res.status(200).send(commentsOnComment)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getReactorsToComment = async (req, res) => {
  try {
    const { comment_id } = req.params

    const { limit = 20, offset = 0 } = req.query

    const { client_user_id } = req.auth

    const commentReactors = await Comment.getReactors({
      comment_id,
      client_user_id,
      limit,
      offset,
    })

    res.status(200).send(commentReactors)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const getReactorsWithReactionToComment = async (req, res) => {
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

    res.status(200).send(commentReactorsWithReaction)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

/* DELETEs */

export const deletePost = async (req, res) => {
  try {
    const { post_id } = req.params
    const { client_user_id } = req.auth

    await Post.delete(post_id, client_user_id)

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const removeReactionToPost = async (req, res) => {
  try {
    const { target_post_id } = req.params
    const { client_user_id } = req.auth

    const { latest_reactions_count } = await Post.removeReaction(
      target_post_id,
      client_user_id
    )

    PostCommentRealtimeService.sendPostMetricsUpdate(target_post_id, {
      post_id: target_post_id,
      latest_reactions_count,
    })

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const removeCommentOnPost = async (req, res) => {
  try {
    const { post_id, comment_id } = req.params
    const { client_user_id } = req.auth

    const { latest_comments_count } = await Post.removeComment({
      post_id,
      comment_id,
      client_user_id,
    })

    PostCommentRealtimeService.sendPostMetricsUpdate(post_id, {
      post_id,
      latest_comments_count,
    })

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const removeCommentOnComment = async (req, res) => {
  try {
    const { parent_comment_id, comment_id } = req.params
    const { client_user_id } = req.auth

    const { latest_comments_count } = await Comment.removeChildComment({
      parent_comment_id,
      comment_id,
      client_user_id,
    })

    PostCommentRealtimeService.sendCommentMetricsUpdate(parent_comment_id, {
      comment_id: parent_comment_id,
      latest_comments_count,
    })

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const removeReactionToComment = async (req, res) => {
  try {
    const { target_comment_id } = req.params
    const { client_user_id } = req.auth

    const { latest_reactions_count } = await Comment.removeReaction(
      target_comment_id,
      client_user_id
    )

    PostCommentRealtimeService.sendCommentMetricsUpdate(target_comment_id, {
      comment_id: target_comment_id,
      latest_reactions_count,
    })

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const deleteRepost = async (req, res) => {
  try {
    const { post_id } = req.params
    const { client_user_id } = req.auth

    await Post.unrepost(post_id, client_user_id)

    res.status(200).send({ msg: "operation successful" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
