import * as passwordResetService from "../../services/auth/passwordReset.service.js"

export const requestPasswordReset = async (req, res) => {
  const { email } = req.body
  try {
    const resp = await passwordResetService.requestPasswordReset(email)

    if (resp.error) return res.status(400).send(resp.error)

    req.session.passwordReset = {
      email,
      emailConfirmed: false,
      passwordResetToken: resp.passwordResetToken,
      passwordResetTokenExpires: resp.passwordResetTokenExpires,
    }

    req.session.cookie.maxAge = 60 * 60 * 1000
    req.session.cookie.path = "/api/auth/forgot_password/confirm_email"

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const confirmEmail = async (req, res) => {
  const { token: inputToken } = req.body

  try {
    if (!req.session?.passwordReset) return res.sendStatus(401)

    const passwordResetSessionData = req.session.passwordReset

    const resp = passwordResetService.confirmEmail({
      inputToken,
      ...passwordResetSessionData,
    })

    if (resp.error) return res.status(400).send(resp.error)

    req.session.passwordReset = {
      email: passwordResetSessionData.email,
      emailConfirmed: true,
    }

    req.session.cookie.maxAge = 60 * 60 * 1000
    req.session.cookie.path = "/api/auth/forgot_password/reset_password"

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const resetPassword = async (req, res) => {
  try {
    if (!req.session?.passwordReset) return res.sendStatus(401)

    const { email } = req.session.passwordReset

    const { newPassword } = req.body

    const resp = await passwordResetService.resetPassword(email, newPassword)

    req.session.destroy()

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
