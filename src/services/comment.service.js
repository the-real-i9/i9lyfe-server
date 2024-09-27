import { Comment } from "../models/comment.model.js"
import { extractHashtags, extractMentions } from "../utils/helpers.js"
import { NotificationService } from "./notification.service.js"
import { PostCommentRealtimeService } from "./realtime/postComment.realtime.service.js"

export class CommentService {
  static async commentOn({
    client_user_id,
    target_comment_id,
    target_comment_owner_user_id,
    comment_text,
    attachment_url,
  }) {
    const mentions = extractMentions(comment_text)
    const hashtags = extractHashtags(comment_text)

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

      new NotificationService(receiver_user_id).sendNotification({
        ...restData,
      })
    })

    // notify comment owner of comment
    if (comment_notif) {
      const { receiver_user_id, ...restData } = comment_notif
      new NotificationService(receiver_user_id).sendNotification({
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

    return commentData
  }

  static async getDetail(comment_id, client_user_id) {
    return await Comment.find(comment_id, client_user_id)
  }

  static async getComments({ comment_id, client_user_id, limit, offset }) {
    return await Comment.getComments({
      comment_id,
      client_user_id,
      limit,
      offset,
    })
  }

  static async removeComment({ parent_comment_id, comment_id, client_user_id }) {
    const { latest_comments_count } = await Comment.removeChildComment({
      parent_comment_id,
      comment_id,
      client_user_id,
    })

    PostCommentRealtimeService.sendCommentMetricsUpdate(parent_comment_id, {
      comment_id: parent_comment_id,
      latest_comments_count,
    })
  }

  static async reactTo({
    client_user_id,
    target_comment_id,
    target_comment_owner_user_id,
    reaction_code_point,
  }) {
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
      new NotificationService(receiver_user_id).sendNotification({
        ...restData,
      })
    }

    // update metrics for comment for all comment watchers
    PostCommentRealtimeService.sendCommentMetricsUpdate(target_comment_id, {
      comment_id: target_comment_id,
      latest_reactions_count,
    })
  }

  static async getReactors({ comment_id, client_user_id, limit, offset }) {
    return await Comment.getReactors({
      comment_id,
      client_user_id,
      limit,
      offset,
    })
  }

  static async getReactorsWithReaction({
    comment_id,
    reaction_code_point,
    client_user_id,
    limit,
    offset,
  }) {
    return await Comment.getReactorsWithReaction({
      comment_id,
      reaction_code_point,
      client_user_id,
      limit,
      offset,
    })
  }

  static async removeReaction(target_comment_id, client_user_id) {
    const { latest_reactions_count } = await Comment.removeReaction(
      target_comment_id,
      client_user_id
    )

    PostCommentRealtimeService.sendCommentMetricsUpdate(target_comment_id, {
      comment_id: target_comment_id,
      latest_reactions_count,
    })
  }
}
