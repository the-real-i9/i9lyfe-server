import { User } from "../../models/user.model"
import { generateCodeWithExpiration, hashPassword, tokenLives } from "../../utils/helpers"
import sendMail from "../mail.service"

export const requestPasswordReset = async (email) => {
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

export const confirmEmail = async (inputCode, sessionData) => {
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

export const resetPassword = async (userEmail, newPassword) => {
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