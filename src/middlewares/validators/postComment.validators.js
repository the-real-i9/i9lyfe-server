import { checkExact, checkSchema, param } from "express-validator"
import { errHandler } from "./miscs.js"

export const validateIdParams = [
  param("*").isInt().withMessage("expected integer value"),
  errHandler,
]

export const createNewPost = [
  checkExact(
    checkSchema(
      {
        media_blobs: {
          isArray: {
            options: { min: 1 },
            errorMessage: "value must me an array of at least one item",
          },
        },
        "media_blobs.*": {
          isArray: {
            options: { min: 1, max: 10 * 1024 ** 2 },
            errorMessage:
              "item must me an array of uint8 integers with a maximum of 10mb",
          },
        },
        type: {
          notEmpty: true,
          isIn: {
            options: ["photo", "video", "reel", "story"],
            errorMessage: "invalid post type",
          },
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
        attachment_blob: {
          optional: true,
          isArray: {
            options: { min: 1, max: 10 * 1024 ** 2 },
            errorMessage:
              "value must me an array of uint8 integers with a maximum of 10mb",
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
          custom: {
            options: (value) => value.codePointAt() >= 0x1f600 && value.codePointAt() <= 0x1faff,
            errorMessage: "invalid emoji"
          }
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
