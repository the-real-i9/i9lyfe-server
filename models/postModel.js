import { extractHashtags, extractMentions } from "../utils/helpers.js"
import {
  createHashtags,
  createMentionsWithTheirNotifications,
  mapUsernamesToUserIds,
} from "../utils/post_comment_dbTasks.js"
import { getDBClient } from "./db.js"

/**
 * @param {object} post
 * @param {number} post.user_id
 * @param {string[]} post.media_urls
 * @param {string} post.type
 * @param {string} post.description
 */
export const createNewPost = async (post) => {
  const dbClient = await getDBClient()
  try {
    dbClient.query("BEGIN")
    const query = {
      text: `INSERT INTO "Post" (user_id, media_urls, type, description) 
      VALUES ($1, $2, $3, $4) 
      RETURNING id, media_urls, type, description, reactions_count, comments_count, reposts_count`,
      values: [post.user_id, post.media_urls, post.type, post.description],
    }

    const result = await dbClient.query(query)
    const { id: post_id } = result.rows[0]

    const hashtags = extractHashtags(post.description)
    if (hashtags) {
      await createHashtags(
        {
          hashtag_container: "post",
          hashtag_container_id: post_id,
          hashtag_names: hashtags,
        },
        dbClient
      )
    }

    const mentions = extractMentions(post.description)
    if (mentions) {
      await createMentionsWithTheirNotifications(
        {
          mention_container: "post",
          mention_container_id: post_id,
          mentioned_user_ids: await mapUsernamesToUserIds(mentions),
          post_owner_user_id: post.user_id,
        },
        dbClient
      )
    }

    dbClient.query("COMMIT")
    return result
  } catch (error) {
    dbClient.query("ROLLBACK")
    throw error
  } finally {
    dbClient.release()
  }
}
