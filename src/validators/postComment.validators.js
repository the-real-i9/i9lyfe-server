import { checkExact, checkSchema } from "express-validator"
import { errHandler } from "./miscs.js"

export const createNewPost = [
  checkExact(
    checkSchema(
      {
        media_data_list: {
          isArray: {
            options: { min: 1 },
            errorMessage: "value must be an array of at least one length",
          },
        },
        "media_data_list.*": {
          isArray: {
            options: { min: 1, max: 8 * 1024 ** 2 },
            errorMessage:
              "item must be an array of uint8 integers containing a maximum of 8mb",
          },
        },
        type: {
          notEmpty: true,
          isIn: {
            options: [["photo", "video", "reel", "story"]],
            errorMessage: "invalid post type",
          },
          custom: {
            if: (value) => value != "photo",
            options: (value, { req }) => req.body.media_data_list.length === 1,
            errorMessage: (value) => `you can't post more than one ${value} at a time`
          }
        },
        description: {
          optional: true,
          notEmpty: true,
        },
      },
      ["body"]
    ),
    { message: "request body contains invalid fields" }
  ),
  errHandler,
]

export const commentOn = [
  checkExact(
    checkSchema(
      {
        comment_text: {
          notEmpty: true,
        },
        attachment_data: {
          optional: true,
          isArray: {
            options: { min: 1, max: 8 * 1024 ** 2 },
            errorMessage:
              "value must me an array of uint8 integers with a maximum of 8mb",
          },
        },
      },
      ["body"]
    ),
    { message: "request body contains invalid fields" }
  ),
  errHandler,
]

export const reactTo = [
  checkExact(
    checkSchema(
      {
        reaction: {
          notEmpty: true,
          isLength: {
            options: { min: 1, max: 1 },
            errorMessage: "one reaction required",
          },
          isSurrogatePair: {
            errorMessage: "invalid reaction",
          },
        },
      },
      ["body"]
    ),
    {
      message: "request body contains invalid fields",
    }
  ),
  errHandler,
]
