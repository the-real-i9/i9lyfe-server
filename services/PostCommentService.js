import { getDBClient } from "../models/db.js"
import {
  createComment,
  createCommentNotification,
  createHashtags,
  createMentions,
  createMentionsNotifications,
  createReaction,
  createReactionNotification,
  deleteCommentOnPost_OR_ReplyToComment,
  getAllCommentsOnPost_OR_RepliesToComment,
  getAllReactorsToPost_OR_Comment,
  getAllReactorsWithReactionToPost_OR_Comment,
  getCommentOnPost_OR_ReplyToComment,
  mapUsernamesToUserIds,
  removeReactionToPost_OR_Comment,
} from "../models/PostCommentModel.js"
import { extractHashtags, extractMentions } from "../utils/helpers.js"
import { NotificationService } from "./NotificationService.js"

class PostORComment {
  constructor(id, user_id) {
    /** @type {number} */
    this.id = id
    /** @type {number} */
    this.user_id = user_id
  }
  /** @returns {"post" | "comment"} */
  which() {
    throw new Error("which method must be implemented")
  }

  /** @returns {"Post" | "Comment"} */
  getTableName() {
    throw new Error("getTableName method must be implemented")
  }
}

export class Post extends PostORComment {
  /**
   * @param {number} user_id
   * @param {number} id
   */
  constructor(id, user_id) {
    super()
    this.id = id
    this.user_id = user_id
  }

  which() {
    return "post"
  }

  getTableName() {
    return "Post"
  }
}

export class Comment extends PostORComment {
  /**
   * @param {number} user_id
   * @param {number} id
   */
  constructor(id, user_id) {
    super()
    this.id = id
    this.user_id = user_id
  }

  which() {
    return "comment"
  }

  getTableName() {
    return "Comment"
  }
}

export class PostCommentService {
  /** @param {Post | Comment} postOrComment  */
  constructor(postOrComment) {
    /** @type {PostORComment} */
    this.postOrComment = postOrComment
  }
  /**
   * @param {object} param0
   * @param {string} param0.content_text Post description or Comment text
   * @param {number} param0.content_owner_user_id
   * @param {import("pg").PoolClient} dbClient
   */
  async handleMentionsAndHashtags(
    { content_text, content_owner_user_id },
    dbClient
  ) {
    await Promise.all([
      this.#handleHashtags(content_text, dbClient),
      this.#handleMentions({ content_text, content_owner_user_id }, dbClient),
    ])
  }

