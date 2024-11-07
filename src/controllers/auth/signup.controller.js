import {
  userRegistrationService,
} from "../../services/auth/auth.service.js"
import * as signupServices from "../../services/auth/signup.auth.services.js"



export const requestNewAccount = async (req, res) => {
  const { email } = req.body

  try {
    const response = await signupServices.requestNewAccount(email)

    if (!response.ok)
      return res.status(response.error.code).send({ msg: response.error.msg })

    req.session.email_verification_state = response.data.sessionData

    res.status(200).send({ msg: response.data.msg })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const verifyEmail = async (req, res) => {
  const {code} = req.body

  try {
    const response = await signupServices.verifyEmail(code, req.session.email_verification_state)

    if (!response.ok) {
      return res.status(response.error.code).send({ msg: response.error.msg })
    }

    req.session.email_verification_state = response.data.sessionData

    res.status(200).send({ msg: response.data.msg })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const registerUser = async (req, res) => {
  try {
    const { email } = req.session.email_verification_state

    const response = await signupServices.registerUser({ email, ...req.body })

    if (!response.ok) {
      return res
        .status(response.error.code)
        .send({ msg: response.error.msg })
    }

    req.session.destroy()

    res.status(201).send(response.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
