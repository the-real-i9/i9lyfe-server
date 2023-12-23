import jwt from "jsonwebtoken"

export const commaSeparateString = (str) => str.replaceAll(" ", ", ")

export const generateCodeWithExpiration = () => {
  const token = Math.trunc(Math.random() * 900000 + 100000)
  const expirationTime = new Date(Date.now() + 1 * 60 * 60 * 1000)

  return [token, expirationTime]
}

/**
 * @param {string|Buffer|JSON} payload
 * @returns {string} A JWT Token
 */
export const generateJwtToken = (payload) => {
  return jwt.sign(payload, process.env.JWT_SECRET, { expiresIn: "1h" })
}
