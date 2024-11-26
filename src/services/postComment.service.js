import * as utilServices from "../services/utility.services.js"
import * as mediaUploadService from "../services/mediaUpload.service.js"
import { Post } from "../models/post.model.js"
import { Comment } from "../models/comment.model.js"
import * as messageBrokerService from "../services/messageBroker.service.js"
import * as realtimeService from "../services/realtime.service.js"

/**
 * @param {object} param0
 * @param {number} param0.client_user_id
 * @param {number[][]} param0.media_data_list
 * @param {"photo" | "video" | "story" | "reel"} param0.type
 * @param {string} param0.description
 */
export const createNewPost = async ({
  client_user_id,
  media_data_list,
  type,
  description,
}) => {
  const hashtags = utilServices.extractHashtags(description)
  const mentions = utilServices.extractMentions(description)

  const media_urls = media_data_list.map(async (media_data) => {
    return await mediaUploadService.upload({
      media_data,
      path_to_dest_folder: `post_medias/user-${client_user_id}`,
    })
  })

  const { new_post_data, mention_notifs } = await Post.create({
    client_user_id,
    media_urls,
    type,
    description,
    mentions,
    hashtags,
  })

  realtimeService.publishNewPost(new_post_data.post_id)

  mention_notifs.forEach((notif) => {
    const { receiver_user_id, ...restData } = notif

    // replace with message broker
    messageBrokerService.sendNewNotification(receiver_user_id, restData)
  })

  return {
    data: new_post_data,
  }
}

export const reactToPost = async ({
  client_user_id,
  post_id,
  post_owner_user_id,
  reaction,
}) => {
  const reaction_code_point = reaction.codePointAt()

  const { reaction_notif, latest_reactions_count } = Post.reactTo({
    client_user_id,
    post_id,
    post_owner_user_id,
    reaction_code_point,
  })

  // notify post owner of reaction
  if (reaction_notif) {
    const { receiver_user_id, ...restData } = reaction_notif

    messageBrokerService.sendNewNotification(receiver_user_id, restData)
  }

  // update metrics for post for all post watchers
  realtimeService.sendPostUpdate(post_id, {
    post_id,
    latest_reactions_count,
  })

  return {
    data: { msg: "operation successful" },
  }
}

export const commentOnPost = async ({
  client_user_id,
  post_id,
  post_owner_user_id,
  comment_text,
  attachment_data,
}) => {
  const mentions = utilServices.extractMentions(comment_text)
  const hashtags = utilServices.extractHashtags(comment_text)

  const attachment_url = await mediaUploadService.upload({
    media_data: attachment_data,
    path_to_dest_folder: `comment_on_post_attachments/user-${client_user_id}`,
  })

  const {
    new_comment_data,
    comment_notif,
    mention_notifs,
    latest_comments_count,
  } = await Post.commentOn({
    post_id,
    post_owner_user_id,
    client_user_id,
    comment_text,
    attachment_url,
    mentions,
    hashtags,
  })

  // notify mentioned users
  mention_notifs.forEach((notif) => {
    const { receiver_user_id, ...restData } = notif

    messageBrokerService.sendNewNotification(receiver_user_id, restData)
  })

  // notify post owner of comment
  if (comment_notif) {
    const { receiver_user_id, ...restData } = comment_notif

    messageBrokerService.sendNewNotification(receiver_user_id, restData)
  }

  realtimeService.sendPostUpdate(post_id, {
    post_id,
    latest_comments_count,
  })

  return {
    data: new_comment_data,
  }
}

export const reactToComment = async ({
  client_user_id,
  comment_id,
  comment_owner_user_id,
  reaction,
}) => {
  const reaction_code_point = reaction.codePointAt()

  const { reaction_notif, latest_reactions_count } = Comment.reactTo({
    client_user_id,
    comment_id,
    comment_owner_user_id,
    reaction_code_point,
  })

  // notify comment owner of reaction
  if (reaction_notif) {
    const { receiver_user_id, ...restData } = reaction_notif

    messageBrokerService.sendNewNotification(receiver_user_id, restData)
  }

  realtimeService.sendCommentUpdate(comment_id, {
    comment_id,
    latest_reactions_count,
  })

  return {
    data: { msg: "operation successful" },
  }
}

export const commentOnComment = async ({
  client_user_id,
  comment_id,
  comment_owner_user_id,
  comment_text,
  attachment_data,
}) => {
  const mentions = utilServices.extractMentions(comment_text)
  const hashtags = utilServices.extractHashtags(comment_text)

  const attachment_url = await mediaUploadService.upload({
    media_data: attachment_data,
    path_to_dest_folder: `comment_on_comment_attachments/user-${client_user_id}`,
  })

  const {
    new_comment_data,
    comment_notif,
    mention_notifs,
    latest_comments_count,
  } = await Comment.commentOn({
    comment_id,
    comment_owner_user_id,
    client_user_id,
    comment_text,
    attachment_url,
    mentions,
    hashtags,
  })

  // notify mentioned users
  mention_notifs.forEach((notif) => {
    const { receiver_user_id, ...restData } = notif

    messageBrokerService.sendNewNotification(receiver_user_id, restData)
  })

  // notify comment owner of comment
  if (comment_notif) {
    const { receiver_user_id, ...restData } = comment_notif

    messageBrokerService.sendNewNotification(receiver_user_id, restData)
  }

  realtimeService.sendCommentUpdate(comment_id, {
    comment_id,
    latest_comments_count,
  })

  return {
    data: new_comment_data,
  }
}

