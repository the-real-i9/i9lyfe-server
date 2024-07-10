import { checkExact, checkSchema, param } from "express-validator"
import { errHandler } from "./errorHandler.js"

export const validateIdParams = [
  param("**").isInt().withMessage("non-integer value"),
  errHandler,
]

export const editProfile = [
  checkExact(
    checkSchema(
      {
        name: {
          optional: true,
          notEmpty: {
            errorMessage: "name value cannot be empty",
          },
        },
        birthday: {
          optional: true,
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

export const updateConnectionStatus = [
  checkExact(
    checkSchema(
      {
        connection_status: {
          isIn: {
            options: ["online", "offline"],
            errorMessage: "value must be either 'online' or 'offline'",
          },
        },
        last_active: {
          custom: {
            options: (value, { req }) =>
              req.body.connection_status === "offline",
            errorMessage: "should only be set if connection status is offline",
            bail: true,
          },
          isDate: {
            errorMessage: "invalid date string format",
          },
        },
      },
      ["body"]
    ),
    { message: "request body contains invalid fields" }
  ),
  errHandler,
]
