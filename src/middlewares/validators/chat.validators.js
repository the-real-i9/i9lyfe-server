import { checkExact, checkSchema } from "express-validator"
import { errHandler } from "./miscs.js"


export const createConversation = [
  checkExact(
    checkSchema(
      {
        partner: {
          isObject: true,
        },
        "partner.user_id": {
          isNumeric: { options: { no_symbols: true } },
        },
        "partner.username": {
          matches: {
            options: /^[a-zA-Z0-9][\w-]+[a-zA-Z0-9]$/,
            errorMessage: "invalid username pattern",
          },
        },
        init_message: {
          isObject: {
            bail: true,
            errorMessage: "invalid props value",
          },
        },
        "init_message.type": {
          isIn: {
            options: [["text", "voice", "image", "audio", "video", "file"]],
            errorMessage: "invalid message type",
          },
        },
        "init_message.props": {
          isObject: {
            bail: true,
            errorMessage: "invalid props value",
          },
        },
        "init_message.props.textContent": {
          custom: {
            options: (value, { req }) =>
              req.body.init_message.type === "text" || !value,
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          notEmpty: {
            if: (value, { req }) => req.body.init_message.type === "text",
            errorMessage: "cannot be empty",
          },
        },
        "init_message.props.media_data": {
          custom: {
            options: (value, { req }) =>
              req.body.init_message.type !== "text" || !value,
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          isArray: {
            if: (value, { req }) => req.body.init_message.type !== "text",
            options: { min: 1, max: 10 * 1024 ** 2 },
            errorMessage:
              "value must me an array of uint8 integers with a maximum of 10mb",
          },
        },
        "init_message.props.duration": {
          custom: {
            options: (value, { req }) =>
              req.body.init_message.type === "voice" || !value,
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          isInt: {
            if: (value, { req }) => req.body.init_message.type === "voice",
            options: { min: 1 },
            errorMessage: "invalid duration: less than 1",
          },
        },
        "init_message.props.extension": {
          custom: {
            options: (value, { req }) =>
              (req.body.init_message.type === "file" &&
                value.startsWith(".")) ||
              !value,
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
        },
        "init_message.props.mimeType": {
          custom: {
            options: (value, { req }) =>
              !["text", "voice"].includes(req.body.init_message.type) || !value,
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          isMimeType: {
            if: (value, { req }) =>
              !["text", "voice"].includes(req.body.init_message.type),
            errorMessage: "invalid mime type",
          },
        },
        "init_message.props.caption": {
          optional: true,
          custom: {
            options: (value, { req }) =>
              !["text", "voice"].includes(req.body.init_message.type) || !value,
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          notEmpty: {
            if: (value, { req }) =>
              !["text", "voice"].includes(req.body.init_message.type),
          },
        },
        "init_message.props.size": {
          custom: {
            options: (value, { req }) =>
              !["text", "voice"].includes(req.body.init_message.type) || !value,
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          isInt: {
            if: (value, { req }) =>
              !["text", "voice"].includes(req.body.init_message.type),
            options: { min: 1, max: 10 * 1024 ** 2 /* 10mb */ },
            errorMessage: "size out of range",
          },
        },
      },
      ["body"]
    ),
    { message: "request body contains invalid fields" }
  ),
  errHandler,
]

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
            options: { min: 1, max: 10 * 1024 ** 2 },
            errorMessage:
              "value must me an array of uint8 integers with a maximum of 10mb",
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
              (req.body.msg_content.type === "file" &&
                value.startsWith(".")) ||
              !value,
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
            options: { min: 1, max: 10 * 1024 ** 2 /* 10mb */ },
            errorMessage: "size out of range",
          },
        },
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
        delivery_time: {
          notEmpty: true,
          custom: {
            options: (value) => !isNaN(Date.parse(value)),
            errorMessage:
              "invalid date",
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
            errorMessage: "invalid delete-for value, should be either 'me' or 'everyone'"
          }
        },
      },
      ["query"]
    ),
    { message: "request body contains invalid fields" }
  ),
  errHandler,
]
