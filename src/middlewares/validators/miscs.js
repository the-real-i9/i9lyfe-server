import { checkExact, checkSchema, validationResult } from "express-validator"

export const errHandler = (req, res, next) => {
  try {
    const result = validationResult(req)
    if (!result.isEmpty()) {
      return res.status(422).send({ error: result.mapped() })
    }

    return next()
  } catch (error) {
    console.error(error)
    return res.sendStatus(500)
  }
}

export const validateIdParams = [
  checkSchema(
    {
      "*": {
        isInt: {
          if: (value, { path }) => path.endsWith("_id"),
          options: { min: 0 },
          errorMessage: "expected an integer value greater than -1",
        },
        isLength: {
          if: (value, { path }) => path === "reaction",
          options: { min: 2, max: 2 },
          errorMessage: "invalid reaction",
        },
        isSurrogatePair: {
          if: (value, { path }) => path === "reaction",
          errorMessage: "invalid reaction",
        },
      },
    },
    ["params"]
  ),
  errHandler,
]

export const limitOffsetSchema = {
  limit: {
    optional: true,
    isInt: {
      options: { min: 1 },
      errorMessage: "must be integer greater than zero",
    },
  },
  offset: {
    optional: true,
    isInt: {
      options: { min: 0 },
      errorMessage: "must be integer greater than -1",
    },
  },
}

export const validateLimitOffset = [
  checkExact(
    checkSchema(
      {
        ...limitOffsetSchema,
      },
      ["query"]
    ),
    { message: "request query parameters contains invalid fields" }
  ),
  errHandler,
]
