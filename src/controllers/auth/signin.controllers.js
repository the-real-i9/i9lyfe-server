import { User } from "../../models/user.model.js"
import * as authServices from "../../services/auth.services.js"

const signinController = async (req, res) => {
  try {
    const { email_or_username, password: inputPassword } = req.body

    const userData = await User.findOneIncPassword(email_or_username)

    if (!userData) {
      return res.status(422).send({ msg: "Incorrect email or password" })
    }

    const { pswd: storedPassword, ...user } = userData

    if (!(await authServices.passwordsMatch(inputPassword, storedPassword))) {
      return res.status(422).send({ msg: "Incorrect email or password" })
    }

    const jwt = authServices.generateJwt({
      client_user_id: user.id,
      client_username: user.username,
    })

    res.status(200).send({
      msg: "Signin success!",
      user,
      jwt,
    })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export default signinController
