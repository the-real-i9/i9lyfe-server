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

/** @param {string[]} cols */
export const generateMultiColumnUpdateSetParameters = (cols) =>
  cols.map((col, i) => `${col} = $${1 + i}`).join(", ")

/**
 * @param {string} columnName The `jsonb` type column
 * @param {string[]} jsonbKeys
 * @param {number} paramNumFrom Starting parameter number. If `1` then we'll start from `$1` and increment futher
 */
export const generateJsonbMultiKeysSetParameters = (
  columnName,
  jsonbKeys,
  paramNumFrom
) => {
  // goal: [columnName] = jsonb_set([columnName], '{key}', '"$[paramNumFrom]"', '{key2}', '"$[paramNumFrom + 1]"')
  return `${columnName} = ${jsonbKeys
    .map(
      (key, i) =>
        `jsonb_set(${columnName}, '{${key}}', '"$${paramNumFrom + i}"')`
    )
    .join(", ")}`
}

/** @param {object[] | object} obj */
const removeNullFields = (obj) => {
  return Object.fromEntries(
    Object.entries(obj).filter(([, v]) =>
        v !== null
    )
  )
}

export const stripNulls = (object) => {
  if (Array.isArray(object)) return object.map(removeNullFields)
  else removeNullFields(object)
}
