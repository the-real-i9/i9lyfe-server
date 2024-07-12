import { validationResult } from "express-validator"

export const errHandler = (req, res, next) => {
  const result = validationResult(req)
  if (!result.isEmpty()) {
    return res.status(422).send({ error: result.mapped() })
  }

  return next()
}

export const limitOffsetSchema = {
  limit: {
    isInt: {
      options: { min: 1 },
      errorMessage: "must be integer greater than zero",
    },
  },
  offset: {
    isInt: {
      options: { min: 0 },
      errorMessage: "must be integer greater than -1",
    },
  },
}
