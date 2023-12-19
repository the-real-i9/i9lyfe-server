// eslint-disable-next-line no-unused-vars
import { Request, Response } from "@types/express"

import { newAccountRequestService } from "../services/authServices"

/**
 * @param {Request} req
 * @param {Response} res
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

export const emailVerificationController = async (req, res) => {
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
}
