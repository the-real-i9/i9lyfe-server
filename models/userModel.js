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
  try {
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
  } catch (error) {
    console.log(error)
    throw error
  }
}

/**
 * @param {string} email
 * @param {string} selectFields
 */
export const getUserByEmail = async (email, selectFields) => {
  try {
    /** @type {import("pg").QueryConfig} */
    const query = {
      text: `SELECT ${commaSeparateString(
        selectFields
      )} FROM "User" WHERE email = $1`,
      values: [email],
    }

    const result = await dbQuery(query)

    return result
  } catch (error) {
    console.log(error)
    throw error
  }
}

/** @param {string} email */
export const userExists = async (email) => {
  try {
    const result = await getUserByEmail(email, "1")
    return result.rowCount > 0 ? true : false
  } catch (error) {
    console.log(error)
    throw error
  }
}

export const changeUserPassword = async (email, newPassword) => {
  try {
    /** @type {import("pg").QueryConfig} */
    const query = {
      text: 'UPDATE "User" SET password = $2 WEHRE email = $1',
      values: [email, newPassword]
    }
  
    await dbQuery(query)
  } catch (error) {
    console.log(error)
    throw error
  }
}
