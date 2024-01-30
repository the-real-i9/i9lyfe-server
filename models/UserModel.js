import {
  commaSeparateString,
  generateMultiColumnUpdateSetParameters,
  stripNulls,
} from "../utils/helpers.js"
import { dbQuery } from "./db.js"

/**
 * @param {Object} fields
 * @param {string} fields.email
 * @param {string} fields.username
 * @param {string} fields.password
 * @param {string} fields.name
 * @param {Date} fields.birthday
 * @param {string} fields.bio
 */
export const createNewUser = async (fields) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `INSERT INTO "User"(email, username, password, name, birthday, bio) 
      VALUES($1, $2, $3, $4, $5, $6) 
      RETURNING id, email, username, name, profile_pic_url`,
    values: [
      fields.email,
      fields.username,
      fields.password,
      fields.name,
      fields.birthday,
      fields.bio,
    ],
  }

  const result = await dbQuery(query)

  return result
}

/**
 * @param {string} email
 * @param {string} selectFields
 */
export const getUserByEmail = async (email, selectFields) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `SELECT ${commaSeparateString(
      selectFields
    )} FROM "User" WHERE email = $1`,
    values: [email],
  }

  const result = await dbQuery(query)

  return result
}

/** @param {string} email */
export const userExists = async (email) => {
  const result = await getUserByEmail(email, "1")
  return result.rowCount > 0 ? true : false
}

export const changeUserPassword = async (email, newPassword) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: 'UPDATE "User" SET password = $2 WEHRE email = $1',
    values: [email, newPassword],
  }

  await dbQuery(query)
}

/**
 * @param {object} param0
 * @param {number} param0.client_user_id
 * @param {number} param0.to_follow_user_id
 */
export const followUser = async ({ client_user_id, to_follow_user_id }) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `
    WITH new_follow_cte AS (
      INSERT INTO "Follow" (follower_user_id, followee_user_id) VALUES ($1, $2) 
    ), follow_notif_cte AS (
      INSERT INTO "Notification" (type, sender_user_id, receiver_user_id) 
      VALUES ($3, $1, $2) 
      RETURNING type, sender_user_id, receiver_user_id
    )
    SELECT notification.type,
      notification.receiver_user_id,
      json_build_object(
        'user_id', sender.id,
        'username', sender.username,
        'profile_pic_url', sender.profile_pic_url
      ) AS sender
    FROM follow_notif_cte notification
    INNER JOIN "User" sender ON sender.id = notification.sender_user_id`,
    values: [client_user_id, to_follow_user_id, "follow"],
  }

  return (await dbQuery(query)).rows[0]
}

/**
 * @param {number} client_user_id
 * @param {number} followee_user_id
 */
export const unfollowUser = async (client_user_id, followee_user_id) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `DELETE FROM "Follow" WHERE follower_user_id = $1 AND followee_user_id = $2`,
    values: [client_user_id, followee_user_id],
  }

  await dbQuery(query)
}

/**
 * @param {number} client_user_id
 * @param {Map<string, any>} updateKVPairs
 */
export const updateUserProfile = async (client_user_id, updateKVPairs) => {
  const [updateSetCols, updateSetValues] = [
    [...updateKVPairs.keys()],
    [...updateKVPairs.values()],
  ]

  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `UPDATE "User" SET ${generateMultiColumnUpdateSetParameters(
      updateSetCols,
      2
    )} 
    WHERE id = $1 
    RETURNING id, email, username, name, profile_pic_url, bio, birthday`,
    values: [client_user_id, ...updateSetValues],
  }

  return (await dbQuery(query)).rows[0]
}

export const uploadProfilePicture = async (client_user_id, profile_pic_url) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `UPDATE "User" SET profile_pic_url = $1 WHERE id = $2`,
    values: [profile_pic_url, client_user_id],
  }

  await dbQuery(query)
}

/* ************* */

// GET user profile data
/** @param {string} username */
export const getUserProfile = async (username, client_user_id) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `
    SELECT "user".id AS user_id, 
      name, 
      username, 
      bio, 
      profile_pic_url, 
      COUNT("followee".id) AS followers_count, 
      COUNT("follower".id) AS following_count,
      CASE
        WHEN "client_follows".id IS NULL THEN false
        ELSE true
      END client_follows
    FROM "User" "user"
    LEFT JOIN "Follow" "followee" ON "followee".followee_user_id = "user".id
    LEFT JOIN "Follow" "follower" ON "follower".follower_user_id = "user".id
    LEFT JOIN "Follow" "client_follows" 
      ON "client_follows".followee_user_id = "user".id AND "client_follows".follower_user_id = $2
    WHERE "user".username = $1
    GROUP BY "user".id,
      name,
      username,
      bio,
      profile_pic_url,
      "client_follows".id`,
    values: [username, client_user_id],
  }

  return (await dbQuery(query)).rows[0]
}

// GET user followers
export const getUserFollowers = async (username, client_user_id) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `
    SELECT "follower_user".id AS user_id, 
      "follower_user".username, 
      "follower_user".bio, 
      "follower_user".profile_pic_url,
      CASE
        WHEN "client_follows".id IS NULL THEN false
        ELSE true
      END client_follows
    FROM "Follow" "follow"
    LEFT JOIN "User" "follower_user" ON "follower_user".id = "follow".follower_user_id
    LEFT JOIN "User" "followee_user" ON "followee_user".id = "follow".followee_user_id
    LEFT JOIN "Follow" "client_follows" 
      ON "client_follows".followee_user_id = "followee_user".id AND "client_follows".follower_user_id = $2
    WHERE "followee_user".username = $1`,
    values: [username, client_user_id],
  }

  return (await dbQuery(query)).rows
}

