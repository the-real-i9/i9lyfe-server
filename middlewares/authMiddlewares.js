import expressSession from "express-session"
import pgSession from "connect-pg-simple"
import { getDBPool } from "../models/db.js"

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 * @param {import('express').NextFunction} next
 */
export const signupProgressValidation = (req, res, next) => {
  const { stage } = req.query

  if (["token_validation", "user_registration"].includes(stage))
    confirmOngoingRegistration(req, res)

  if (stage === "token_validation") rejectConfirmedEmail(req, res)

  if (stage === "user_registration") rejectUnconfirmedEmail(req, res)

  next()
}

export const passwordResetProgressValidation = (req, res, next) => {
  const { stage } = req.query

  if (["token_validation", "password_reset"].includes(stage))
    confirmOngoingRegistration(req, res)

  if (stage === "token_validation") rejectConfirmedEmail(req, res)

  if (stage === "password_reset") rejectUnconfirmedEmail(req, res)

  next()
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 * @param {import('express').NextFunction} next
 */
const confirmOngoingRegistration = (req, res) => {
  if (!req.session.email_confirmation_data) {
    return res.status(403).send({ errorMessage: "No ongoing registration!" })
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 * @param {import('express').NextFunction} next
 */
const rejectConfirmedEmail = (req, res) => {
  if (req.session.email_confirmation_data.confirmed) {
    return res
      .status(403)
      .send({ errorMessage: "Your email has already being verified!" })
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 * @param {import('express').NextFunction} next
 */
const rejectUnconfirmedEmail = (req, res) => {
  if (!req.session.email_confirmation_data.confirmed) {
    return res
      .status(403)
      .send({ errorMessage: "Your email has not been verified!" })
  }
}

/**
 * @param {string} storeTableName
 * @param {string} sessionSecret
 * @param {string} cookiePath
 * @returns
 */

const PGStore = pgSession(expressSession)

export const expressSessionMiddleware = (
  storeTableName,
  sessionSecret,
  cookiePath
) =>
  expressSession({
    store: new PGStore({
      pool: getDBPool(),
      tableName: storeTableName,
      createTableIfMissing: true,
    }),
    resave: false,
    saveUninitialized: false,
    secret: sessionSecret,
    cookie: {
      maxAge: 1 * 60 * 60 * 1000,
      secure: false,
      path: cookiePath,
    },
  })
