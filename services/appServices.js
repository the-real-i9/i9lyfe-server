import { createNewPost, getPost } from "../models/PostCommentModel.js"
import { getDBClient } from "../models/db.js"
import { Post, PostCommentService } from "./PostCommentService.js"

export class PostService {
  /**
   * @param {object} post
   * @param {number} post.client_user_id
   * @param {string[]} post.media_urls
   * @param {string} post.type
   * @param {string} post.description
   */
  async create({ client_user_id, media_urls, type, description }) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      const result = await createNewPost(
        { client_user_id, media_urls, type, description },
        dbClient
      )

      const postData = {
        ...result.rows[0],
        reactions_count: 0,
        comments_count: 0,
        reposts_count: 0,
      }

      await new PostCommentService(
        new Post(postData.id, client_user_id)
      ).handleMentionsAndHashtags(description, dbClient)

      dbClient.query("COMMIT")

      return {
        ok: true,
        err: null,
        data: postData,
      }
    } catch (error) {
      dbClient.query("ROLLBACK")
      throw error
    } finally {
      dbClient.release()
    }
  }

  /* A repost is a hasOne relationship: Repost hasOne Post */
  async repost(reposter_user_id, post_id) {
    try {
    } catch (error) {}
  }

  async get(post_id, client_user_id) {
    const result = await getPost(post_id, client_user_id)

    return result
  }
}
