import { passwordResetService } from "../../services/authServices.js"
import {
  EmailConfirmationService,
  PasswordResetEmailConfirmationStrategy,
} from "../../services/emailConfirmationService.js"

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const passwordResetController = async (req, res) => {
  const { stage } = req.query

  const stageHandlers = {
    password_reset_request: (req, res) => passwordResetRequestHandler(req, res),
    email_confirmation: (req, res) =>
      passwordResetEmailConfirmationHandler(req, res),
    password_reset: (req, res) => passwordResetHandler(req, res),
  }
  stageHandlers[stage](req, res)
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
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
    console.log(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
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
    console.log(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
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
    console.log(error)
    res.sendStatus(500)
  }
}
