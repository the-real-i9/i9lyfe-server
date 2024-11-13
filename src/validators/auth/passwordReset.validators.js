import { checkExact, checkSchema } from "express-validator"
import { errHandler } from "./miscs.js"

export const requestPasswordReset = [
  checkExact(
    checkSchema(
      {
        email: { isEmail: { errorMessage: "invalid email" } },
      },
      ["body"]
    ),
    { message: "request body contains invalid fields" }
  ),
  errHandler,
]

export const confirmEmail = [
  checkExact(
    checkSchema(
      {
        code: {
          isNumeric: {
            options: { no_symbols: true },
            errorMessage: "invalid non-numeric code value",
          },
          isLength: {
            options: { min: 6, max: 6 },
            errorMessage: "code must be 6 digits",
          },
        },
      },
      ["body"]
    ),
    { message: "request body contains invalid fields" }
  ),
  errHandler,
]

export const resetPassword = [
  checkExact(
    checkSchema(
      {
        new_password: {
          isLength: {
            options: { min: 8 },
            errorMessage: "password too short",
          },
        },
        confirm_new_password: {
          custom: {
            options: (value, { req }) => value === req.body.new_password,
            errorMessage: "password mismatch",
          },
        },
      },
      ["body"]
    ),
    { message: "request body contains invalid fields" }
  ),
  errHandler,
]
