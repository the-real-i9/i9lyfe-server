import { userRegistrationService } from "../../services/authServices.js"
import {
  EmailConfirmationService,
  SignupEmailConfirmationStrategy,
} from "../../services/EmailConfirmationService.js"

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const signupController = async (req, res) => {
  const { stage } = req.query

  const stageHandlers = {
    new_account_request: (req, res) => newAccountRequestHandler(req, res),
    email_verification: (req, res) => emailVerificationHandler(req, res),
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
      return res.status(response.err.code).send({ reason: response.err.reason })

    res.status(200).send({ msg: response.successMessage })
  } catch (error) {
    console.error(error)
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
      return res.status(response.err.code).send({ reason: response.err.reason })
    }

    res.status(200).send({ msg: response.successMessage })
  } catch (error) {
    console.error(error)
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
      return res
        .status(response.error.code)
        .send({ reason: response.error.reason })
    }

    req.session.destroy()

    res.status(201).send({
      msg: "Registration success! You're automatically logged in.",
      userData: response.data.userData,
      jwtToken: response.data.jwtToken,
    })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
