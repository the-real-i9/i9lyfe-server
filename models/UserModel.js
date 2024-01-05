import { commaSeparateString } from "../utils/helpers.js"
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
 * @param {number} client_user_id
 * @param {number} to_follow_user_id
 */
export const followUser = async (client_user_id, to_follow_user_id) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `INSERT INTO "Follow" (follower_user_id, followee_user_id) VALUES ($1, $2)`,
    values: [client_user_id, to_follow_user_id],
  }

  await dbQuery(query)
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
 * @param {object} updatedUserInfoKVPairs
 */
export const updateUserProfile = async (client_user_id, updatedUserInfoKVPairs) => {
  const keys = Object.keys(updatedUserInfoKVPairs)
  const values = Object.values(updatedUserInfoKVPairs)

  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `UPDATE "User" SET ${keys
      .map((key, i) => `${key} = $${i + 1}`)
      .join(", ")} WHERE id = $${
      keys.length + 1
    } RETURNING id, email, username, name, profile_pic_url`,
    values: [...values, client_user_id],
  }

  return await dbQuery(query)
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
    SELECT "user".id AS owner_user_id,
      "user".username AS owner_username,
      "user".profile_pic_url AS owner_profile_pic_url,
      "post".id AS post_id,
      type,
      media_urls,
      description,
      COUNT(DISTINCT "any_reaction".id)::INTEGER AS reactions_count,
      COUNT(DISTINCT "any_comment".id)::INTEGER AS comments_count, 
      COUNT(DISTINCT "any_repost".id)::INTEGER AS reposts_count,
      COUNT(DISTINCT "any_saved_post".id)::INTEGER AS saves_count,
      "client_reaction".reaction_code_point AS client_reaction,
      CASE
        WHEN "client_repost".id IS NULL THEN false
        ELSE true
      END AS client_reposted,
      CASE
        WHEN "client_saved_post".id IS NULL THEN false
        ELSE true
      END AS client_saved
    FROM "Post" "post"
    INNER JOIN "User" "user" ON "user".id = "post".user_id
    LEFT JOIN "PostCommentReaction" "any_reaction" ON "any_reaction".post_id = "post".id 
    LEFT JOIN "Comment" "any_comment" ON "any_comment".post_id = "post".id
    LEFT JOIN "Repost" "any_repost" ON "any_repost".post_id = "post".id
    LEFT JOIN "SavedPost" "any_saved_post" ON "any_saved_post".post_id = "post".id
    LEFT JOIN "PostCommentReaction" "client_reaction" 
      ON "client_reaction".post_id = "post".id AND "client_reaction".reactor_user_id = $2
    LEFT JOIN "Repost" "client_repost" 
      ON "client_repost".post_id = "post".id AND "client_repost".reposter_user_id = $2
    LEFT JOIN "SavedPost" "client_saved_post" 
      ON "client_saved_post".post_id = "post".id AND "client_saved_post".saver_user_id = $2
    WHERE "user".username = $1
    GROUP BY owner_user_id, 
      owner_username, 
      owner_profile_pic_url, 
      "post".id, 
      type, 
      media_urls, 
      description, 
      client_reaction, 
      client_reposted,
      client_saved`,
    values: [username, client_user_id],
  }

  return (await dbQuery(query)).rows
}

// GET posts user has been mentioned in
export const getMentionedPosts = async (client_user_id) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `
    SELECT "user".id AS owner_user_id,
      "user".username AS owner_username,
      "user".profile_pic_url AS owner_profile_pic_url,
      "post".id AS post_id,
      type,
      media_urls,
      description,
      COUNT(DISTINCT "any_reaction".id)::INTEGER AS reactions_count,
      COUNT(DISTINCT "any_comment".id)::INTEGER AS comments_count, 
      COUNT(DISTINCT "any_repost".id)::INTEGER AS reposts_count,
      COUNT(DISTINCT "any_saved_post".id)::INTEGER AS saves_count,
      "client_reaction".reaction_code_point AS client_reaction,
      CASE
        WHEN "client_repost".id IS NULL THEN false
        ELSE true
      END AS client_reposted,
      CASE
        WHEN "client_saved_post".id IS NULL THEN false
        ELSE true
      END AS client_saved
    FROM "Post" "post"
    INNER JOIN "User" "user" ON "user".id = "post".user_id
    LEFT JOIN "PostCommentReaction" "any_reaction" ON "any_reaction".post_id = "post".id 
    LEFT JOIN "Comment" "any_comment" ON "any_comment".post_id = "post".id
    LEFT JOIN "Repost" "any_repost" ON "any_repost".post_id = "post".id
    LEFT JOIN "SavedPost" "any_saved_post" ON "any_saved_post".post_id = "post".id
    LEFT JOIN "PostCommentReaction" "client_reaction" 
      ON "client_reaction".post_id = "post".id AND "client_reaction".reactor_user_id = $1
    LEFT JOIN "Repost" "client_repost" 
      ON "client_repost".post_id = "post".id AND "client_repost".reposter_user_id = $1
    LEFT JOIN "SavedPost" "client_saved_post" 
      ON "client_saved_post".post_id = "post".id AND "client_saved_post".saver_user_id = $1
    INNER JOIN "PostCommentMention" "mention" ON "mention".post_id = "post".id AND "mention".user_id = $1
    GROUP BY owner_user_id, 
      owner_username, 
      owner_profile_pic_url, 
      "post".id, 
      type, 
      media_urls, 
      description, 
      client_reaction, 
      client_reposted,
      client_saved`,
    values: [client_user_id],
  }

  return (await dbQuery(query)).rows
}

// GET posts reacted by user
export const getReactedPosts = async (client_user_id) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `
    SELECT "user".id AS owner_user_id,
      "user".username AS owner_username,
      "user".profile_pic_url AS owner_profile_pic_url,
      "post".id AS post_id,
      type,
      media_urls,
      description,
      COUNT(DISTINCT "any_reaction".id)::INTEGER AS reactions_count,
      COUNT(DISTINCT "any_comment".id)::INTEGER AS comments_count, 
      COUNT(DISTINCT "any_repost".id)::INTEGER AS reposts_count,
      COUNT(DISTINCT "any_saved_post".id)::INTEGER AS saves_count,
      "client_reaction".reaction_code_point AS client_reaction,
      CASE
        WHEN "client_repost".id IS NULL THEN false
        ELSE true
      END AS client_reposted,
      CASE
        WHEN "client_saved_post".id IS NULL THEN false
        ELSE true
      END AS client_saved
    FROM "Post" "post"
    INNER JOIN "User" "user" ON "user".id = "post".user_id
    LEFT JOIN "PostCommentReaction" "any_reaction" ON "any_reaction".post_id = "post".id 
    LEFT JOIN "Comment" "any_comment" ON "any_comment".post_id = "post".id
    LEFT JOIN "Repost" "any_repost" ON "any_repost".post_id = "post".id
    LEFT JOIN "SavedPost" "any_saved_post" ON "any_saved_post".post_id = "post".id
    LEFT JOIN "Repost" "client_repost" 
      ON "client_repost".post_id = "post".id AND "client_repost".reposter_user_id = $1
    LEFT JOIN "SavedPost" "client_saved_post" 
      ON "client_saved_post".post_id = "post".id AND "client_saved_post".saver_user_id = $1
    INNER JOIN "PostCommentReaction" "client_reaction" 
      ON "client_reaction".post_id = "post".id AND "client_reaction".reactor_user_id = $1
    GROUP BY owner_user_id, 
      owner_username, 
      owner_profile_pic_url, 
      "post".id, 
      type, 
      media_urls, 
      description, 
      client_reaction, 
      client_reposted,
      client_saved`,
    values: [client_user_id],
  }

  return (await dbQuery(query)).rows
}

// GET posts saved by this user
export const getSavedPosts = async (client_user_id) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `
    SELECT "user".id AS owner_user_id,
      "user".username AS owner_username,
      "user".profile_pic_url AS owner_profile_pic_url,
      "post".id AS post_id,
      type,
      media_urls,
      description,
      COUNT(DISTINCT "any_reaction".id)::INTEGER AS reactions_count,
      COUNT(DISTINCT "any_comment".id)::INTEGER AS comments_count, 
      COUNT(DISTINCT "any_repost".id)::INTEGER AS reposts_count,
      COUNT(DISTINCT "any_saved_post".id)::INTEGER AS saves_count,
      "client_reaction".reaction_code_point AS client_reaction,
      CASE
        WHEN "client_repost".id IS NULL THEN false
        ELSE true
      END AS client_reposted,
      CASE
        WHEN "client_saved_post".id IS NULL THEN false
        ELSE true
      END AS client_saved
    FROM "Post" "post"
    INNER JOIN "User" "user" ON "user".id = "post".user_id
    LEFT JOIN "PostCommentReaction" "any_reaction" ON "any_reaction".post_id = "post".id 
    LEFT JOIN "Comment" "any_comment" ON "any_comment".post_id = "post".id
    LEFT JOIN "Repost" "any_repost" ON "any_repost".post_id = "post".id
    LEFT JOIN "SavedPost" "any_saved_post" ON "any_saved_post".post_id = "post".id
    LEFT JOIN "PostCommentReaction" "client_reaction" 
      ON "client_reaction".post_id = "post".id AND "client_reaction".reactor_user_id = $1
    LEFT JOIN "Repost" "client_repost" 
      ON "client_repost".post_id = "post".id AND "client_repost".reposter_user_id = $1
    INNER JOIN "SavedPost" "client_saved_post" ON "client_saved_post".post_id = "post".id AND "client_saved_post".saver_user_id = $1
    GROUP BY owner_user_id, 
      owner_username, 
      owner_profile_pic_url, 
      "post".id, 
      type, 
      media_urls, 
      description, 
      client_reaction, 
      client_reposted,
      client_saved`,
    values: [client_user_id],
  }

  return (await dbQuery(query)).rows
}

// GET user notifications
export const getNotifications = async (/* client_user_id */) => {}
