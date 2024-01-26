import {
  createNewPost,
  createRepost,
  deletePost,
  deleteRepost,
  getFeedPosts,
  getPost,
  savePost,
  unsavePost,
} from "../models/PostCommentModel.js"
import { getDBClient } from "../models/db.js"
import { Post, PostCommentService } from "./PostCommentService.js"
import { PostCommentRealtimeService } from "./RealtimeServices/PostCommentRealtimeService.js"

export class PostService {
  /**
   * @param {object} post
   * @param {number} post.client_user_id
   * @param {string[]} post.media_urls
   * @param {string} post.type
   * @param {string} post.description
   */
  async createPost({ client_user_id, media_urls, type, description }) {
    const dbClient = await getDBClient()
    try {
      await dbClient.query("BEGIN")

      const postData = {
        ...(await createNewPost(
          {
            client_user_id,
            media_urls,
            type,
            description,
          },
          dbClient
        )),
        reactions_count: 0,
        comments_count: 0,
        reposts_count: 0,
      }

      await new PostCommentService(
        new Post(postData.post_id, client_user_id)
      ).handleMentionsAndHashtags(
        {
          content_text: description,
          content_owner_user_id: client_user_id,
        },
        dbClient
      )

      dbClient.query("COMMIT")

      /* Realtime new post */
      new PostCommentRealtimeService().sendNewPost(client_user_id, postData)

      return postData
    } catch (error) {
      dbClient.query("ROLLBACK")
      throw error
    } finally {
      dbClient.release()
    }
  }

  /* A repost is a hasOne relationship: Repost hasOne Post */
  async repostPost(reposter_user_id, post_id) {
    await createRepost(reposter_user_id, post_id)

    /* Realtime: latestRepostsCount */
  }

  async getFeedPosts({ client_user_id, limit, offset }) {
    return await getFeedPosts({ client_user_id, limit, offset })
  }

  async getPost(post_id, client_user_id) {
    return await getPost(post_id, client_user_id)
  }

  async savePost(post_id, client_user_id) {
    const latestSavesCount = await savePost(post_id, client_user_id)

    /* Realtime: latestSavesCount */
    new PostCommentRealtimeService().sendPostCommentMetricsUpdate(post_id, {
      post_id,
      saves_count: latestSavesCount + 1,
    })
  }
  
  async unsavePost(post_id, client_user_id) {
    const latestSavesCount = await unsavePost(post_id, client_user_id)
    
    /* Realtime: latestSavesCount */
    new PostCommentRealtimeService().sendPostCommentMetricsUpdate(post_id, {
      post_id,
      saves_count: latestSavesCount - 1,
    })
  }

  async deletePost(post_id, client_user_id) {
    await deletePost(post_id, client_user_id)
  }

  async deleteRepost(reposted_post_id, client_user_id) {
    await deleteRepost(reposted_post_id, client_user_id)

    /* Realtime: latestRepostsCount */
  }
}
