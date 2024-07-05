import {
  emailConfirmationService,
  userRegistrationService,
} from "../../services/auth/auth.service.js"
import { SignupEmailConfirmationStrategy } from "../../services/auth/emailConfirmationStrategy.auth.service.js"



export const requestNewAccountController = async (req, res) => {
  const { email } = req.body

  try {
    const response = await emailConfirmationService(
      new SignupEmailConfirmationStrategy()
    ).handleEmailSubmission(email)

    if (!response.ok)
      return res.status(response.error.code).send({ msg: response.error.msg })

    req.session.email_verification_state = response.data.sessionData

    res.status(200).send({ msg: response.data.msg })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const verifyEmailController = async (req, res) => {
  const {code} = req.body

  try {
    const response = await emailConfirmationService(
      new SignupEmailConfirmationStrategy()
    ).handleCodeValidation(code, req.session.email_verification_state)

    if (!response.ok) {
      return res.status(response.error.code).send({ msg: response.error.msg })
    }

    req.session.email_verification_state = response.data.sessionData

    res.status(200).send({ msg: response.data.msg })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export const registerUserController = async (req, res) => {
  try {
    const { email } = req.session.email_verification_state
    const response = await userRegistrationService({ email, ...req.body })

    if (!response.ok) {
      return res
        .status(response.error.code)
        .send({ msg: response.error.msg })
    }

    req.session.destroy()

    res.status(201).send(response.data)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}
