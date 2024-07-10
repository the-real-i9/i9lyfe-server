/**
 * @typedef {import("express").Request} ExpressRequest
 * @typedef {import("express").Response} ExpressResponse
 * @typedef {import("express").NextFunction} NextFunction
 */

import Input from "./Input.js"

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 * @param {NextFunction} next
 */

export function validateIdParams(req, res, next) {
  const params = Object.keys(req.params)

  for (const param of params) {
    const v = new Input(param, req.params[param]).isNumeric()

    if (v.error) {
      return res.status(422).send({error: v.error})
    }
  }

  return next()
}

export function createConversation(req, res, next) {
  const { partner, init_message } = req.body

  
}

export function sendMessage(req, res, next) {
  const { msg_content } = req.body


}

export function ackMessageDelivered(req, res, next) {
  const { delivery_time } = req.body


}
export function reactToMessage(req, res, next) {
  const { reaction } = req.body


}