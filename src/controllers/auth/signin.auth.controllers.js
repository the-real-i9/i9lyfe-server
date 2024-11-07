import * as signinServices from "../../services/auth/signin.auth.services.js"

const signinController = async (req, res) => {
  try {
    const { email_or_username, password } = req.body

    const response = await signinServices.signin(email_or_username, password)

    if (!response.ok) {
      return res.status(response.error.code).send({ msg: response.error.msg })
    }

    res.status(200).send(response.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export default signinController