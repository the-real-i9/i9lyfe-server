import { body, checkExact, checkSchema, oneOf } from "express-validator"
import { errHandler } from "./miscs"

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
          matches: {
            options: /^[\w-]{3,}$/,
            errorMessage: "invalid username format",
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
          isDate: {
            errorMessage: "invalid date string format",
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

export const signin = [
  oneOf(
    [
      body("email_or_username").isEmail(),
      body("email_or_username").matches(/^[\w-]{3,}$/),
    ],
    { message: "invalid email or username pattern", errorType: "least_errored" }
  ),
  checkExact(
    checkSchema(
      {
        email_or_username: {
          notEmpty: { errorMessage: "email or username is required" },
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
