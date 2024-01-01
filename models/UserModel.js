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
 * @param {number} user_id
 * @param {number} to_follow_user_id
 */
export const followUser = async (user_id, to_follow_user_id) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `INSERT INTO "Follow" (follower_user_id, followee_user_id) VALUES ($1, $2)`,
    values: [user_id, to_follow_user_id],
  }

  await dbQuery(query)
}

/**
 * @param {number} follower_user_id
 * @param {number} followee_user_id
 */
export const unfollowUser = async (follower_user_id, followee_user_id) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `DELETE FROM "Follow" WHERE follower_user_id = $1 AND followee_user_id = $2`,
    values: [follower_user_id, followee_user_id],
  }

  await dbQuery(query)
}

/**
 * @param {number} user_id
 * @param {object} updatedUserInfoKVPairs
 */
export const updateUserProfile = async (user_id, updatedUserInfoKVPairs) => {
  const keys = Object.keys(updatedUserInfoKVPairs)
  const values = Object.values(updatedUserInfoKVPairs)

  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `UPDATE "User" SET ${keys
      .map((key, i) => `${key} = $${i + 1}`)
      .join(", ")} WHERE id = $${keys.length + 1}`,
    values: [...values, user_id],
  }

  await dbQuery(query)
}

export const uploadProfilePicture = async (user_id, profile_pic_url) => {
  /** @type {import("pg").QueryConfig} */
  const query = {
    text: `UPDATE "User" SET profile_pic_url = $1 WHERE id = $2`,
    values: [profile_pic_url, user_id],
  }

  await dbQuery(query)
}
