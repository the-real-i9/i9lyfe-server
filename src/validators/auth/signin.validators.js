import { checkExact, checkSchema } from "express-validator"
import { errHandler } from "./miscs.js"

export const signin = [
  checkExact(
    checkSchema(
      {
        email_or_username: {
          notEmpty: { errorMessage: "email or username is required" },
          isEmail: {
            if: (value) => value.includes("@"),
            errorMessage: "invalid email",
          },
          matches: {
            if: (value) => !value.includes("@"),
            options: /^[a-zA-Z0-9][\w-]+[a-zA-Z0-9]$/,
            errorMessage: "invalid username pattern",
          },
        },
        password: {
          notEmpty: {
            errorMessage: "password is required",
          },
        },
      },
      ["body"]
    ),
    { message: "request body contains invalid fields" }
  ),
  errHandler,
]
