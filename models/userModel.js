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
      text: 'INSERT INTO "User"(email, username, password, name, birthday, bio) VALUES($1, $2, $3, $4, $5, $6) ',
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

// const getUserById = (id) => {}

/**
 * @param {string} email
 * @param {string} fieldString
 */
export const getUserByEmail = async (email, colsString) => {
  try {
    /** @type {import("pg").QueryConfig} */
    const query = {
      text: `SELECT ${commaSeparateString(
        colsString
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

// const getUserByUsername = (username) => {}
