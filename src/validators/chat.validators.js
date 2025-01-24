import { checkExact, checkSchema } from "express-validator"
import { errHandler } from "./miscs.js"

export const sendMessage = [
  checkExact(
    checkSchema(
      {
        msg_content: {
          isObject: {
            bail: true,
            errorMessage: "invalid props value",
          },
        },
        "msg_content.type": {
          isIn: {
            options: [["text", "voice", "image", "audio", "video", "file"]],
            errorMessage: "invalid message type",
          },
        },
        "msg_content.props": {
          isObject: {
            bail: true,
            errorMessage: "invalid props value",
          },
        },
        "msg_content.props.textContent": {
          custom: {
            options: (value, { req }) =>
              req.body.msg_content.type === "text" || !value,
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          notEmpty: {
            if: (value, { req }) => req.body.msg_content.type === "text",
            errorMessage: "cannot be empty",
          },
        },
        "msg_content.props.media_data": {
          custom: {
            options: (value, { req }) =>
              req.body.msg_content.type !== "text" || !value,
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          isArray: {
            if: (value, { req }) => req.body.msg_content.type !== "text",
            options: { min: 1, max: 8 * 1024 ** 2 },
            errorMessage:
              "value must me an array of uint8 integers with a maximum of 8mb",
          },
        },
        "msg_content.props.duration": {
          custom: {
            options: (value, { req }) =>
              req.body.msg_content.type === "voice" || !value,
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          isInt: {
            if: (value, { req }) => req.body.msg_content.type === "voice",
            options: { min: 1 },
            errorMessage: "invalid duration: less than 1",
          },
        },
        "msg_content.props.extension": {
          custom: {
            options: (value, { req }) =>
              req.body.msg_content.type === "file" || !value,
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
        },
        "msg_content.props.mimeType": {
          custom: {
            options: (value, { req }) =>
              !["text", "voice"].includes(req.body.msg_content.type) || !value,
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          isMimeType: {
            if: (value, { req }) =>
              !["text", "voice"].includes(req.body.msg_content.type),
            errorMessage: "invalid mime type",
          },
        },
        "msg_content.props.caption": {
          optional: true,
          custom: {
            options: (value, { req }) =>
              !["text", "voice"].includes(req.body.msg_content.type) || !value,
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          notEmpty: {
            if: (value, { req }) =>
              !["text", "voice"].includes(req.body.msg_content.type),
          },
        },
        "msg_content.props.size": {
          custom: {
            options: (value, { req }) =>
              !["text", "voice"].includes(req.body.msg_content.type) || !value,
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          isInt: {
            if: (value, { req }) =>
              !["text", "voice"].includes(req.body.msg_content.type),
            options: { min: 1, max: 8 * 1024 ** 2 /* 8mb */ },
            errorMessage: "size out of range",
          },
        },
        at: {
          custom: {
            options: (value) => !isNaN(Date.parse(value)),
            errorMessage: "invalid date specified"
          }
        }
      },
      ["body"]
    ),
    { message: "request body contains invalid fields" }
  ),
  errHandler,
]

export const ackMessageDelivered = [
  checkExact(
    checkSchema(
      {
        delivered_at: {
          notEmpty: true,
          custom: {
            options: (value) => !isNaN(Date.parse(value)),
            errorMessage: "invalid date",
          },
        },
      },
      ["body"]
    ),
    { message: "request body contains invalid fields" }
  ),
  errHandler,
]

export const ackMessageRead = [
  checkExact(
    checkSchema(
      {
        read_at: {
          notEmpty: true,
          custom: {
            options: (value) => !isNaN(Date.parse(value)),
            errorMessage: "invalid date",
          },
        },
      },
      ["body"]
    ),
    { message: "request body contains invalid fields" }
  ),
  errHandler,
]

export const reactToMessage = [
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
    { message: "request body contains invalid fields" }
  ),
  errHandler,
]

export const deleteMessage = [
  checkExact(
    checkSchema(
      {
        delete_for: {
          isIn: {
            options: [["me", "everyone"]],
            errorMessage:
              "invalid delete-for value, should be either 'me' or 'everyone'",
          },
        },
      },
      ["query"]
    ),
    { message: "request body contains invalid fields" }
  ),
  errHandler,
]
