import { User } from "../../models/user.model.js"
import * as authServices from "../../services/auth.services.js"
import * as mailService from "../../services/mail.service.js"

export const requestPasswordReset = async (req, res) => {
  const { email } = req.body
  try {
    if (!(await User.exists(email)))
      return res.status(422).send({ msg: "No user with this email exists." })

    const [token, tokenExpires] = authServices.generateCodeWithExpiration()

    mailService.sendMail({
      to: email,
      subject: "i9lyfe - Confirm your email: Password Reset",
      html: `<p>Your password reset token is <strong>${token}</strong>.</p>`,
    })

    req.session.password_reset_email_confirmation_state = {
      email,
      emailConfirmed: false,
      passwordResetToken: token,
      passwordResetTokenExpires: tokenExpires,
    }

    res
      .status(200)
      .send({
        msg: `Enter the 6-digit number token sent to ${email} to reset your password`,
      })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const confirmEmail = async (req, res) => {
  const { code } = req.body

  try {
    const { email, passwordResetToken, passwordResetTokenExpires } =
      req.session.password_reset_email_confirmation_state

    if (passwordResetToken !== Number(code)) {
      return res
        .status(422)
        .send({
          msg: "Incorrect password reset token! Check or Re-submit your email.",
        })
    }

    if (!authServices.isTokenAlive(passwordResetTokenExpires)) {
      return res
        .status(422)
        .send({ msg: "Password reset token expired! Re-submit your email." })
    }

    req.session.password_reset_email_confirmation_state = {
      email,
      emailConfirmed: true,
      passwordResetToken: null,
      passwordResetTokenExpires: null,
    }

    res.status(200).send({ msg: `Your email ${email} has been verified!` })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const resetPassword = async (req, res) => {
  try {
    const { email } = req.session.password_reset_email_confirmation_state

    const { newPassword } = req.body

    const passwordHash = await authServices.hashPassword(newPassword)

    await User.changePassword(email, passwordHash)

    mailService.sendMail({
      to: email,
      subject: "i9lyfe - Password reset successful",
      html: `<p>${email}, your password has been changed successfully!</p>`,
    })

    req.session.destroy()

    res.status(200).send({ msg: "Your password has been changed successfully" })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
