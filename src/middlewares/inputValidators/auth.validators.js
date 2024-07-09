import Input from "./Input.js"

/**
 * @typedef {import("express").Request} ExpressRequest
 * @typedef {import("express").Response} ExpressResponse
 * @typedef {import("express").NextFunction} NextFunction
 */

export function requestNewAccount(req, res, next) {
  const { email } = req.body

  const v = new Input("email", email).notEmpty().isEmail()
  if (v.error) {
    return res.status(422).send({ error: v.error })
  }

  return next()
}

export function verifyEmail(req, res, next) {
  const { code } = req.body

  const v = new Input("code", code).notEmpty().isNumeric()
  if (v.error) {
    return res.status(422).send({ error: v.error })
  }

  return next()
}

export function registerUser(req, res, next) {
  const { username, name, password, birthday } = req.body

  let v = new Input("username", username).notEmpty().isValidUsername()

  if (v.error) {
    return res.status(422).send({ error: v.error })
  }

  v = new Input("name", name).notEmpty().min(1)

  if (v.error) {
    return res.status(422).send({ error: v.error })
  }

  v = new Input("password", password).notEmpty().min(8)
  if (v.error) {
    return res.status(422).send({ error: v.error })
  }

  v = new Input("birthday", birthday).notEmpty().isDate()
  if (v.error) {
    return res.status(422).send({ error: v.error })
  }

  return next()
}

export function signin(req, res, next) {
  const { emailOrUsername, password } = req.body

  let v = new Input("emailOrUsername", emailOrUsername).notEmpty()

  if (v.error) {
    return res.status(422).send({ error: v.error })
  }

  v.isEmail()

  if (v.error) {
    // is not email, but it could be a valid username
    v.error = null // reset error value

    v.isValidUsername()

    if (v.error) {
      // is not a valid username either
      v.error.msg = "invalid emailOrUsername value"
      return res.status(422).send({ error: v.error })
    }
  }

  v = new Input("password", password).notEmpty()
  if (v.error) {
    return res.status(422).send({ error: v.error })
  }

  return next()
}

export function requestPasswordReset(req, res, next) {
  const { email } = req.body

  const v = new Input("email", email).notEmpty().isEmail()
  if (v.error) {
    return res.status(422).send({ error: v.error })
  }

  return next()
}

export function confirmEmail(req, res, next) {
  const { code } = req.body

  const v = new Input("code", code).notEmpty().isNumeric()
  if (v.error) {
    return res.status(422).send({ error: v.error })
  }

  return next()
}

/**
 * @param {ExpressRequest} req
 * @param {ExpressResponse} res
 * @param {NextFunction} next
 */
export function resetPassword(req, res, next) {
  const { newPassword, confirmNewPassword } = req.body

  if (newPassword !== confirmNewPassword) {
    return res.status(422).send({ error: { field: "cofirmNewPassword", msg: "password mismatch" } })
  }

  const v = new Input("newPassword").notEmpty().min(8)

  if (v.error) {
    return res.status(422).send({ error: v.error })
  }

  return next()
}
