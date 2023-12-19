import { commaSeparateString } from "../utils/helpers.js"
import { dbQuery } from "./db.js"

const createNewUser = (fields) => {}

const getUserById = (id) => {}

/**
 * @param {string} email
 * @param {string} fieldString
 */
export const getUserByEmail = async (email, colsString) => {
  try {
    /** @type {import("pg").QueryConfig} */
    const query = {
      text: `SELECT ${commaSeparateString(colsString)} FROM "User" WHERE email = $1`,
      values: [email],
    }

    const result = await dbQuery(query)

    return result
  } catch (error) {
    // console.log(error)
  }
}

const getUserByUsername = (username) => {}
