import expressSession from "express-session"
import pgSession from "connect-pg-simple"
import { getDBPool } from "../models/db.js"

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 * @param {import('express').NextFunction} next
 */
export const signupProgressValidation = (req, res, next) => {
  const { step } = req.query

  if (["verify_email", "register_user"].includes(step))
    confirmOngoingRegistration(req, res)

  if (step === "verify_email") rejectVerifiedEmail(req, res)

  if (step === "register_user") rejectUnverifiedEmail(req, res)

  next()
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 * @param {import('express').NextFunction} next
 */
const confirmOngoingRegistration = (req, res) => {
  if (!req.session.potential_user_verification_data) {
    return res.status(403).send({ errorMessage: "No ongoing registration!" })
  }
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 * @param {import('express').NextFunction} next
 */
const rejectVerifiedEmail = (req, res) => {
  if (req.session.potential_user_verification_data.verified) {
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
const rejectUnverifiedEmail = (req, res) => {
  if (!req.session.potential_user_verification_data.verified) {
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
