import { checkExact, checkSchema } from "express-validator"
import { errHandler } from "./miscs.js"

export const requestNewAccount = [
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

export const verifyEmail = [
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

export const registerUser = [
  checkExact(
    checkSchema(
      {
        username: {
          isLength: {
            options: { min: 3 },
            errorMessage: "username too short",
          },
          matches: {
            options: /^[a-zA-Z0-9][\w-]+[a-zA-Z0-9]$/,
            errorMessage: "invalid username pattern",
          },
        },
        password: {
          isLength: {
            options: { min: 8 },
            errorMessage: "password too short",
          },
        },
        name: {
          notEmpty: {
            errorMessage: "name value cannot be empty",
          },
        },
        birthday: {
          notEmpty: true,
          isDate: {
            errorMessage:
              "invalid date format (expects: YYYY/MM/DD or YYYY-MM-DD)",
          },
        },
        bio: {
          optional: true,
          isLength: { options: { max: 150 } },
          errorMessage: "too many characters (max is 150)",
        },
      },
      ["body"]
    ),
    { message: "request body contains invalid fields" }
  ),
  errHandler,
]
