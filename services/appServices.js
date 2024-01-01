import { createNewPost } from "../models/PostCommentModel.js"
import { getDBClient } from "../models/db.js"
import { Post, PostCommentService } from "./PostCommentService.js"

export class PostService {
  /**
   * @param {object} post
   * @param {number} post.user_id
   * @param {string[]} post.media_urls
   * @param {string} post.type
   * @param {string} post.description
   */
  async create({ user_id, media_urls, type, description }) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      const result = await createNewPost(
        { user_id, media_urls, type, description },
        dbClient
      )

      const postData = result.rows[0]

      await new PostCommentService(
        new Post(user_id, postData.id)
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
  async repost({ reposter_user_id, post_id }) {
    try {
      
    } catch (error) {
      
    }
  }
}