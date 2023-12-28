import { getDBClient } from "../models/db.js"
import {
  createComment,
  createCommentNotification,
  createHashtags,
  createMentions,
  createMentionsNotifications,
  createNewPost,
  createReaction,
  createReactionNotification,
  incrementCommentsCount,
  incrementReactionsCount,
  mapUsernamesToUserIds,
} from "../models/PostCommentModel.js"
import { extractHashtags } from "../utils/helpers.js"

class PostORComment {
  constructor(user_id, id) {
    /** @type {number} */
    this.user_id = user_id
    /** @type {number} */
    this.id = id
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
  constructor(user_id, id) {
    this.user_id = user_id
    this.id = id
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
  constructor(user_id, id) {
    this.user_id = user_id
    this.id = id
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
   * @param {string} description
   * @param {import("pg").PoolClient} dbClient
   */
  async handleMentionsAndHashtags(description, dbClient) {
    await this.#handleHashtags(description, dbClient)
    await this.#handleMentions(description, dbClient)
  }

  /**
   * @param {string} description
   * @param {import("pg").PoolClient} dbClient
   */
  async #handleHashtags(description, dbClient) {
    const hashtags = extractHashtags(description)
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
   * @param {string} description
   * @param {import("pg").PoolClient} dbClient
   */
  async #handleMentions(description, dbClient) {
    const mentions = extractMentions(description)

    if (mentions) {
      await createMentions(
        {
          post_or_comment: this.postOrComment.which(),
          post_or_comment_id: this.postOrComment.id,
          mentioned_user_ids: await mapUsernamesToUserIds(mentions),
        },
        dbClient
      )

      await createMentionsNotifications(
        {
          post_or_comment: this.postOrComment.which(),
          post_or_comment_id: this.postOrComment.id,
          receiver_user_ids: await mapUsernamesToUserIds(mentions),
          sender_user_id: this.postOrComment.user_id,
        },
        dbClient
      )
    }
  }

  async addReaction({ reactor_user_id, reaction_code_point }) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      const result = await createReaction(
        {
          reactor_user_id,
          post_or_comment: this.postOrComment.which(),
          post_or_comment_id: this.postOrComment.id,
          reaction_code_point,
        },
        dbClient
      )

      const { id: reaction_id } = result.rows[0]

      await this.#incrementReactionsCount(dbClient)
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

  /** @param {import("pg").PoolClient} dbClient  */
  async #incrementReactionsCount(dbClient) {
    await incrementReactionsCount(
      {
        post_or_comment_table: this.postOrComment.getTableName(),
        post_or_comment_id: this.postOrComment.id,
      },
      dbClient
    )
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
    await createReactionNotification(
      {
        sender_user_id: reactor_user_id,
        receiver_user_id: this.postOrComment.user_id,
        post_or_comment: this.postOrComment.which(),
        post_or_comment_id: this.postOrComment.id,
        reaction_id,
      },
      dbClient
    )
  }

  async addComment({ commenter_user_id, comment_text, attachment_url }) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      const result = await createComment({
        commenter_user_id,
        comment_text,
        attachment_url,
        post_or_comment: this.postOrComment.which(),
        post_or_comment_id: this.postOrComment.id,
      }, dbClient)

      const { id: new_comment_id } = result.rows[0]

      await this.#incrementCommentsCount(dbClient)
      await this.#createCommentNotification(
        { commenter_user_id, new_comment_id },
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

  async #incrementCommentsCount(dbClient) {
    await incrementCommentsCount({
      post_or_comment_table: this.postOrComment.getTableName(),
      post_or_comment_id: this.postOrComment.id,
    }, dbClient)
  }

  async #createCommentNotification({ commenter_user_id, new_comment_id }) {
    await createCommentNotification({
      sender_user_id: commenter_user_id,
      receiver_user_id: this.postOrComment.user_id,
      post_or_comment: this.postOrComment.which(),
      post_or_comment_id: this.postOrComment.id,
      new_comment_id,
    })
  }
}
