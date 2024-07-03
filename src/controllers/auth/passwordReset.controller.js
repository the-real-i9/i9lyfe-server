import {
  emailConfirmationService,
  passwordResetService,
} from "../../services/auth/auth.service.js"
import { PasswordResetEmailConfirmationStrategy } from "../../services/auth/emailConfirmationStrategy.auth.service.js"

/**
 * @typedef {import("express").Request} ExpressRequest
 * @typedef {import("express").Response} ExpressResponse
 */

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */

export const requestPasswordResetController = async (req, res) => {
  const { email } = req.body
  try {
    const response = await emailConfirmationService(
      new PasswordResetEmailConfirmationStrategy()
    ).handleEmailSubmission(email)

    if (!response.ok)
      return res.status(response.error.code).send({ msg: response.error.msg })

    req.session.password_reset_email_confirmation_state =
      response.data.sessionData

    res.status(200).send({ msg: response.data.msg })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const confirmEmailController = async (req, res) => {
  const { code } = req.body

  try {
    const response = await emailConfirmationService(
      new PasswordResetEmailConfirmationStrategy()
    ).handleCodeValidation(
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
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const resetPasswordController = async (req, res) => {
  try {
    const { email } = req.session.password_reset_email_confirmation_state
    const { newPassword } = req.body

    const response = await passwordResetService(email, newPassword)

    req.session.destroy()

    res.status(200).send({ msg: response.data.msg })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}
