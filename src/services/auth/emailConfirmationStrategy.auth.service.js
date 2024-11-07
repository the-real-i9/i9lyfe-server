import { User } from "../../models/user.model.js"
import {
  generateCodeWithExpiration,
  tokenLives,
} from "../../utils/helpers.js"
import sendMail from "../mail.service.js"


export class EmailConfirmationStrategy {
  /**
   * @param {string} email
   * @returns {*} data
   */
  async handleEmailSubmission(email) {
    throw new Error("handleEmailSubmission must be implemented")
  }

  /**
   * @param {number} inputCode
   * @param {*} sessionData
   * @returns {*} data
   */
  async handleCodeValidation(inputCode, sessionData) {
    throw new Error("handleCodeValidation must be implemented")
  }
}

export class PasswordResetEmailConfirmationStrategy extends EmailConfirmationStrategy {
  /**
   * @param {string} email
   * @returns {*} data
   */
  async handleEmailSubmission(email) {
    if (!(await User.exists(email)))
      return {
        ok: false,
        error: {
          code: 422,
          msg: "No user with this email exists.",
        },
        data: null,
      }

    const [token, tokenExpires] = generateCodeWithExpiration()

    sendMail({
      to: email,
      subject: "i9lyfe - Confirm your email: Password Reset",
      html: `<p>Your password reset token is <strong>${token}</strong>.</p>`,
    })

    return {
      ok: true,
      error: null,
      data: {
        msg: `Enter the 6-digit number token sent to ${email} to reset your password`,
        sessionData: {
          email,
          emailConfirmed: false,
          passwordResetToken: token,
          passwordResetTokenExpires: tokenExpires,
        },
      },
    }
  }

  /**
   * @param {number} inputCode
   * @param {*} sessionData
   * @returns {*} data
   */
  async handleCodeValidation(inputCode, sessionData) {
    const { email, passwordResetToken, passwordResetTokenExpires } =
      sessionData
    

    if (passwordResetToken !== inputCode) {
      return {
        ok: false,
        error: {
          code: 422,
          msg:
            "Incorrect password reset token! Check or Re-submit your email.",
        },
        data: null,
      }
    }

    if (!tokenLives(passwordResetTokenExpires)) {
      return {
        ok: false,
        error: {
          code: 422,
          msg: "Password reset token expired! Re-submit your email.",
        },
        data: null
      }
    }


    return {
      ok: true,
      error: null,
      data: {
        msg: `Your email ${email} has been verified!`,
        sessionData: {
          email,
          emailConfirmed: true,
          passwordResetToken: null,
          passwordResetTokenExpires: null,
        }
      }
    }
  }
}
