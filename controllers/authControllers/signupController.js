import { userRegistrationService } from "../../services/authServices.js"
import {
  EmailConfirmationService,
  SignupEmailConfirmationStrategy,
} from "../../services/emailConfirmationService.js"

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
    const response = await new EmailConfirmationService(
      new SignupEmailConfirmationStrategy()
    ).handleEmailSubmission(req)

    if (!response.ok)
      return res.status(response.error.code).send({ msg: response.error.msg })

    res.status(response.success.code).send({ msg: response.success.msg })
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
    const response = await new EmailConfirmationService(
      new SignupEmailConfirmationStrategy()
    ).handleTokenValidation(req)

    if (!response.ok) {
      return res.status(response.error.code).send({ msg: response.error.msg })
    }

    res.status(response.success.code).send({ msg: response.success.msg })
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
    const { email } = req.session.email_verification_data
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
