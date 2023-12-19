import { newAccountRequestService } from "../services/authServices.js"

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const newAccountRequestController = async (req, res) => {
  try {
    const { email } = req.body

    const response = await newAccountRequestService(email)

    return res
      .status(response.statusCode)
      .send({ message: response.statusMessage })
  } catch (error) {
    console.log(error)
    res.sendStatus(500)
  }
}

/* export const emailVerificationController = async (req, res) => {
  try {
  } catch (error) {}
}

export const signupController = async (req, res) => {
  try {
  } catch (error) {}
}

export const signinController = async (req, res) => {
  try {
  } catch (error) {}
} */