  /**
   * @param {string} content_text
   * @param {import("pg").PoolClient} dbClient
   */
  async #handleHashtags(content_text, dbClient) {
    const hashtags = extractHashtags(content_text)
    if (hashtags) {
      await createHashtags(
        {
          post_or_comment: this.postOrComment.which(),
          post_or_comment_id: this.postOrComment.id,
          hashtag_names: hashtags,
        },
        dbClient
      )
    }
  }

  /**
   * @param {object} param0
   * @param {string} param0.content_text
   * @param {number} param0.content_owner_user_id
   * @param {import("pg").PoolClient} dbClient
   */
  async #handleMentions({ content_text, content_owner_user_id }, dbClient) {
    const mentions = extractMentions(content_text)

    if (mentions) {
      const mentioned_user_ids = await mapUsernamesToUserIds(mentions, dbClient)
      await createMentions(
        {
          post_or_comment: this.postOrComment.which(),
          post_or_comment_id: this.postOrComment.id,
          mentioned_user_ids,
        },
        dbClient
      )

      const mentionNotifications = await createMentionsNotifications(
        {
          post_or_comment: this.postOrComment.which(),
          post_or_comment_id: this.postOrComment.id,
          receiver_user_ids: mentioned_user_ids.filter(
            (id) => id !== content_owner_user_id
          ),
          sender_user_id: content_owner_user_id,
        },
        dbClient
      )

      mentionNotifications.forEach((notifData) => {
        const { receiver_user_id, ...restData } = notifData

        new NotificationService(receiver_user_id).pushNotification({
          ...restData,
          in: this.postOrComment.which(),
        })
      })
    }
  }

  async addReaction(reactor_user_id, reaction_code_point) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      const reaction_id = await createReaction(
        {
          reactor_user_id,
          post_or_comment: this.postOrComment.which(),
          post_or_comment_id: this.postOrComment.id,
          reaction_code_point,
        },
        dbClient
      )

      await this.#createReactionNotification(
        { reactor_user_id, reaction_id },
        dbClient
      )

      dbClient.query("COMMIT")
    } catch (error) {
      dbClient.query("ROLLBACK")
      throw error
    } finally {
      dbClient.release()
    }
  }

  /**
   * @param {object} param0
   * @param {number} param0.reactor_user_id
   * @param {number} param0.reaction_id
   * @param {import("pg").PoolClient} dbClient
   */
  async #createReactionNotification(
    { reactor_user_id, reaction_id },
    dbClient
  ) {
    if (reactor_user_id === this.postOrComment.user_id) return
    const notifData = await createReactionNotification(
      {
        sender_user_id: reactor_user_id,
        receiver_user_id: this.postOrComment.user_id,
        post_or_comment: this.postOrComment.which(),
        post_or_comment_id: this.postOrComment.id,
        reaction_id,
      },
      dbClient
    )

    const { receiver_user_id, ...restData } = notifData

    new NotificationService(receiver_user_id).pushNotification({
      ...restData,
      to: this.postOrComment.which(),
    })
  }

  async addComment({ commenter_user_id, comment_text, attachment_url }) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      const commentData = {
        ...(await createComment(
          {
            commenter_user_id,
            comment_text,
            attachment_url,
            post_or_comment: this.postOrComment.which(),
            post_or_comment_id: this.postOrComment.id,
          },
          dbClient
        )),
        reactions_count: 0,
        replies_count: 0,
      }

      const { id: new_comment_id } = commentData

      await Promise.all([
        this.handleMentionsAndHashtags(
          {
            content_text: comment_text,
            content_owner_user_id: commenter_user_id,
          },
          dbClient
        ),
        this.#createCommentNotification(
          { commenter_user_id, new_comment_id },
          dbClient
        ),
      ])

      dbClient.query("COMMIT")

      return commentData
    } catch (error) {
      dbClient.query("ROLLBACK")
      throw error
    } finally {
      dbClient.release()
    }
  }

  async #createCommentNotification(
    { commenter_user_id, new_comment_id },
    dbClient
  ) {
    if (commenter_user_id === this.postOrComment.user_id) return
    const notifData = await createCommentNotification(
      {
        sender_user_id: commenter_user_id,
        receiver_user_id: this.postOrComment.user_id,
        post_or_comment: this.postOrComment.which(),
        post_or_comment_id: this.postOrComment.id,
        new_comment_id,
      },
      dbClient
    )

    const { receiver_user_id, ...restData } = notifData
    new NotificationService(receiver_user_id).pushNotification({
      ...restData,
      on: this.postOrComment.which(),
    })
  }

  async addReply({ replier_user_id, reply_text, attachment_url }) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      // Note: A Reply is also a form of Coment. It's just a  comment that belongs to a Comment
      const replyData = {
        ...(await createComment(
          {
            commenter_user_id: replier_user_id,
            comment_text: reply_text,
            attachment_url,
            post_or_comment: this.postOrComment.which(),
            post_or_comment_id: this.postOrComment.id,
          },
          dbClient
        )),
        reactions_count: 0,
        replies_count: 0,
      }

      const { id: new_reply_id } = replyData

      await Promise.all([
        this.handleMentionsAndHashtags(
          {
            content_text: reply_text,
            content_owner_user_id: replier_user_id,
          },
          dbClient
        ),
        this.#createReplyNotification(
          { replier_user_id, new_reply_id },
          dbClient
        ),
      ])

      dbClient.query("COMMIT")

      return replyData
    } catch (error) {
      dbClient.query("ROLLBACK")
      throw error
    } finally {
      dbClient.release()
    }
  }

  async #createReplyNotification({ replier_user_id, new_reply_id }, dbClient) {
    if (replier_user_id === this.postOrComment.user_id) return
    const notifData = await createCommentNotification(
      {
        sender_user_id: replier_user_id,
        receiver_user_id: this.postOrComment.user_id,
        post_or_comment: this.postOrComment.which(),
        post_or_comment_id: this.postOrComment.id,
        new_comment_id: new_reply_id,
      },
      dbClient
    )

    const { receiver_user_id, ...restData } = notifData
    new NotificationService(receiver_user_id).pushNotification({
      ...restData,
      type: "reply",
      to: this.postOrComment.which(),
    })
  }

  async getAllCommentsORReplies(client_user_id) {
    const result = await getAllCommentsOnPost_OR_RepliesToComment({
      post_or_comment: this.postOrComment.which(),
      post_or_comment_id: this.postOrComment.id,
      client_user_id,
    })

    return result
  }

  async getCommentORReply(comment_or_reply_id, client_user_id) {
    const result = await getCommentOnPost_OR_ReplyToComment({
      post_or_comment: this.postOrComment.which(),
      comment_or_reply_id,
      client_user_id,
    })

    return result
  }

  async getAllReactors(client_user_id) {
    const result = await getAllReactorsToPost_OR_Comment({
      post_or_comment: this.postOrComment.which(),
      post_or_comment_id: this.postOrComment.id,
      client_user_id,
    })

    return result
  }

  async getAllReactorsWithReaction(reaction_code_point, client_user_id) {
    const result = await getAllReactorsWithReactionToPost_OR_Comment({
      post_or_comment: this.postOrComment.which(),
      post_or_comment_id: this.postOrComment.id,
      reaction_code_point,
      client_user_id,
    })

    return result
  }

  async removeReaction() {
    await removeReactionToPost_OR_Comment({
      post_or_comment: this.postOrComment.which(),
      post_or_comment_id: this.postOrComment.id,
      reactor_user_id: this.postOrComment.user_id,
    })
  }

  async deleteCommentORReply() {
    await deleteCommentOnPost_OR_ReplyToComment({
      comment_or_reply_id: this.postOrComment.id,
      commenter_or_replier_user_id: this.postOrComment.user_id,
    })
  }
}
