import { userExists } from "../../models/user.model.js"
import {
  generateCodeWithExpiration,
  tokenLives,
  tokensMatch,
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

export class SignupEmailConfirmationStrategy extends EmailConfirmationStrategy {
  /**
   * @param {string} email
   * @returns {*} data
   */
  async handleEmailSubmission(email) {
    if (await userExists(email))
      return {
        ok: false,
        error: {
          code: 422,
          msg: "A user with this email already exists.",
        },
        data: null,
      }

    const [code, codeExpires] = generateCodeWithExpiration()

    sendMail({
      to: email,
      subject: "i9lyfe - Verify your email",
      html: `<p>Your email verification code is <strong>${code}</strong></p>`,
    })

    return {
      ok: true,
      error: null,
      data: {
        msg: `Enter the 6-digit code sent to ${email} to verify your email`,
        sessionData: {
          email,
          verified: false,
          verificationCode: code,
          verificationCodeExpires: codeExpires,
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
    const { email, verificationCode, verificationCodeExpires } = sessionData

    if (!tokensMatch(Number(verificationCode), Number(inputCode))) {
      return {
        ok: false,
        error: {
          code: 422,
          msg: "Incorrect verification code! Check or Re-submit your email.",
        },
        data: null,
      }
    }

    if (!tokenLives(verificationCodeExpires)) {
      return {
        ok: false,
        error: {
          code: 422,
          msg: "Verification code expired! Re-submit your email.",
        },
        data: null,
      }
    }

    sendMail({
      to: email,
      subject: "i9lyfe - Email verification success",
      html: `<p>Your email <strong>${email}</strong> has been verified!</p>`,
    })

    return {
      ok: true,
      error: null,
      data: {
        msg: `Your email ${email} has been verified!`,
        sessionData: {
          email,
          verified: true,
          verificationCode: null,
          verificationCodeExpires: null,
        },
      },
    }
  }
}

export class PasswordResetEmailConfirmationStrategy extends EmailConfirmationStrategy {
  /**
   * @param {string} email
   * @returns {*} data
   */
  async handleEmailSubmission(email) {
    if (!(await userExists(email)))
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
    

    if (!tokensMatch(passwordResetToken, inputCode)) {
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
