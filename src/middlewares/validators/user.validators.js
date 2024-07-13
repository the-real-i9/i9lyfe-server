import { checkExact, checkSchema } from "express-validator"
import { errHandler, limitOffsetSchema } from "./miscs.js"

export const editProfile = [
  checkExact(
    checkSchema(
      {
        "": {
          custom: {
            options: (value) => !!Object.keys(value).length,
            errorMessage: "must contain at least one field to update",
          },
        },
        name: {
          optional: true,
          notEmpty: true,
        },
        birthday: {
          optional: true,
          isDate: {
            errorMessage:
              "invalid date format (expects: YYYY/MM/DD or YYYY-MM-DD)",
            bail: true,
          },
        },
        bio: {
          optional: true,
          notEmpty: true,
          isLength: {
            options: { max: 150 },
            errorMessage: "too many characters (max is 150)",
          },
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
            options: [["online", "offline"]],
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
            errorMessage:
              "invalid date format (expects: YYYY/MM/DD or YYYY-MM-DD)",
          },
        },
      },
      ["body"]
    ),
    { message: "request body contains invalid fields" }
  ),
  errHandler,
]

export const getNotifications = [
  checkExact(
    checkSchema(
      {
        from: {
          notEmpty: true,
          isDate: {
            errorMessage:
              "invalid date format (expects: YYYY/MM/DD or YYYY-MM-DD)",
            bail: true,
          },
          isBefore: {
            options: new Date(),
            errorMessage: "invalid time period",
          },
        },
        ...limitOffsetSchema,
      },
      ["query"]
    ),
    { message: "request query parameters contains invalid fields" }
  ),
  errHandler,
]
