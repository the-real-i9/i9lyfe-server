import * as passwordResetService from "../../services/auth/passwordReset.service.js"

export const requestPasswordReset = async (req, res) => {
  const { email } = req.body
  try {
    const resp = await passwordResetService.requestPasswordReset(email)

    if (resp.error) return res.status(400).send(resp.error)

    req.session.passwordReset = {
      step: "confirm email",
      data: {
        email,
        emailConfirmed: false,
        passwordResetToken: resp.passwordResetToken,
        passwordResetTokenExpires: resp.passwordResetTokenExpires,
      },
    }

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const confirmEmail = async (req, res) => {
  const { token: inputToken } = req.body

  try {
    if (req.session?.passwordReset.step != "confirm email")
      return res.status(400).send({ msg: "Invalid cookie at endpoint" })

    const passwordResetSessionData = req.session.passwordReset.data

    const resp = passwordResetService.confirmEmail({
      inputToken,
      ...passwordResetSessionData,
    })

    if (resp.error) return res.status(400).send(resp.error)

    req.session.passwordReset = {
      step: "reset password",
      data: {
        email: passwordResetSessionData.email,
        emailConfirmed: true,
      },
    }

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const resetPassword = async (req, res) => {
  try {
    if (req.session?.passwordReset.step != "reset password")
      return res.status(400).send({ msg: "Invalid cookie at endpoint" })

    const { email } = req.session.passwordReset.data

    const { newPassword } = req.body

    const resp = await passwordResetService.resetPassword(email, newPassword)

    req.session.destroy()

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
