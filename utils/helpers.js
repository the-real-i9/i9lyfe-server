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
export const generateJwtToken = (payload) =>
  jwt.sign(payload, process.env.JWT_SECRET)

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
  return matches ? [...new Set(matches)] : []
}

/** @param {string} text */
export const extractHashtags = (text) => {
  const matches = text.match(/(?<=#)\w+/g)
  return matches ? [...new Set(matches)] : []
}

/**
 * @param {object} param0
 * @param {number} param0.rowsCount
 * @param {number} param0.columnsCount
 * @param {number} param0.paramNumFrom
 */
export const generateMultiRowInsertValuesParameters = ({
  rowsCount,
  columnsCount,
  paramNumFrom = 1,
}) => {
  // Case: (rowsCount: 3, columnsCount: 2)
  let paramString = ""

  for (let ri = 0; ri < rowsCount; ri++) {
    paramString += "("
    for (let ci = 0; ci < columnsCount; ci++) {
      // 1st row
        // (0 * 2) + 0 + 1 = 1
        // (0 * 2) + 1 + 1 = 2
      // 2nd row
        // (1 * 2) + 0 + 1 = 3
        // (1 * 2) + 1 + 1 = 4
      const n = (ri * columnsCount) + ci + paramNumFrom
      paramString += "$" + n + ", "
    }
    paramString = paramString.slice(0, -2)
    paramString += "), "
  }
  paramString = paramString.slice(0, -2)


  return paramString
}

/**
 * @param {string[]} cols
 * @param {number} paramNumFrom
 */
export const generateMultiColumnUpdateSetParameters = (
  cols,
  paramNumFrom = 1
) => cols.map((col, i) => `${col} = $${paramNumFrom + i}`).join(", ")

/**
 * @param {object} param0
 * @param {string} param0.columnName The `jsonb` type column
 * @param {string[]} param0.jsonbKeys
 * @param {number} param0.paramNumFrom Starting parameter number. If `1` then we'll start from `$1` and increment futher
 */
export const generateJsonbMultiKeysSetParameters = ({
  columnName,
  jsonbKeys,
  paramNumFrom = 1,
}) => {
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
    Object.entries(obj).reduce((acc, [k, v]) => {
      if (v !== null)
        acc.push([
          k,
          Object.prototype.toString.call(v) === "[object Object]"
            ? removeNullFields(v)
            : v,
        ])
      return acc
    }, [])
  )
}

export const stripNulls = (object) => {
  return Array.isArray(object)
    ? object.map(removeNullFields)
    : removeNullFields(object)
}
