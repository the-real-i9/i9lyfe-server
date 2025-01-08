import { checkExact, checkSchema } from "express-validator"
import { errHandler, limitOffsetSchema } from "./miscs.js"

export const searchAndFilter = [
  checkExact(
    checkSchema(
      {
        term: {
          optional: true,
          notEmpty: {
            errorMessage: "what do you want to search?"
          },
        },
        filter: {
          optional: true,
          isIn: {
            options: [["user", "photo", "video", "reel", "story", "hashtag"]],
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

export const getHashtagPosts = [
  checkExact(
    checkSchema(
      {
        filter: {
          optional: true,
          isIn: {
            options: [["photo", "video", "reel", "story"]],
            errorMessage: "invalid post type filter",
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
