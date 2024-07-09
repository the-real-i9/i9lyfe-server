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
export function editProfile(req, res, next) {
  const props = Object.keys(req.body)

  if (!props.length) {
    return res.status(422).send({
      error: `empty fields provided`,
    })
  }

  if (!props.every((p) => ["name", "birthday", "bio"].includes(p))) {
    return res.status(422).send({
      error: `you can only change your "name", "birthday" or "bio" through here`,
    })
  }

  if (req.body.birthday) {
    const v = new Input("birthday", req.body.birthday).isDate()
    if (v.error) {
      return res.status(422).send({ error: v.error })
    }
  }

  if (req.body.name) {
    const v = new Input("name", req.body.name).notEmpty().min(1)
    if (v.error) {
      return res.status(422).send({ error: v.error })
    }
  }

  return next()
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 * @param {NextFunction} next
 */
export function updateConnectionStatus(req, res, next) {
  const { connection_status, last_active } = req.body

  if (!["online", "offline"].includes(connection_status)) {
    return res.status(422).send({
      error: {
        field: "connection_status",
        msg: "connection_status's value must be either 'online' or 'offline'",
      },
    })
  }

  if (connection_status === "online" && last_active) {
    return res.status(422).send({
      error: {
        field: "last_active",
        msg: "connection_status of 'online' cannot have a last_active",
      },
    })
  }

  if (connection_status === "offline") {
    const v = new Input("last_active", last_active).notEmpty().isDate()
    if (v.error) {
      return res.status(422).send({ error: v.error })
    }
  }

  return next()
}
