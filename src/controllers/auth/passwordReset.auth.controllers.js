import * as passwordResetServices from "../../services/auth/passwordReset.auth.services.js"

export const requestPasswordReset = async (req, res) => {
  const { email } = req.body
  try {
    const response = await passwordResetServices.requestPasswordReset(email)

    if (!response.ok)
      return res.status(response.error.code).send({ msg: response.error.msg })

    req.session.password_reset_email_confirmation_state =
      response.data.sessionData

    res.status(200).send({ msg: response.data.msg })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const confirmEmail = async (req, res) => {
  const { code } = req.body

  try {
    const response = await passwordResetServices.confirmEmail(
      code,
      req.session.password_reset_email_confirmation_state
    )

    if (!response.ok) {
      return res.status(response.error.code).send({ msg: response.error.msg })
    }

    req.session.password_reset_email_confirmation_state =
      response.data.sessionData

    res.status(200).send({ msg: response.data.msg })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const resetPassword = async (req, res) => {
  try {
    const { email } = req.session.password_reset_email_confirmation_state
    const { newPassword } = req.body

    const response = await passwordResetServices.resetPassword(email, newPassword)

    req.session.destroy()

    res.status(200).send({ msg: response.data.msg })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
