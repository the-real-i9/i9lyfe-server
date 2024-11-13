import jwt from "jsonwebtoken"
import bcrypt from "bcrypt"
import { randomInt } from "node:crypto"

/**
 * @returns {string}
 */
export const generateJwt = (payload) =>
  jwt.sign(payload, process.env.JWT_SECRET)

export const renewJwtToken = (socket) => {
  const { client_user_id, client_username } = socket.jwt_payload

  const newJwtToken = generateJwt({ client_user_id, client_username })

  socket.emit("renewed jwt", newJwtToken)
}

export const generateTokenWithExpiration = () => {
  const token = randomInt(100000, 999999)
  const expires = new Date(Date.now() + 1 * 60 * 60 * 1000)

  return { token, expires }
}

export const hashPassword = async (password) => {
  return bcrypt.hash(password, 10)
}

export const passwordsMatch = async (inputPassword, storedPassword) => {
  return bcrypt.compare(inputPassword, storedPassword)
}

export const isTokenAlive = (tokenExpiration) =>
  Date.now() < new Date(tokenExpiration)
