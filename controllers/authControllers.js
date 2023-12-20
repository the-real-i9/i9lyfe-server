import {
  emailVerificationService,
  newAccountRequestService,
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

    req.session.potential_user_verfification_data = response.verfData

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
      req.session.potential_user_verfification_data,
      userInputCode
    )

    if (!response.ok) {
      return res.status(response.err.code).json({ reason: response.err.reason })
    }

    req.session.potential_user_verfification_data = response.updatedVerfdata

    res.status(200).send({
      msg: `Your email ${req.session.potential_user_verfification_data.email} has been verified!`,
    })
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}

export const signupController = async (req, res) => {
  try {
  } catch (error) {}
}

export const signinController = async (req, res) => {
  try {
  } catch (error) {}
}
