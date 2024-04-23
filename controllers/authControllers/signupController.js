import { userRegistrationService } from "../../services/authServices.js"
import {
  EmailConfirmationService,
  SignupEmailConfirmationStrategy,
} from "../../services/EmailConfirmationService.js"

export const signupController = async (req, res) => {
  const { stage } = req.params

  const stageHandlers = {
    request_new_account: (req, res) => newAccountRequestController(req, res),
    verify_email: (req, res) => emailVerificationController(req, res),
    register_user: (req, res) => userRegistrationController(req, res),
  }
  stageHandlers[stage](req, res)
}

const newAccountRequestController = async (req, res) => {
  try {
    const response = await new EmailConfirmationService(
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

const emailVerificationController = async (req, res) => {
  try {
    const response = await new EmailConfirmationService(
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

const userRegistrationController = async (req, res) => {
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
