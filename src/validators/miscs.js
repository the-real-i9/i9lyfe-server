import { checkExact, checkSchema, validationResult } from "express-validator"

export const errHandler = (req, res, next) => {
  try {
    const result = validationResult(req)
    if (!result.isEmpty()) {
      return res.status(400).send({
        validation_error: result
          .formatWith((err) => `${err.msg}`)
          .mapped(),
      })
    }

    return next()
  } catch (error) {
    console.error(error)
    return res.sendStatus(500)
  }
}

export const validateParams = [
  checkSchema(
    {
      "*": {
        isUUID: {
          if: (value, { path }) => path.endsWith("_id"),
          errorMessage: "expected a UUID string",
        },
        isLength: {
          if: (value, { path }) => path.endsWith("_username"),
          options: { min: 3 },
          errorMessage: "suspected incorrect username: too short",
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
