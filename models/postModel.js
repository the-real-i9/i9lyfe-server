import { extractHashtags, extractMentions } from "../utils/helpers.js"
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
      await createHashtag(
        {
          hashtag_bearer: "post",
          hashtag_bearer_id: post_id,
          hashtag_names: hashtags,
        },
        dbClient
      )
    }

    const mentions = extractMentions(post.description)
    if (mentions) {
      const mentionIds = await mapUsernamesToUserIds(mentions, dbClient)
      await createMention(
        {
          mention_bearer: "post",
          mention_bearer_id: post_id,
          mentioned_users: mentionIds,
        },
        dbClient
      )
    }

    dbClient.query("COMMIT")
    return result
  } catch (error) {
    dbClient.query("ROLLBACK")
    console.log(error)
    throw error
  } finally {
    dbClient.release()
  }
}

/**
 * @param {string[]} usernames
 * @param {import("pg").PoolClient} dbClient
 * @returns {Promise<number[]>}
 */
export const mapUsernamesToUserIds = async (usernames, dbClient) => {
  return await Promise.all(
    usernames.map(async (username, i) => {
      const query = {
        text: 'SELECT id FROM "User" WHERE username = $1',
        values: [username],
      }
      return (await dbClient.query(query)).rows[i].id
    })
  )
}

/**
 * @param {object} param0
 * @param {number} param0.mention_bearer_id
 * @param {string[]} param0.mentioned_users
 * @param {string} param0.mention_bearer
 * @param {import("pg").PoolClient} dbClient
 */
const createMention = async (
  { mention_bearer, mention_bearer_id, mentioned_users },
  dbClient
) => {
  const query = {
    text: `INSERT INTO "Mention" (${mention_bearer}_id, user_id) 
    VALUES ${multipleInsertPlaceholders(mentioned_users)}`,
    values: multipleInsertReplacers(mention_bearer_id, mentioned_users),
  }

  await dbClient.query(query)
}

/**
 * @param {object} param0
 * @param {string} param0.hashtag_bearer_id
 * @param {string[]} param0.hashtag_names
 * @param {string} param0.hashtag_bearer
 * @param {import("pg").PoolClient} dbClient
 */
export const createHashtag = async (
  { hashtag_bearer, hashtag_bearer_id, hashtag_names },
  dbClient
) => {
  const query = {
    text: `INSERT INTO "Hashtag" (${hashtag_bearer}_id, hashtag_name) 
    VALUES ${multipleInsertPlaceholders(hashtag_names)}`,
    values: multipleInsertReplacers(hashtag_bearer_id, hashtag_names),
  }

  await dbClient.query(query)
}

/** @param {string[]} items */
export const multipleInsertPlaceholders = (items) =>
  items.map((id, i) => `($${i * 2 + 1}, $${i * 2 + 2})`).join(", ")

/**
 * Each `item` in the second replacer for each item; `rpl1` is the first.
 * @param {number} rpl1
 * @param {string[]} items
 */
export const multipleInsertReplacers = (rpl1, items) =>
  items.map((rpl2) => [rpl1, rpl2]).flat()

export const createReaction = async () => {}

export const createNewComment = async () => {}
