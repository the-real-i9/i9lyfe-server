import { userSigninService } from "../../services/auth/auth.service.js"

export const signinController = async (req, res) => {
  try {
    const { email, password } = req.body

    const response = await userSigninService(email, password)

    if (!response.ok) {
      return res.status(response.err.code).send({ reason: response.err.reason })
    }

    res.status(200).send({
      msg: "Signin success!",
      userData: response.data.userData,
      jwtToken: response.data.jwtToken,
    })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}
