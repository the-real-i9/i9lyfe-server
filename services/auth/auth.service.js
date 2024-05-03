import bcrypt from "bcrypt"
import { generateJwtToken } from "../../utils/helpers.js"
import sendMail from "../mail.service.js"
import { changeUserPassword, createUser, getUser } from "../../models/user.model.js"

/** @typedef {import('express').Request} ExpRequest */
/** @typedef {import('express').Response} ExpResponse */

/** @param {string} password */
const hashPassword = async (password) => {
  return await bcrypt.hash(password, 10)
}

/**
 * @param {string} passwordInput Password supplied by user
 * @param {string} passwordHash The hashed version stored in database
 * @returns The compare result; true if both match and false otherwise
 */
const passwordMatch = async (passwordInput, passwordHash) => {
  return await bcrypt.compare(passwordInput, passwordHash)
}

/**
 * @param {Object} userDataInput
 * @param {string} userDataInput.email
 * @param {string} userDataInput.username
 * @param {string} userDataInput.password
 * @param {string} userDataInput.name
 * @param {Date} userDataInput.birthday
 * @param {string} userDataInput.bio
 */
export const userRegistrationService = async (userDataInput) => {
  const passwordHash = await hashPassword(userDataInput.password)

  const userData = await createUser({
    ...userDataInput,
    password: passwordHash,
    birthday: new Date(userDataInput.birthday),
  })

  const jwtToken = generateJwtToken({
    client_user_id: userData.id,
    client_username: userData.username,
  })

  return {
    ok: true,
    error: null,
    data: { userData, jwtToken },
  }
}

/**
 * @param {string} email
 * @param {string} passwordInput
 */
export const userSigninService = async (email, passwordInput) => {
  const user = await getUser(email, true)

  if (!user) {
    return {
      ok: false,
      err: { code: 422, reason: "Incorrect email or password" },
      data: null,
    }
  }

  const { password: passwordHash, ...userData } = user
  if (!(await passwordMatch(passwordInput, passwordHash))) {
    return {
      ok: false,
      err: { code: 422, reason: "Incorrect email or password" },
      data: null,
    }
  }

  const jwtToken = generateJwtToken({
    client_user_id: userData.id,
    client_username: userData.username,
  })
  return {
    ok: true,
    err: null,
    data: {
      userData, // observe that password has been excluded above
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
    data: null,
  }
}

/** @param {import("./emailConfirmationStrategy.auth.service.js").EmailConfirmationStrategy} emailConfirmationStrategy */
export const emailConfirmationService = (emailConfirmationStrategy) => {
  return {
    /** @param {ExpRequest} req */
    async handleEmailSubmission(req) {
      return await emailConfirmationStrategy.handleEmailSubmission(req)
    },

    /** @param {ExpRequest} req */
    async handleTokenValidation(req) {
      return await emailConfirmationStrategy.handleTokenValidation(req)
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
