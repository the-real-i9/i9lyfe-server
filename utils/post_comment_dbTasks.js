import { getDBClient } from "../models/db.js"
import { capitalize } from "./helpers.js"

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
 * @param {number} param0.mention_container_id
 * @param {number[]} param0.mentioned_user_ids
 * @param {"post" | "comment"} param0.mention_container
 * @param {import("pg").PoolClient} dbClient
 */
export const createMentionsWithTheirNotifications = async (
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
    text: `INSERT INTO "PostComment_Notification" (notification_type, sender_id, receiver_id, notification_through_${mention_container}_id) 
    VALUES ${multipleRowsParameters(mentioned_user_ids.length, 5)}`,
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
export const createHashtags = async (
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
export const multipleRowsParameters = (rowsCount, fieldsCountPerRow) =>
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
 *
 * @param {object} param0
 * @param {number} param0.user_id
 * @param {"post" | "comment"} param0.reaction_receiver Post `id` or Comment `id`
 * @param {number} param0.reaction_receiver_id
 * @param {number} param0.reaction_code_point
 */
export const createReaction = async ({
  user_id,
  reaction_receiver,
  reaction_receiver_id,
  reaction_code_point,
}) => {
  const dbClient = await getDBClient()
  try {
    await dbClient.query("BEGIN")
    const createNewReaction = {
      text: `INSERT INTO "Reaction" (user_id, ${reaction_receiver}_id, reaction_code_point) 
      VALUES ($1, $2, $3)`,
      values: [user_id, reaction_receiver_id, reaction_code_point],
    }

    const incrementReactionsCount = {
      text: `UPDATE "${capitalize(reaction_receiver)}" 
      SET reactions_count = reactions_count + 1 
      WHERE id = $1`,
      values: [reaction_receiver_id],
    }

    await dbClient.query(createNewReaction)
    await dbClient.query(incrementReactionsCount)

    dbClient.query("COMMIT")
  } catch (error) {
    dbClient.query("ROLLBACK")
    throw error
  } finally {
    dbClient.release()
  }
}
