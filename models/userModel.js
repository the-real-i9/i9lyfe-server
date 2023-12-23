import { commaSeparateString } from "../utils/helpers.js"
import { dbQuery } from "./db.js"

/**
 * @param {Object} fields
 * @param {string} fields.userMail
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
      text: `INSERT INTO "User"(userMail, username, password, name, birthday, bio) 
      VALUES($1, $2, $3, $4, $5, $6) 
      RETURNING id, userMail, username, name, profile_pic_url`,
      values: [
        fields.userMail,
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
 * @param {string} userMail
 * @param {string} selectFields
 */
export const getUserByEmail = async (userMail, selectFields) => {
  try {
    /** @type {import("pg").QueryConfig} */
    const query = {
      text: `SELECT ${commaSeparateString(
        selectFields
      )} FROM "User" WHERE userMail = $1`,
      values: [userMail],
    }

    const result = await dbQuery(query)

    return result
  } catch (error) {
    console.log(error)
    throw error
  }
}

/** @param {string} userMail */
export const userExists = async (userMail) => {
  try {
    const result = await getUserByEmail(userMail, "1")
    return result.rowCount > 0 ? true : false
  } catch (error) {
    console.log(error)
    throw error
  }
}

export const changeUserPassword = async (userMail, newPassword) => {
  try {
    /** @type {import("pg").QueryConfig} */
    const query = {
      text: 'UPDATE "User" SET password = $2 WEHRE email = $1',
      values: [userMail, newPassword]
    }
  
    await dbQuery(query)
  } catch (error) {
    console.log(error)
    throw error
  }
}
