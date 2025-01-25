import * as securityServices from "../security.services.js"
import { User } from "../../models/user.model.js"

export const signin = async (email_or_username, inputPassword) => {
  const userData = await User.findOne(email_or_username)

  if (!userData)
    return {
      error: { msg: "Incorrect email or password" },
    }

  const { password: storedPassword, ...user } = userData

  if (!(await securityServices.passwordsMatch(inputPassword, storedPassword)))
    return {
      error: { msg: "Incorrect email or password" },
    }

  const jwt = securityServices.generateJwt({
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
