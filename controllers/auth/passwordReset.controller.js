import { passwordResetService } from "../../services/auth/auth.service.js"
import {
  EmailConfirmationService,
  PasswordResetEmailConfirmationStrategy,
} from "../../services/auth/emailConfirmationStrategy.auth.service.js"

/**
 * @typedef {import("express").Request} ExpressRequest
 * @typedef {import("express").Response} ExpressResponse
 */

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
export const passwordResetController = async (req, res) => {
  const { step } = req.params

  const stepHandlers = {
    request_password_reset: (req, res) => passwordResetRequestHandler(req, res),
    confirm_email: (req, res) =>
      passwordResetEmailConfirmationHandler(req, res),
    reset_password: (req, res) => passwordResetHandler(req, res),
  }
  stepHandlers[step](req, res)
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */

const passwordResetRequestHandler = async (req, res) => {
  try {
    const response = await new EmailConfirmationService(
      new PasswordResetEmailConfirmationStrategy()
    ).handleEmailSubmission(req)

    if (!response.ok)
      return res.status(response.err.code).send({ reason: response.err.reason })

    res.status(200).send({ msg: response.successMessage })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
const passwordResetEmailConfirmationHandler = async (req, res) => {
  try {
    const response = await new EmailConfirmationService(
      new PasswordResetEmailConfirmationStrategy()
    ).handleTokenValidation(req)

    if (!response.ok) {
      return res.status(response.err.code).send({ reason: response.err.reason })
    }

    res.status(200).send({ msg: response.successMessage })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 */
const passwordResetHandler = async (req, res) => {
  try {
    const { email: userEmail } =
      req.session.password_reset_email_confirmation_data
    const { newPassword } = req.body
    const response = await passwordResetService(userEmail, newPassword)
    if (!response.ok) {
      res.status(response.err.code).send({ reason: response.err.reason })
    }

    res.status(200).send({ msg: "Your password has been changed successfully" })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}
