import { userSigninService } from "../../services/authServices.js"

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const signinController = async (req, res) => {
  try {
    const { email, password } = req.body

    const response = await userSigninService(email, password)

    if (!response.ok) {
      return res.status(response.err.code).send({ reason: response.err.reason })
    }

    res
      .status(200)
      .send({
        userData: response.data.userData,
        jwtToken: response.data.jwtToken,
      })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
