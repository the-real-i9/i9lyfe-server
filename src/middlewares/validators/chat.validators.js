import { checkExact, checkSchema, param } from "express-validator"
import { errHandler } from "./miscs.js"

export const validateIdParams = [
  param("**").isInt().withMessage("expected integer value"),
  errHandler,
]

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
            options: /^[\w-]{3,}$/,
            errorMessage: "invalid username format",
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
            options: ["text", "voice", "image", "audio", "video", "file"],
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
            options: (value, { req }) => req.body.init_message.type === "text",
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          notEmpty: true,
        },
        "init_message.props.data": {
          custom: {
            options: (value, { req }) => req.body.init_message.type !== "text",
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          isArray: { options: { min: 1 } },
        },
        "init_message.props.duration": {
          custom: {
            options: (value, { req }) => req.body.init_message.type === "voice",
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          isInt: { options: { min: 1 } },
        },
        "init_message.props.extension": {
          custom: {
            options: (value, { req }) =>
              req.body.init_message.type === "file" && value.startsWith("."),
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
        },
        "init_message.props.mimeType": {
          custom: {
            options: (value, { req }) =>
              !["text", "voice"].includes(req.body.init_message.type),
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          isMimeType: true,
        },
        "init_message.props.caption": {
          optional: true,
          custom: {
            options: (value, { req }) =>
              !["text", "voice"].includes(req.body.init_message.type),
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          notEmpty: true,
        },
        "init_message.props.size": {
          custom: {
            options: (value, { req }) =>
              !["text", "voice"].includes(req.body.init_message.type),
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          isInt: { options: { min: 1, max: 10 * 1024 ** 2 /* 10mb */ } },
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
            options: ["text", "voice", "image", "audio", "video", "file"],
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
            options: (value, { req }) => req.body.msg_content.type === "text",
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          notEmpty: true,
        },
        "msg_content.props.data": {
          custom: {
            options: (value, { req }) => req.body.msg_content.type !== "text",
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          isArray: { options: { min: 1 } },
        },
        "msg_content.props.duration": {
          custom: {
            options: (value, { req }) => req.body.msg_content.type === "voice",
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          isInt: { options: { min: 1 } },
        },
        "msg_content.props.extension": {
          custom: {
            options: (value, { req }) =>
              req.body.msg_content.type === "file" && value.startsWith("."),
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
        },
        "msg_content.props.mimeType": {
          custom: {
            options: (value, { req }) =>
              !["text", "voice"].includes(req.body.msg_content.type),
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          isMimeType: true,
        },
        "msg_content.props.caption": {
          optional: true,
          custom: {
            options: (value, { req }) =>
              !["text", "voice"].includes(req.body.msg_content.type),
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          notEmpty: true,
        },
        "msg_content.props.size": {
          custom: {
            options: (value, { req }) =>
              !["text", "voice"].includes(req.body.msg_content.type),
            errorMessage: "invalid property for the specified type",
            bail: true,
          },
          isInt: { options: { min: 1, max: 10 * 1024 ** 2 /* 10mb */ } },
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
          isString: true,
        },
      },
      ["body"]
    ),
    { message: "request body contains invalid fields" }
  ),
  errHandler,
]
