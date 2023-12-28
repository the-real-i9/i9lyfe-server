import { capitalize } from "../utils/helpers.js"
import { getDBClient } from "./db.js"

/**
 * @param {object} post
 * @param {number} post.user_id
 * @param {string[]} post.media_urls
 * @param {string} post.type
 * @param {string} post.description
 */
export const createNewPost = async (
  { user_id, media_urls, type, description },
  dbClient
) => {
  const query = {
    text: `INSERT INTO "Post" (user_id, media_urls, type, description) 
      VALUES ($1, $2, $3, $4) 
      RETURNING id, media_urls, type, description, reactions_count, comments_count, reposts_count`,
    values: [user_id, media_urls, type, description],
  }

  const result = await dbClient.query(query)

  return result
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
 * @param {number} param0.post_or_comment_id
 * @param {number[]} param0.mentioned_user_ids
 * @param {"post" | "comment"} param0.post_or_comment
 * @param {import("pg").PoolClient} dbClient
 */
export const createMentions = async (
  { post_or_comment, post_or_comment_id, mentioned_user_ids },
  dbClient
) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `INSERT INTO "Mention" (${post_or_comment}_id, user_id) 
    VALUES ${multipleRowsParameters(mentioned_user_ids.length, 2)}`,
    values: mentioned_user_ids
      .map((mentioned_user_id) => [post_or_comment_id, mentioned_user_id])
      .flat(),
  }

  await dbClient.query(query)
}

/**
 * @param {object} param0
 * @param {number} param0.post_or_comment_id
 * @param {number[]} param0.mentioned_user_ids
 * @param {"post" | "comment"} param0.post_or_comment
 * @param {number} param0.post_or_comment_user_id
 * @param {import("pg").PoolClient} dbClient
 */
export const createMentionsNotifications = async (
  {
    post_or_comment,
    post_or_comment_id,
    mentioned_user_ids,
    post_or_comment_user_id,
  },
  dbClient
) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `INSERT INTO "PostComment_Notification" (type, sender_id, receiver_id, notification_through_${post_or_comment}_id) 
    VALUES ${multipleRowsParameters(mentioned_user_ids.length, 4)}`,
    values: mentioned_user_ids
      .map((mentioned_user_id) => [
        "mention",
        post_or_comment_user_id,
        mentioned_user_id,
        post_or_comment_id,
      ])
      .flat(),
  }

  await dbClient.query(query)
}

/**
 * @param {object} param0
 * @param {"post" | "comment"} param0.post_or_comment
 * @param {number} param0.post_or_comment_id
 * @param {string[]} param0.hashtag_names
 * @param {import("pg").PoolClient} dbClient
 */
export const createHashtags = async (
  { post_or_comment, post_or_comment_id, hashtag_names },
  dbClient
) => {
  const query = {
    text: `INSERT INTO "Hashtag" (${post_or_comment}_id, hashtag_name) 
    VALUES ${multipleRowsParameters(hashtag_names.length, 2)}`,
    values: hashtag_names
      .map((hashtag_name) => [post_or_comment_id, hashtag_name])
      .flat(),
  }

  await dbClient.query(query)
}

/**
 * @param {number} rowsCount
 * @param {number} fieldsCountPerRow
 */
const multipleRowsParameters = (rowsCount, fieldsCountPerRow) =>
  Array(rowsCount)
    .fill()
    .map(
      (r, ri) =>
        `(${Array(fieldsCountPerRow)
          .fill()
          .map((f, fi) => `$${ri * fieldsCountPerRow + (fi + 1)}`)
          .join(", ")})`
    )
    .join(", ")

/**
 * @param {object} param0
 * @param {number} param0.user_id
 * @param {number} param0.post_or_comment_user_id
 * @param {"post" | "comment"} param0.post_or_comment Post `id` or Comment `id`
 * @param {number} param0.post_or_comment_id
 * @param {number} param0.reaction_code_point
 * @param {import("pg").PoolClient} dbClient
 */
export const createReaction = async (
  { reactor_id, post_or_comment, post_or_comment_id, reaction_code_point },
  dbClient
) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `INSERT INTO "Reaction" (reactor_id, ${post_or_comment}_id, reaction_code_point) 
      VALUES ($1, $2, $3) RETURNING id`,
    values: [reactor_id, post_or_comment_id, reaction_code_point],
  }

  const result = await dbClient.query(query)

  return result
}

/**
 * @param {object} param0
 * @param {"post" | "comment"} param0.post_or_comment
 * @param {number} param0.post_or_comment_id Post `id` or Comment `id`
 * @param {import("pg").PoolClient} dbClient
 */
export const incrementReactionsCount = async (
  { post_or_comment, post_or_comment_id },
  dbClient
) => {
  const query = {
    text: `UPDATE "${capitalize(post_or_comment)}" 
    SET reactions_count = reactions_count + 1 
    WHERE id = $1`,
    values: [post_or_comment_id],
  }

  await dbClient.query(query)
}

/**
 * @param {object} param0
 * @param {number} param0.sender_id
 * @param {number} param0.receiver_id
 * @param {"post" | "comment"} param0.post_or_comment
 * @param {number} param0.post_or_comment_id Post `id` or Comment `id`
 * @param {number} param0.reaction_id
 */
export const createReactionNotification = async (
  { sender_id, receiver_id, post_or_comment, post_or_comment_id, reaction_id },
  dbClient
) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `INSERT INTO "PostComment_Notification" (type, sender_id, receiver_id, notification_through_${post_or_comment}_id, object_created_id),
    VALUES ($1, $2, $3, $4, $5)`,
    values: [
      "reaction",
      sender_id,
      receiver_id,
      post_or_comment_id,
      reaction_id,
    ],
  }

  await dbClient.query(query)
}
