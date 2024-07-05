import { userSigninService } from "../../services/auth/auth.service.js"

export const signinController = async (req, res) => {
  try {
    const { emailOrUsername, password } = req.body

    const response = await userSigninService(emailOrUsername, password)

    if (!response.ok) {
      return res.status(response.error.code).send({ msg: response.error.msg })
    }

    res.status(200).send({
      msg: "Signin success!",
      userData: response.data.user,
      jwtToken: response.data.jwtToken,
    })
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}
