import { userSigninService } from "../../services/auth/auth.service.js"

const signinController = async (req, res) => {
  try {
    const { emailOrUsername, password } = req.body

    const response = await userSigninService(emailOrUsername, password)

    if (!response.ok) {
      return res.status(response.error.code).send({ msg: response.error.msg })
    }

    res.status(200).send(response.data)
  } catch (error) {
    // console.error(error)
    res.sendStatus(500)
  }
}

export default signinController