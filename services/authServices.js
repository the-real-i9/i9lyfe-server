import { getUserByEmail } from "../models/userModel.js"

/** @param {string} email */
export const userAlreadyExists = async (email) => {
  const result = await getUserByEmail(email, "1")
  return result.rowCount > 0 ? true : false
}

/** @param {string} email */
export const newAccountRequestService = async (email) => {
  // check if a user with email exists, if yes return { ok: false, err: { code: 422, reason: 'A user with the email already exists' } }; else
  // generate a 6-digit verification code and send to the email
  // create a potential_users session for the email, along with verification data
  // return {ok: true, err: null}
  try {
    if (await userAlreadyExists(email))
      return {
        statusCode: 422,
        statusMessage: "A user with this email already exists",
      }
    return {
      statusCode: 200,
      statusMessage: `Enter the 6-digit code sent to ${email} from i9apps`,
    }
  } catch (error) {
    console.log(error)
    return {statusCode: 500, statusMessage: "Server Error"}
  }
}

// export const emailVerificationService = (email) => {}
