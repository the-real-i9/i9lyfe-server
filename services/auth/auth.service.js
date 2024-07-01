import bcrypt from "bcrypt"
import { generateJwtToken } from "../../utils/helpers.js"
import sendMail from "../mail.service.js"
import {
  changeUserPassword,
  createUser,
  signIn,
  userExists,
} from "../../models/user.model.js"

/** @param {string} password */
const hashPassword = async (password) => {
  return await bcrypt.hash(password, process.env.HASH_SALT || 10)
}

/**
 * @param {object} info
 * @param {string} info.email
 * @param {string} info.username
 * @param {string} info.password
 * @param {string} info.name
 * @param {Date} info.birthday
 * @param {string} info.bio
 */
export const userRegistrationService = async (info) => {
  if (await userExists(info.username)) {
    return {
      ok: false,
      error: {
        code: 422,
        msg: "Username already taken. Try another."
      },
      data: null,
    }
  }

  const passwordHash = await hashPassword(info.password)

  const userData = await createUser({
    ...info,
    password: passwordHash,
    birthday: new Date(info.birthday),
  })

  const jwtToken = generateJwtToken({
    client_user_id: userData.id,
    client_username: userData.username,
  })

  return {
    ok: true,
    error: null,
    data: { msg: "Signup success!", userData, jwtToken },
  }
}

/**
 * @param {string} emailOrUsername
 * @param {string} passwordInput
 */
export const userSigninService = async (emailOrUsername, passwordInput) => {
  const passwordInputHash = await hashPassword(passwordInput)

  const user = await signIn(emailOrUsername, passwordInputHash)

  if (!user) {
    return {
      ok: false,
      error: { code: 422, msg: "Incorrect email or password" },
      data: null,
    }
  }

  const jwtToken = generateJwtToken({
    client_user_id: user.id,
    client_username: user.username,
  })

  return {
    ok: true,
    error: null,
    data: {
      user, // observe that password has been excluded above
      jwtToken,
    },
  }
}

export const passwordResetService = async (userEmail, newPassword) => {
  const passwordHash = await hashPassword(newPassword)

  await changeUserPassword(userEmail, passwordHash)

  sendMail({
    to: userEmail,
    subject: "i9lyfe - Password reset successful",
    html: `<p>${userEmail}, your password has been changed successfully!</p>`,
  })

  return {
    ok: true,
    err: null,
    data: {
      msg: "Your password has been changed successfully"
    },
  }
}

/** @param {import("./emailConfirmationStrategy.auth.service.js").EmailConfirmationStrategy} emailConfirmationStrategy */
export const emailConfirmationService = (emailConfirmationStrategy) => {
  return {
    /**
     * @param {string} email
     * @returns {*} data
     */
    async handleEmailSubmission(email) {
      return await emailConfirmationStrategy.handleEmailSubmission(email)
    },

    /**
     * @param {number} inputCode
     * @param {*} sessionData
     * @returns {*} data
     */
    async handleCodeValidation(inputCode, sessionData) {
      return await emailConfirmationStrategy.handleCodeValidation(
        inputCode,
        sessionData
      )
    },
  }
}

/**
 * @param {import("socket.io").Socket} socket
 */
export const renewJwtToken = (socket) => {
  const { client_user_id, client_username } = socket.jwt_payload

  const newJwtToken = generateJwtToken({ client_user_id, client_username })

  socket.emit("renewed jwt", newJwtToken)
}
