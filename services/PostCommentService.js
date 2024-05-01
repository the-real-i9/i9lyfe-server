import { getDBClient } from "../models/db.js"
import * as PCM from "../models/PostCommentModel.js"
import { extractHashtags, extractMentions } from "../utils/helpers.js"
import { NotificationService } from "./NotificationService.js"
import { PostCommentRealtimeService } from "./RealtimeServices/PostCommentRealtimeService.js"

class Entity {
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

export class Post extends Entity {
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

export class Comment extends Entity {
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
  /** @param {Post | Comment} entity  */
  constructor(entity) {
    /** @type {Entity} */
    this.entity = entity
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
    const hashtags = extractHashtags(content_text)
    const mentions = extractMentions(content_text)
    await Promise.all([
      this.#handleHashtags(hashtags, dbClient),
      this.#handleMentions({ mentions, content_owner_user_id }, dbClient),
    ])
  }

  /**
   * @param {string[]} hashtags
   * @param {import("pg").PoolClient} dbClient
   */
  async #handleHashtags(hashtags, dbClient) {
    if (!hashtags) return
    await PCM.createHashtags(
      {
        entity: this.entity.which(),
        entity_id: this.entity.id,
        hashtag_names: hashtags,
      },
      dbClient
    )
  }

  /**
   * @param {object} param0
   * @param {string[]} param0.mentions
   * @param {number} param0.content_owner_user_id
   * @param {import("pg").PoolClient} dbClient
   */
  async #handleMentions({ mentions, content_owner_user_id }, dbClient) {
    if (!mentions) return
    const mentioned_user_ids = await PCM.mapUsernamesToUserIds(
      mentions,
      dbClient
    )
    const mentionNotifications = await PCM.createMentions(
      {
        entity: this.entity.which(),
        entity_id: this.entity.id,
        mentioned_user_ids: mentioned_user_ids.filter(
          (id) => id !== content_owner_user_id
        ),
        content_owner_user_id,
      },
      dbClient
    )

    mentionNotifications.forEach((notifData) => {
      const { receiver_user_id, ...restData } = notifData

      new NotificationService(receiver_user_id).pushNotification({
        ...restData,
        in: this.entity.which(),
      })
    })
  }

  async #createReaction(reactor_user_id, reaction_code_point) {
    const data = await PCM.createReaction({
      reactor_user_id,
      entity: this.entity.which(),
      entity_id: this.entity.id,
      content_owner_user_id: this.entity.user_id,
      reaction_code_point,
    })

    return data
  }

  #sendReactionPushNotification(notifData) {
    if (!notifData) return
    const { receiver_user_id, ...restData } = notifData

    new NotificationService(receiver_user_id).pushNotification({
      ...restData,
      to: this.entity.which(),
    })
  }

  #sendLatestEntityMetric(metricKey, metricValue) {
    new PostCommentRealtimeService().sendEntityMetricsUpdate({
      entity: this.entity.which(),
      entity_id: this.entity.id,
      data: {
        [`${this.entity.which()}_id`]: this.entity.id,
        metricKey: metricValue,
      },
    })
  }

  async addReaction(reactor_user_id, reaction_code_point) {
    const data = await this.#createReaction(
      reactor_user_id,
      reaction_code_point
    )
    this.#sendReactionPushNotification(data.notifData)
    this.#sendLatestEntityMetric(
      "reactions_count",
      data.currentReactionsCount + 1
    )
  }

  async #createComment(
    { commenter_user_id, comment_text, attachment_url },
    dbClient
  ) {
    const data = await PCM.createComment(
      {
        commenter_user_id,
        comment_text,
        attachment_url,
        entity: this.entity.which(),
        entity_id: this.entity.id,
        content_owner_user_id: this.entity.user_id,
      },
      dbClient
    )

    return data
  }

  #sendCommentPushNotification(notifData) {
    if (!notifData) return
    const { receiver_user_id, ...restData } = notifData
    new NotificationService(receiver_user_id).pushNotification({
      ...restData,
      on: this.entity.which(),
    })
  }

  async addComment({ commenter_user_id, comment_text, attachment_url }) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      const data = await this.#createComment(
        { commenter_user_id, comment_text, attachment_url },
        dbClient
      )

      await this.handleMentionsAndHashtags(
        {
          content_text: comment_text,
          content_owner_user_id: commenter_user_id,
        },
        dbClient
      )

      await dbClient.query("COMMIT")

      this.#sendCommentPushNotification(data.notifData)

      this.#sendLatestEntityMetric(
        "comments_count",
        data.currentCommentsCount + 1
      )

      const commentData = await this.getComment(
        data.comment_id,
        commenter_user_id
      )

      return commentData
    } catch (error) {
      dbClient.query("ROLLBACK")
      throw error
    } finally {
      dbClient.release()
    }
  }

  async getComments({ client_user_id, limit, offset }) {
    return await PCM.getComments({
      entity: this.entity.which(),
      entity_id: this.entity.id,
      client_user_id,
      limit,
      offset,
    })
  }

  async getComment(comment_id, client_user_id) {
    return await PCM.getComment({
      comment_id,
      client_user_id,
    })
  }

  async getReactors({ client_user_id, limit, offset }) {
    return await PCM.getReactors({
      entity: this.entity.which(),
      entity_id: this.entity.id,
      client_user_id,
      limit,
      offset,
    })
  }

  async getReactorsWithReaction({
    reaction_code_point,
    client_user_id,
    limit,
    offset,
  }) {
    return await PCM.getReactorsWithReaction({
      entity: this.entity.which(),
      entity_id: this.entity.id,
      reaction_code_point,
      client_user_id,
      limit,
      offset,
    })
  }

  async removeReaction() {
    const currentReactionsCount = await PCM.removeReaction({
      entity: this.entity.which(),
      entity_id: this.entity.id,
      reactor_user_id: this.entity.user_id,
    })

    this.#sendLatestEntityMetric("reactions_count", currentReactionsCount - 1)
  }

  async deleteComment(comment_id) {
    const currentCommentsCount = await PCM.deleteComment({
      entity: this.entity.which(),
      entity_id: this.entity.id,
      comment_id,
    })

    this.#sendLatestEntityMetric("comments_count", currentCommentsCount - 1)
  }
}
