import { userExists } from "../models/userModel"
import { generateCodeWithExpiration } from "../utils/helpers"
import sendMail from "./mailingService"

/**
 * Stragegy pattern + Template Method pattern
 * @interface
 * @abstract
 */
class EmailConfirmationStrategy {
  async handleEmailSubmission() {
    throw new Error("handleEmailSubmission must be implemented")
  }

  async handleTokenValidation() {
    throw new Error("handleTokenValidation must be implemented")
  }
}

export class SignupEmailConfirmationStrategy extends EmailConfirmationStrategy {
  /** @param {import('express').Request} req */
  async handleEmailSubmission(req) {
    try {
      const { email } = req.body

      if (await userExists(email))
        return {
          ok: false,
          error: {
            code: 422,
            msg: "A user with this email already exists.",
          },
          success: null,
        }

      const [code, codeExpires] = generateCodeWithExpiration()

      req.session.email_verification_data = {
        email,
        verified: false,
        verificationCode: code,
        verificationCodeExpires: codeExpires,
      }

      sendMail({
        to: email,
        subject: "i9lyfe - Verify your email",
        html: `<p>Your email verification code is <strong>${code}</strong></p>`,
      })

      return {
        ok: true,
        error: null,
        success: {
          code: 200,
          msg: `Enter the 6-digit code sent to ${email} to verify your email`,
        },
      }
    } catch (error) {
      console.log(error)
      return {
        ok: false,
        error: {
          code: 500,
          msg: "Internal server error",
        },
        success: null,
      }
    }
  }

  /** @param {import('express').Request} req */
  async handleTokenValidation(req) {
    try {
      const { email, verificationCode, verificationCodeExpires } =
        req.session.email_verification_data
      const { code: userInputCode } = req.body

      if (!tokensMatch(verificationCode, userInputCode)) {
        return {
          ok: false,
          error: {
            code: 422,
            msg: "Incorrect verification code! Check or Re-submit your email.",
          },
          success: null,
        }
      }

      if (!tokenLives(verificationCodeExpires)) {
        return {
          ok: false,
          error: {
            code: 422,
            msg: "Verification code expired! Re-submit your email.",
          },
          success: null,
        }
      }

      req.session.email_verification_data = {
        email,
        verified: true,
        verificationCode: null,
        verificationCodeExpires: null,
      }

      sendMail({
        to: email,
        subject: "i9lyfe - Email verification success",
        html: `<p>Your email <strong>${email}</strong> has been verified!</p>`,
      })

      return {
        ok: true,
        error: null,
        success: {
          code: 200,
          msg: `Your email ${email} has been verified!`,
        },
      }
    } catch (error) {
      console.log(error)
      return {
        ok: false,
        error: {
          code: 500,
          msg: "Internal server error",
        },
        success: null,
      }
    }
  }
}


export class PasswordResetEmailConfirmationStrategy extends EmailConfirmationStrategy {
  /** @param {import('express').Request} req */
  async handleEmailSubmission(req) {
    try {
      const { email } = req.body

      if (!(await userExists(email)))
        return {
          ok: false,
          error: {
            code: 422,
            msg: "No user with this email exists.",
          },
          success: null,
        }

      const [token, tokenExpires] = generateCodeWithExpiration()

      req.session.password_reset_email_confirmation_data = {
        email,
        emailConfirmed: false,
        passwordResetToken: token,
        passwordResetTokenExpires: tokenExpires,
      }

      sendMail({
        to: email,
        subject: "i9lyfe - Confirm your email: Password Reset",
        html: `<p>Your password reset token is <strong>${token}</strong>.</p>`,
      })

      return {
        ok: true,
        error: null,
        success: {
          code: 200,
          msg: `Enter the 6-digit number token sent to ${email} to reset your password`,
        },
      }
    } catch (error) {
      console.log(error)
      return {
        ok: false,
        error: {
          code: 500,
          msg: "Internal server error",
        },
        success: null,
      }
    }
  }

  /** @param {import('express').Request} req */
  async handleTokenValidation(req) {
    try {
      const { email, passwordResetToken, passwordResetTokenExpires } =
        req.session.password_reset_email_confirmation_data
      const { token: userInputToken } = req.body

      if (!tokensMatch(passwordResetToken, userInputToken)) {
        return {
          ok: false,
          error: {
            code: 422,
            msg: "Incorrect password reset token! Check or Re-submit your email.",
          },
          success: null,
        }
      }

      if (!tokenLives(passwordResetTokenExpires)) {
        return {
          ok: false,
          error: {
            code: 422,
            msg: "Password reset token expired! Re-submit your email.",
          },
          success: null,
        }
      }

      req.session.password_reset_email_confirmation_data = {
        email,
        emailConfirmed: true,
        passwordResetToken: null,
        passwordResetTokenExpires: null,
      }

      return {
        ok: true,
        error: null,
        success: {
          code: 200,
          msg: `Your email ${email} has been verified!`,
        },
      }
    } catch (error) {
      console.log(error)
      return {
        ok: false,
        error: {
          code: 500,
          msg: "Internal server error",
        },
        success: null,
      }
    }
  }
}


/**
 * @param {number} compareToken
 * @param {number} inputToken
 */
const tokensMatch = (compareToken, inputToken) => compareToken === inputToken

/** @param {Date} tokenExpiration */
const tokenLives = (tokenExpiration) => Date.now() < new Date(tokenExpiration)
