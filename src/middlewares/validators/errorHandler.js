import { validationResult } from "express-validator"

export const errHandler = (req, res, next) => {
  const result = validationResult(req)
  if (!result.isEmpty()) {
    return res.status(422).send({ error: result.mapped() })
  }

  return next()
}