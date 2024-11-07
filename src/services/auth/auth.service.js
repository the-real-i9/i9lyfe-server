import bcrypt from "bcrypt"
import { generateJwt } from "../../utils/helpers.js"
import sendMail from "../mail.service.js"
import { User } from "../../models/user.model.js"



/**
 * @param {string} emailOrUsername
 * @param {string} passwordInput
 */
export const userSigninService = async (emailOrUsername, passwordInput) => {
  const userData = await User.findOneForAuth(emailOrUsername)
  
  if (!userData) {
    return {
      ok: false,
      error: { code: 422, msg: "Incorrect email or password" },
      data: null,
    }
  }

  const { pswd: storedPswd, ...user } = userData

  if (!(await bcrypt.compare(passwordInput, storedPswd))) {
    return {
      ok: false,
      error: { code: 422, msg: "Incorrect email or password" },
      data: null,
    }
  }

  const jwt = generateJwt({
    client_user_id: user.id,
    client_username: user.username,
  })

  return {
    ok: true,
    error: null,
    data: {
      msg: "Signin success!",
      user,
      jwt,
    },
  }
}

export const passwordResetService = async (userEmail, newPassword) => {
  const passwordHash = await hashPassword(newPassword)

  await User.changePassword(userEmail, passwordHash)

  sendMail({
    to: userEmail,
    subject: "i9lyfe - Password reset successful",
    html: `<p>${userEmail}, your password has been changed successfully!</p>`,
  })

  return {
    ok: true,
    err: null,
    data: {
      msg: "Your password has been changed successfully",
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

  const newJwtToken = generateJwt({ client_user_id, client_username })

  socket.emit("renewed jwt", newJwtToken)
}
