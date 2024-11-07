import bcrypt from "bcrypt"
import { User } from "../../models/user.model"
import { generateJwt } from "../../utils/helpers"

/**
 * @param {string} emailOrUsername
 * @param {string} passwordInput
 */
export const signin = async (emailOrUsername, passwordInput) => {
  const userData = await User.findOneForAuth(emailOrUsername)
  
  if (!userData) {
    return {
      ok: false,
      error: { code: 422, msg: "Incorrect email or password" },
      data: null,
    }
  }

  const { pswd: storedPswd, ...user } = userData

  if (!(await bcrypt.compare(passwordInput, storedPswd))) {
    return {
      ok: false,
      error: { code: 422, msg: "Incorrect email or password" },
      data: null,
    }
  }

  const jwt = generateJwt({
    client_user_id: user.id,
    client_username: user.username,
  })

  return {
    ok: true,
    error: null,
    data: {
      msg: "Signin success!",
      user,
      jwt,
    },
  }
}