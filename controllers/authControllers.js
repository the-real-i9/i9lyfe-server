import {
  emailVerificationService,
  newAccountRequestService,
  userRegistrationService,
} from "../services/authServices.js"

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const newAccountRequestController = async (req, res) => {
  try {
    const { email } = req.body

    const response = await newAccountRequestService(email)
    if (!response.ok) {
      return res.status(response.err.code).send({ reason: response.err.reason })
    }

    req.session.potential_user_verification_data = response.data.verfData

    res.status(200).send({
      msg: `Enter the 6-digit code sent to ${email} to verify your email`,
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
export const emailVerificationController = async (req, res) => {
  try {
    const { code: userInputCode } = req.body

    const response = emailVerificationService(
      req.session.potential_user_verification_data,
      userInputCode
    )

    if (!response.ok) {
      return res.status(response.err.code).json({ reason: response.err.reason })
    }

    req.session.potential_user_verification_data = response.data.updatedVerfdata

    res.status(200).send({
      msg: `Your email ${req.session.potential_user_verification_data.email} has been verified!`,
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
export const signupController = async (req, res) => {
  try {
    const { email } = req.session.potential_user_verification_data
    const response = await userRegistrationService({ email, ...req.body })

    if (!response.ok) {
      return res.sendStatus(500)
    }

    req.session.destroy()

    res
      .status(201)
      .send({
        msg: "Registration success! You're automatically logged in.",
        jwtToken: response.data.jwtToken,
      })
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}

export const signinController = async (req, res) => {
  try {
  } catch (error) {}
}
