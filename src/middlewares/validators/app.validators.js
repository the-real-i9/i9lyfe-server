import { checkExact, checkSchema } from "express-validator"
import { errHandler, limitOffsetSchema } from "./miscs.js"

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

export const searchUsersToChat = [
  checkExact(
    checkSchema(
      {
        search: {
          matches: {
            options: /^[\w-]{3,}$/,
            errorMessage: "invalid username format",
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

export const searchAndFilter = [
  checkExact(
    checkSchema(
      {
        search: {
          optional: true,
          notEmpty: {
            errorMessage: "what do you wanna search"
          },
        },
        filter: {
          optional: true,
          isIn: {
            options: ["user", "photo", "video", "reel", "story", "hashtag"],
            errorMessage: "invalid filter value",
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
