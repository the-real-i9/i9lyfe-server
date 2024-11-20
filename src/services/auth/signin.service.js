import * as securityServices from "../security.services.js"
import { User } from "../../models/user.model.js"

export const signin = async (email_or_username, inputPassword) => {
  const userData = await User.findOneIncPassword(email_or_username)

  if (!userData)
    return {
      error: { msg: "Incorrect email or password" },
    }

  const { pswd: storedPassword, ...user } = userData

  if (!(await securityServices.passwordsMatch(inputPassword, storedPassword)))
    return {
      error: { msg: "Incorrect email or password" },
    }

  const jwt = securityServices.generateJwt({
    client_user_id: user.id,
    client_username: user.username,
  })

  return {
    data: {
      msg: "Signin success!",
      user,
      jwt,
    },
  }
}
