import * as mailService from "../mail.service.js"
import * as authUtilServices from "../utils/auth.utilServices.js"
import { User } from "../../models/user.model.js"

export const requestPasswordReset = async (email) => {
  if (!(await User.exists(email)))
    return {
      error: { msg: "No user with this email exists." },
    }

  const { token: passwordResetToken, expires: passwordResetTokenExpires } =
    authUtilServices.generateTokenWithExpiration()

  mailService.sendMail({
    to: email,
    subject: "i9lyfe - Confirm your email: Password Reset",
    html: `<p>Your password reset token is <strong>${passwordResetToken}</strong>.</p>`,
  })

  return {
    passwordResetToken,
    passwordResetTokenExpires,
    data: {
      msg: `Enter the 6-digit number token sent to ${email} to reset your password`,
    },
  }
}

export const confirmEmail = ({
  email,
  inputToken,
  passwordResetToken,
  passwordResetTokenExpires,
}) => {
  if (passwordResetToken !== Number(inputToken)) {
    return {
      error: {
        msg: "Incorrect password reset token! Check or Re-submit your email.",
      },
    }
  }

  if (!authUtilServices.isTokenAlive(passwordResetTokenExpires)) {
    return {
      error: { msg: "Password reset token expired! Re-submit your email." },
    }
  }

  return {
    data: { msg: `Your email ${email} has been verified!` },
  }
}

export const resetPassword = async (email, newPassword) => {
  const passwordHash = await authUtilServices.hashPassword(newPassword)

  await User.changePassword(email, passwordHash)

  mailService.sendMail({
    to: email,
    subject: "i9lyfe - Password reset successful",
    html: `<p>${email}, your password has been changed successfully!</p>`,
  })

  return {
    data: { msg: "Your password has been changed successfully" },
  }
}
