import {
  capitalize,
  extractHashtags,
  extractMentions,
} from "../utils/helpers.js"
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

/**
 * @param {string[]} usernames
 * @param {import("pg").PoolClient} dbClient
 * @returns {Promise<number[]>}
 */
const mapUsernamesToUserIds = async (usernames, dbClient) => {
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
 * @param {number} param0.mention_container_id
 * @param {number[]} param0.mentioned_user_ids
 * @param {"post" | "comment"} param0.mention_container
 * @param {import("pg").PoolClient} dbClient
 */
const createMentionsWithTheirNotifications = async (
  {
    mention_container,
    mention_container_id,
    mentioned_user_ids,
    post_owner_user_id,
  },
  dbClient
) => {
  /** @type {import("pg").QueryConfig} */
  const mentionQuery = {
    text: `INSERT INTO "Mention" (${mention_container}_id, user_id) 
    VALUES ${multipleRowsParameters(mentioned_user_ids.length, 2)}`,
    values: mentioned_user_ids
      .map((mention_user_id) => [mention_container_id, mention_user_id])
      .flat(),
  }

  const mentionNotificationQuery = {
    text: `INSERT INTO "PostComment_Notification" (type, sender_id, receiver_id, notification_through_${mention_container}_id) 
    VALUES ${multipleRowsParameters(mentioned_user_ids.length, 4)}`,
    values: mentioned_user_ids
      .map((mention_user_id) => [
        "mention",
        post_owner_user_id,
        mention_user_id,
        mention_container_id,
      ])
      .flat(),
  }

  await dbClient.query(mentionQuery)
  await dbClient.query(mentionNotificationQuery)
}

/**
 * @param {object} param0
 * @param {number} param0.hashtag_container_id
 * @param {string[]} param0.hashtag_names
 * @param {"post" | "comment"} param0.hashtag_container
 * @param {import("pg").PoolClient} dbClient
 */
const createHashtags = async (
  { hashtag_container, hashtag_container_id, hashtag_names },
  dbClient
) => {
  const query = {
    text: `INSERT INTO "Hashtag" (${hashtag_container}_id, hashtag_name) 
    VALUES ${multipleRowsParameters(hashtag_names.length, 2)}`,
    values: hashtag_names
      .map((hashtag_name) => [hashtag_container_id, hashtag_name])
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
 * @param {number} param0.post_owner_user_id
 * @param {"post" | "comment"} param0.reacted_to Post `id` or Comment `id`
 * @param {number} param0.reacted_to_id
 * @param {number} param0.reaction_code_point
 */
export const createReaction = async ({
  user_id,
  post_owner_user_id,
  reacted_to,
  reacted_to_id,
  reaction_code_point,
}) => {
  const dbClient = await getDBClient()
  try {
    await dbClient.query("BEGIN")
    const query = {
      text: `INSERT INTO "Reaction" (user_id, ${reacted_to}_id, reaction_code_point) 
      VALUES ($1, $2, $3) RETURNING id`,
      values: [user_id, reacted_to_id, reaction_code_point],
    }

    const { id: reaction_id } = (await dbClient.query(query)).rows[0]

    await incrementReactionsCount(
      { reacted_to, reacted_to_id },
      dbClient
    )

    await createReactionNotification(
      {
        sender_id: user_id,
        receiver_id: post_owner_user_id,
        reacted_to,
        reacted_to_id,
        reaction_id,
      },
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
 * @param {"post" | "comment"} param0.reacted_to
 * @param {number} param0.reacted_to_id Post `id` or Comment `id`
 */
const incrementReactionsCount = async (
  { reacted_to, reacted_to_id },
  dbClient
) => {
  const query = {
    text: `UPDATE "${capitalize(reacted_to)}" 
    SET reactions_count = reactions_count + 1 
    WHERE id = $1`,
    values: [reacted_to_id],
  }

  await dbClient.query(query)
}

/**
 * @param {object} param0
 * @param {number} param0.sender_id
 * @param {number} param0.receiver_id
 * @param {"post" | "comment"} param0.reacted_to
 * @param {number} param0.reacted_to_id Post `id` or Comment `id`
 * @param {number} param0.reaction_id
 */
const createReactionNotification = async (
  { sender_id, receiver_id, reacted_to, reacted_to_id, reaction_id },
  dbClient
) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `INSERT INTO "PostComment_Notification" (type, sender_id, receiver_id, notification_through_${reacted_to}_id, object_created_id),
    VALUES ($1, $2, $3, $4, $5)`,
    values: ["reaction", sender_id, receiver_id, reacted_to_id, reaction_id],
  }

  await dbClient.query(query)
}
