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

  if (["email_verification", "user_registration"].includes(stage))
    confirmOngoingRegistration(res, req.session.email_verification_data)

  if (stage === "email_verification") rejectConfirmedEmail(res, req.session.email_verification_data.verified)

  if (stage === "user_registration") rejectUnconfirmedEmail(res, req.session.email_verification_data.verified)

  return next()
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 * @param {import('express').NextFunction} next
 */
export const passwordResetProgressValidation = (req, res, next) => {
  const { stage } = req.query

  if (["email_confirmation", "password_reset"].includes(stage))
    confirmOngoingRegistration(res, req.session.password_reset_email_confirmation_data)

  if (stage === "email_confirmation") rejectConfirmedEmail(res, req.session.password_reset_email_confirmation_data.emailConfirmed)

  if (stage === "password_reset") rejectUnconfirmedEmail(res, req.session.password_reset_email_confirmation_data.emailConfirmed)

  next()
}

/**
 * @param {import('express').Response} res
 * @param {Object} sessionData
 */
const confirmOngoingRegistration = (res, sessionData) => {
  if (!sessionData) {
    return res.status(403).send({ errorMessage: "No ongoing registration!" })
  }
}

/**
 * @param {import('express').Response} res
 * @param {boolean} emailValidationStatus
 */
const rejectConfirmedEmail = (res, emailValidationStatus) => {
  if (emailValidationStatus) {
    return res
      .status(403)
      .send({ errorMessage: "Your email has already being verified!" })
  }
}

/**
 * @param {import('express').Response} res
 * @param {boolean} emailValidationStatus
 */
const rejectUnconfirmedEmail = (res, emailValidationStatus) => {
  if (!emailValidationStatus) {
    return res
      .status(403)
      .send({ errorMessage: "Your email has not been verified!" })
  }
}


const PGStore = pgSession(expressSession)

/**
 * @param {string} storeTableName
 * @param {string} sessionSecret
 * @param {string} cookiePath
 * @returns
 */
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