export const createRepost = async (post_id, client_user_id) => {
  await Post.repost(post_id, client_user_id)

  return {
    data: { msg: "operation successful" },
  }
}

export const savePost = async (post_id, client_user_id) => {
  const { latest_saves_count } = await Post.save(post_id, client_user_id)

  realtimeService.sendPostUpdate(post_id, {
    post_id,
    latest_saves_count,
  })

  return {
    data: { msg: "operation successful" },
  }
}

export const unsavePost = async (post_id, client_user_id) => {
  const { latest_saves_count } = await Post.unsave(post_id, client_user_id)

  realtimeService.sendPostUpdate(post_id, {
    post_id,
    latest_saves_count,
  })

  return {
    data: { msg: "operation successful" },
  }
}

/* The GETs */

export const getPost = async (post_id, client_user_id) => {
  const post = await Post.find({
    post_id,
    client_user_id,
    if_recommended: false,
  })

  return {
    data: post,
  }
}

export const getCommentsOnPost = async ({
  post_id,
  client_user_id,
  limit,
  offset,
}) => {
  const commentsOnPost = await Post.getComments({
    post_id,
    client_user_id,
    limit,
    offset,
  })

  return {
    data: commentsOnPost,
  }
}

export const getComment = async (comment_id, client_user_id) => {
  const comment = await Comment.find(comment_id, client_user_id)

  return {
    data: comment,
  }
}

export const getReactorsToPost = async ({
  post_id,
  client_user_id,
  limit,
  offset,
}) => {
  const postReactors = await Post.getReactors({
    post_id,
    client_user_id,
    limit,
    offset,
  })

  return {
    data: postReactors,
  }
}

export const getReactorsWithReactionToPost = async ({
  post_id,
  reaction,
  client_user_id,
  limit,
  offset,
}) => {
  const reactorsWithReaction = await Post.getReactorsWithReaction({
    post_id,
    reaction_code_point: reaction.codePointAt(),
    client_user_id,
    limit,
    offset,
  })

  return {
    data: reactorsWithReaction,
  }
}

export const getCommentsOnComment = async ({
  comment_id,
  client_user_id,
  limit,
  offset,
}) => {
  const commentsOnComment = await Comment.getComments({
    comment_id,
    client_user_id,
    limit,
    offset,
  })

  return {
    data: commentsOnComment,
  }
}

export const getReactorsToComment = async ({
  comment_id,
  client_user_id,
  limit,
  offset,
}) => {
  const commentReactors = await Comment.getReactors({
    comment_id,
    client_user_id,
    limit,
    offset,
  })

  return {
    data: commentReactors,
  }
}

export const getReactorsWithReactionToComment = async ({
  comment_id,
  reaction,
  client_user_id,
  limit,
  offset,
}) => {
  const commentReactorsWithReaction = await Comment.getReactorsWithReaction({
    comment_id,
    reaction_code_point: reaction.codePointAt(),
    client_user_id,
    limit,
    offset,
  })

  return {
    data: commentReactorsWithReaction,
  }
}

/* DELETEs */

export const deletePost = async (post_id, client_user_id) => {
  await Post.delete(post_id, client_user_id)

  return {
    data: { msg: "operation successful" },
  }
}

export const removeReactionToPost = async (post_id, client_user_id) => {
  const { latest_reactions_count } = await Post.removeReaction(
    post_id,
    client_user_id
  )

  realtimeService.sendPostUpdate(post_id, {
    post_id,
    latest_reactions_count,
  })

  return {
    data: { msg: "operation successful" },
  }
}

export const removeCommentOnPost = async ({
  post_id,
  comment_id,
  client_user_id,
}) => {
  const { latest_comments_count } = await Post.removeComment({
    post_id,
    comment_id,
    client_user_id,
  })

  realtimeService.sendPostUpdate(post_id, {
    post_id,
    latest_comments_count,
  })

  return {
    data: { msg: "operation successful" },
  }
}

export const removeCommentOnComment = async ({
  parent_comment_id,
  comment_id,
  client_user_id,
}) => {
  const { latest_comments_count } = await Comment.removeChildComment({
    parent_comment_id,
    comment_id,
    client_user_id,
  })

  realtimeService.sendCommentUpdate(parent_comment_id, {
    comment_id: parent_comment_id,
    latest_comments_count,
  })

  return {
    data: { msg: "operation successful" },
  }
}

export const removeReactionToComment = async (
  comment_id,
  client_user_id
) => {
  const { latest_reactions_count } = await Comment.removeReaction(
    comment_id,
    client_user_id
  )

  realtimeService.sendCommentUpdate(comment_id, {
    comment_id,
    latest_reactions_count,
  })

  return {
    data: { msg: "operation successful" },
  }
}

export const deleteRepost = async (post_id, client_user_id) => {
  await Post.unrepost(post_id, client_user_id)

  return {
    data: { msg: "operation successful" },
  }
}
