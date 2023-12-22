import { userExists } from "../../models/userModel.js"
import { userRegistrationService } from "../../services/authServices.js"
import { emailConfirmationService } from "../../services/emailConfirmationService.js"
import {
  EmailVerificationSuccessMailSender,
  EmailVerificationTokenMailSender,
} from "../../services/mailingService.js"

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const signupController = async (req, res) => {
  const { stage } = req.query

  const stageHandlers = {
    email_submission: (req, res) => newAccountRequestHandler(req, res),
    token_validation: (req, res) => emailVerificationHandler(req, res),
    user_registration: (req, res) => userRegistrationHandler(req, res),
  }
  stageHandlers[stage](req, res)
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
const newAccountRequestHandler = async (req, res) => {
  try {
    const { email } = req.body

    if (await userExists(email))
      return {
        ok: false,
        err: {
          code: 422,
          reason: "A user with this email already exists",
        },
        data: null,
      }

    const response = await emailConfirmationService(req, {
      tokenMailSender: new EmailVerificationTokenMailSender(),
    })
    if (!response.ok)
      return res.status(response.err.code).send({ reason: response.err.reason })

    res.status(200).send({
      msg: `Enter the 6-digit token sent to ${email} to verify your email`,
    })
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
const emailVerificationHandler = async (req, res) => {
  try {
    const response = await emailConfirmationService(req, {
      primaryMailSender: new EmailVerificationSuccessMailSender(),
    })

    if (!response.ok) {
      return res.status(response.err.code).send({ reason: response.err.reason })
    }

    res.status(200).send({
      msg: `Your email ${req.session.email_confirmation_data.email} has been verified!`,
    })
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
const userRegistrationHandler = async (req, res) => {
  try {
    const { email } = req.session.potential_user_verification_data
    const response = await userRegistrationService({ email, ...req.body })

    if (!response.ok) {
      return res.sendStatus(500)
    }

    req.session.destroy()

    res.status(201).send({
      msg: "Registration success! You're automatically logged in.",
      jwtToken: response.data.jwtToken,
    })
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}