// GET user followings
export const getUserFollowing = async (username, client_user_id) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `
    SELECT "followee_user".id AS user_id, 
      "followee_user".username, 
      "followee_user".bio, 
      "followee_user".profile_pic_url,
      CASE
        WHEN "client_follows".id IS NULL THEN false
        ELSE true
      END client_follows
    FROM "Follow" "follow"
    LEFT JOIN "User" "follower_user" ON "follower_user".id = "follow".follower_user_id
    LEFT JOIN "User" "followee_user" ON "followee_user".id = "follow".followee_user_id
    LEFT JOIN "Follow" "client_follows" 
      ON "client_follows".followee_user_id = "followee_user".id AND "client_follows".follower_user_id = $2
    WHERE "follower_user".username = $1`,
    values: [username, client_user_id],
  }

  return (await dbQuery(query)).rows
}

// GET user posts
/** @param {string} username */
export const getUserPosts = async (username, client_user_id) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `
    SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      post_id,
      type,
      media_urls,
      description,
      reactions_count,
      comments_count,
      reposts_count,
      saves_count,
      CASE 
        WHEN reactor_user_id = $2 THEN reaction_code_point
        ELSE NULL
      END AS client_reaction,
      CASE 
        WHEN reposter_user_id = $2 THEN true
        ELSE false
      END AS client_reposted,
      CASE 
        WHEN saver_user_id = $2 THEN true
        ELSE false
      END AS client_saved
    FROM "AllPostsView"
    WHERE owner_username = $1`,
    values: [username, client_user_id],
  }

  return (await dbQuery(query)).rows
}

// GET posts user has been mentioned in
export const getMentionedPosts = async (client_user_id) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `
    SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      apv.post_id,
      type,
      media_urls,
      description,
      reactions_count,
      comments_count,
      reposts_count,
      saves_count,
      CASE 
        WHEN reactor_user_id = $1 THEN reaction_code_point
        ELSE NULL
      END AS client_reaction,
      CASE 
        WHEN reposter_user_id = $1 THEN true
        ELSE false
      END AS client_reposted,
      CASE 
        WHEN saver_user_id = $1 THEN true
        ELSE false
      END AS client_saved
    FROM "AllPostsView" apv
    INNER JOIN "PostCommentMention" mention ON mention.post_id = apv.post_id AND mention.user_id = $1`,
    values: [client_user_id],
  }

  return (await dbQuery(query)).rows
}

// GET posts reacted by user
export const getReactedPosts = async (client_user_id) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `
    SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      post_id,
      type,
      media_urls,
      description,
      reactions_count,
      comments_count,
      reposts_count,
      saves_count,
      CASE 
        WHEN reactor_user_id = $1 THEN reaction_code_point
        ELSE NULL
      END AS client_reaction,
      CASE 
        WHEN reposter_user_id = $1 THEN true
        ELSE false
      END AS client_reposted,
      CASE 
        WHEN saver_user_id = $1 THEN true
        ELSE false
      END AS client_saved
    FROM "AllPostsView"
    WHERE reactor_user_id = $1`,
    values: [client_user_id],
  }

  return (await dbQuery(query)).rows
}

// GET posts saved by this user
export const getSavedPosts = async (client_user_id) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `
    SELECT json_build_object(
        'user_id', owner_user_id,
        'username', owner_username,
        'profile_pic_url', owner_profile_pic_url
      ) AS owner_user,
      post_id,
      type,
      media_urls,
      description,
      reactions_count,
      comments_count,
      reposts_count,
      saves_count,
      CASE 
        WHEN reactor_user_id = $1 THEN reaction_code_point
        ELSE NULL
      END AS client_reaction,
      CASE 
        WHEN reposter_user_id = $1 THEN true
        ELSE false
      END AS client_reposted,
      CASE 
        WHEN saver_user_id = $1 THEN true
        ELSE false
      END AS client_saved
    FROM "AllPostsView"
    WHERE saver_user_id = $1`,
    values: [client_user_id],
  }

  return (await dbQuery(query)).rows
}

/**
 * @param {number} user_id
 * @param {"online" | "offline"} connection_status
 */
export const updateUserConnectionStatus = async (
  user_id,
  connection_status
) => {
  /** @type {PgQueryConfig} */
  const query = {
    text: `
    UPDATE "User" SET connection_status = $1, last_active = $2 WHERE id = $3`,
    values: [
      connection_status,
      connection_status === "offline" ? new Date() : null,
      user_id,
    ],
  }

  await dbQuery(query)
}

// GET user notifications
/**
 *
 * @param {number} client_user_id
 * @param {Date} from
 */
export const getUnreadNotifications = async (client_user_id, from_date) => {
  const query = {
    text: `
    SELECT * 
    FROM "Notification" 
    WHERE receiver_user_id = $1 AND created_at >= $2`,
    values: [client_user_id, from_date],
  }

  return stripNulls((await dbQuery(query)).rows)
}

export const getUnreadNotificationsCount = async (client_user_id) => {
  const query = {
    text: `
    SELECT COUNT(id) AS count 
    FROM "Notification" 
    WHERE receiver_user_id = $1 AND is_read = false
    `,
    values: [client_user_id],
  }

  return (await dbQuery(query)).rows[0].count
}

export const getUserFolloweesIds = async (user_id) => {
  const query = {
    text: `
    SELECT followee_user_id
    FROM "Follow"
    WHERE follower_user_id = $1`,
    values: [user_id],
  }

  return (await dbQuery(query)).rows
}
