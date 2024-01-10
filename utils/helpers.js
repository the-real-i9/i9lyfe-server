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
  return jwt.sign(payload, process.env.JWT_SECRET, { expiresIn: "5h" })
}

/**
 * @param {number} compareToken
 * @param {number} inputToken
 */
export const tokensMatch = (compareToken, inputToken) =>
  compareToken === inputToken

/** @param {Date} tokenExpiration */
export const tokenLives = (tokenExpiration) =>
  Date.now() < new Date(tokenExpiration)

/** @param {string} text */
export const extractMentions = (text) => {
  const matches = text.match(/(?<=@)\w+/g)
  return matches && [...new Set(matches)]
}

/** @param {string} text */
export const extractHashtags = (text) => {
  const matches = text.match(/(?<=#)\w+/g)
  return matches && [...new Set(matches)]
}

/**
 * @param {number} rowsCount
 * @param {number} columnsCount
 */
export const generateMultiRowInsertValuesParameters = (
  rowsCount,
  columnsCount
) =>
  Array(rowsCount)
    .fill()
    .map(
      (r, ri) =>
        `(${Array(columnsCount)
          .fill()
          .map((f, fi) => `$${ri * columnsCount + (fi + 1)}`)
          .join(", ")})`
    )
    .join(", ")

export const generateMultiColumnUpdateSetParameters = (keys) =>
  keys.map((key, i) => `${key} = $${i + 1}`).join(", ")
