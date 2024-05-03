import { userExists } from "../../models/user.model.js"
import {
  generateCodeWithExpiration,
  tokenLives,
  tokensMatch,
} from "../../utils/helpers.js"
import sendMail from "../mail.service.js"

/**
 * Stragegy pattern
 * @interface
 * @abstract
 */
export class EmailConfirmationStrategy {
  /** @param {import('express').Request} req */
  async handleEmailSubmission(req) {
    throw new Error("handleEmailSubmission must be implemented")
  }

  /** @param {import('express').Request} req */
  async handleTokenValidation(req) {
    throw new Error("handleTokenValidation must be implemented")
  }
}

export class SignupEmailConfirmationStrategy extends EmailConfirmationStrategy {
  /** @param {import('express').Request} req */
  async handleEmailSubmission(req) {
    const { email } = req.body

    if (await userExists(email))
      return {
        ok: false,
        err: {
          code: 422,
          reason: "A user with this email already exists.",
        },
        successMessage: null,
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
      err: null,
      successMessage: `Enter the 6-digit code sent to ${email} to verify your email`,
    }
  }

  /** @param {import('express').Request} req */
  async handleTokenValidation(req) {
    const { email, verificationCode, verificationCodeExpires } =
      req.session.email_verification_data
    const { code: userInputCode } = req.body

    if (!tokensMatch(Number(verificationCode), Number(userInputCode))) {
      return {
        ok: false,
        err: {
          code: 422,
          reason: "Incorrect verification code! Check or Re-submit your email.",
        },
        successMessage: null,
      }
    }

    if (!tokenLives(verificationCodeExpires)) {
      return {
        ok: false,
        err: {
          code: 422,
          reason: "Verification code expired! Re-submit your email.",
        },
        successMessage: null,
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
      err: null,
      successMessage: `Your email ${email} has been verified!`,
    }
  }
}

export class PasswordResetEmailConfirmationStrategy extends EmailConfirmationStrategy {
  /** @param {import('express').Request} req */
  async handleEmailSubmission(req) {
    const { email } = req.body

    if (!(await userExists(email)))
      return {
        ok: false,
        err: {
          code: 422,
          reason: "No user with this email exists.",
        },
        successMessage: null,
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
      successMessage: `Enter the 6-digit number token sent to ${email} to reset your password`,
    }
  }

  /** @param {import('express').Request} req */
  async handleTokenValidation(req) {
    const { email, passwordResetToken, passwordResetTokenExpires } =
      req.session.password_reset_email_confirmation_data
    const { token: userInputToken } = req.body

    if (!tokensMatch(passwordResetToken, userInputToken)) {
      return {
        ok: false,
        err: {
          code: 422,
          reason:
            "Incorrect password reset token! Check or Re-submit your email.",
        },
        successMessage: null,
      }
    }

    if (!tokenLives(passwordResetTokenExpires)) {
      return {
        ok: false,
        err: {
          code: 422,
          reason: "Password reset token expired! Re-submit your email.",
        },
        successMessage: null,
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
      err: null,
      successMessage: `Your email ${email} has been verified!`,
    }
  }
}


