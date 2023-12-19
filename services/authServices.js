import { getUserByEmail } from "../models/userModel.js"

const userAlreadyExists = async (email) => {
  const user = await getUserByEmail(email, "1")
  return user ? true : false
}

export const newAccountRequestService = async (email) => {
  // check if a user with email exists, if yes return { ok: false, err: { code: 422, reason: 'A user with the email already exists' } }; else
  // generate a 6-digit verification code and send to the email
  // create a potential_users session for the email, along with verification data
  // return {ok: true, err: null}
  try {
    if (await userAlreadyExists(email)) {
      return {
        statusCode: 422,
        statusMessage: "A user with this email already exists",
      }
    }
  } catch (error) {
    console.log(error)
  }
}

export const emailVerificationService = (email) => {}
