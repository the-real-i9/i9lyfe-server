import jwt from "jsonwebtoken"
import bcrypt from "bcrypt"
import { randomInt } from "node:crypto"

/**
 * @returns {string}
 */
export const generateJwt = (payload) =>
  jwt.sign(payload, process.env.JWT_SECRET)

export const renewJwtToken = (socket) => {
  const { client_username } = socket.jwt_payload

  const newJwtToken = generateJwt({ client_username })

  socket.emit("renewed jwt", newJwtToken)
}

export const generateTokenWithExpiration = () => {
  let token = randomInt(100000, 999999)

  if (process.env.GO_ENV !== "production") {
    token = Number(process.env.DUMMY_VERF_TOKEN)
  }

  const expires = new Date(Date.now() + 60 * 60 * 1000)

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
