import { PostCommentRealtimeService } from "./realtime/postComment.realtime.service.js"
import { extractHashtags, extractMentions } from "../utils/helpers.js"
import { NotificationService } from "./notification.service.js"
import { Post } from "../models/post.model.js"
import { Comment } from "../models/comment.model.js"
import {
  uploadCommentAttachmentData,
  uploadPostMediaDataList,
} from "./mediaUploader.service.js"

export class PostService {
  /**
   * @param {object} post
   * @param {number} post.client_user_id
   * @param {string[]} post.media_urls
   * @param {string} post.type
   * @param {string} post.description
   */
  static async create({ client_user_id, media_data_list, type, description }) {
    const hashtags = extractHashtags(description)
    const mentions = extractMentions(description)

    const media_urls = await uploadPostMediaDataList(media_data_list)

    const { new_post_id, mention_notifs } = await Post.create({
      client_user_id,
      media_urls,
      type,
      description,
      mentions,
      hashtags,
    })

    const postData = await Post.find(new_post_id, client_user_id)

    /* Realtime new post */
    PostCommentRealtimeService.sendNewPost(client_user_id, postData)

    mention_notifs.forEach((notif) => {
      const { receiver_user_id, ...restData } = notif

      NotificationService.sendNotification(receiver_user_id, {
        ...restData,
      })
    })

    return postData
  }

  static async getDetail(post_id, client_user_id) {
    return await Post.find(post_id, client_user_id)
  }

  static async commentOn({
    client_user_id,
    target_post_id,
    target_post_owner_user_id,
    comment_text,
    attachment_data,
  }) {
    const mentions = extractMentions(comment_text)
    const hashtags = extractHashtags(comment_text)

    const attachment_url = await uploadCommentAttachmentData(attachment_data)

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

      new NotificationService(receiver_user_id).sendNotification({
        ...restData,
      })
    })

    // notify post owner of comment
    if (comment_notif) {
      const { receiver_user_id, ...restData } = comment_notif
      new NotificationService(receiver_user_id).sendNotification({
        ...restData,
      })
    }

    // update metrics for post for all post watchers
    PostCommentRealtimeService.sendPostMetricsUpdate(target_post_id, {
      post_id: target_post_id,
      latest_comments_count,
    })

    // return comment data back to client
    const commentData = await Comment.find(new_comment_id, client_user_id)

    return commentData
  }

  static async getComments({ post_id, client_user_id, limit, offset }) {
    return await Post.getComments({
      post_id,
      client_user_id,
      limit,
      offset,
    })
  }

  static async removeComment({ post_id, comment_id, client_user_id }) {
    const { latest_comments_count } = await Post.removeComment({
      post_id,
      comment_id,
      client_user_id,
    })

    PostCommentRealtimeService.sendPostMetricsUpdate(post_id, {
      post_id,
      latest_comments_count,
    })
  }

  static async reactTo({
    client_user_id,
    target_post_id,
    target_post_owner_user_id,
    reaction_code_point,
  }) {
    const { reaction_notif, latest_reactions_count } = Post.reactTo({
      client_user_id,
      target_post_id,
      target_post_owner_user_id,
      reaction_code_point,
    })

    // notify post owner of reaction
    if (reaction_notif) {
      const { receiver_user_id, ...restData } = reaction_notif
      new NotificationService(receiver_user_id).sendNotification({
        ...restData,
      })
    }

    // update metrics for post for all post watchers
    PostCommentRealtimeService.sendPostMetricsUpdate(target_post_id, {
      post_id: target_post_id,
      latest_reactions_count,
    })
  }

  static async getReactors({ post_id, client_user_id, limit, offset }) {
    return await Post.getReactors({
      post_id,
      client_user_id,
      limit,
      offset,
    })
  }

  static async getReactorsWithReaction({
    post_id,
    reaction_code_point,
    client_user_id,
    limit,
    offset,
  }) {
    return await Post.getReactorsWithReaction({
      post_id,
      reaction_code_point,
      client_user_id,
      limit,
      offset,
    })
  }

  static async removeReaction(target_post_id, client_user_id) {
    const { latest_reactions_count } = await Post.removeReaction(
      target_post_id,
      client_user_id
    )

    PostCommentRealtimeService.sendPostMetricsUpdate(target_post_id, {
      post_id: target_post_id,
      latest_reactions_count,
    })
  }

  static async save(post_id, client_user_id) {
    const { latest_saves_count } = await Post.save(post_id, client_user_id)

    /* Realtime: currentSavesCount */
    PostCommentRealtimeService.sendPostMetricsUpdate(post_id, {
      post_id,
      latest_saves_count,
    })
  }

  static async unsave(post_id, client_user_id) {
    const { latest_saves_count } = await Post.unsave(post_id, client_user_id)

    /* Realtime: currentSavesCount */
    PostCommentRealtimeService.sendPostMetricsUpdate(post_id, {
      post_id,
      latest_saves_count,
    })
  }

  static async repost(post_id, client_user_id) {
    await Post.repost(post_id, client_user_id)
  }

  static async delete(post_id, client_user_id) {
    await Post.delete(post_id, client_user_id)
  }

  static async unrepost(post_id, client_user_id) {
    await Post.unrepost(post_id, client_user_id)
  }
}
