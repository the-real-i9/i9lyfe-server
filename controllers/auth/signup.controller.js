import { emailConfirmationService, userRegistrationService } from "../../services/auth/auth.service.js"
import {
  SignupEmailConfirmationStrategy,
} from "../../services/auth/emailConfirmationStrategy.auth.service.js"

export const signupController = async (req, res) => {
  const { step } = req.params

  const stepHandlers = {
    request_new_account: (req, res) => newAccountRequestHandler(req, res),
    verify_email: (req, res) => emailVerificationHandler(req, res),
    register_user: (req, res) => userRegistrationHandler(req, res),
  }
  stepHandlers[step](req, res)
}

const newAccountRequestHandler = async (req, res) => {
  try {
    const response = await emailConfirmationService(
      new SignupEmailConfirmationStrategy()
    ).handleEmailSubmission(req)

    if (!response.ok)
      return res.status(response.err.code).send({ reason: response.err.reason })

    res.status(200).send({ msg: response.successMessage })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

const emailVerificationHandler = async (req, res) => {
  try {
    const response = await emailConfirmationService(
      new SignupEmailConfirmationStrategy()
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
    // console.error(error)
    res.sendStatus(500)
  }
}
