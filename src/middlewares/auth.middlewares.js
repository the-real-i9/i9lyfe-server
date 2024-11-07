import expressSession from "express-session"
import pgSession from "connect-pg-simple"
import { getDBPool } from "../configs/db.js"

export const proceedEmailVerification = (req, res, next) => {
  const signupInSession = req.session.email_verification_state

  if (!signupInSession) {
    return res.status(403).send({ errorMessage: "No ongoing registration!" })
  }

  const emailIsVerified = req.session.email_verification_state.verified

  if (emailIsVerified) {
    return res
      .status(403)
      .send({ errorMessage: "Your email has already being verified!" })
  }

  return next()
}

export const proceedUserRegistration = (req, res, next) => {
  const signupInSession = req.session.email_verification_state

  if (!signupInSession) {
    return res.status(403).send({ errorMessage: "No ongoing registration!" })
  }

  const emailIsVerified = req.session.email_verification_state.verified

  if (!emailIsVerified) {
    return res
      .status(403)
      .send({ errorMessage: "Your email has not been verified!" })
  }

  return next()
}

export const proceedEmailConfirmation = (req, res, next) => {
  const passwordResetInSession =
    req.session.password_reset_email_confirmation_state

  if (!passwordResetInSession) {
    return res.status(403).send({ errorMessage: "No ongoing password reset!" })
  }

  const emailIsConfirmed =
    req.session.password_reset_email_confirmation_state.emailConfirmed

  if (emailIsConfirmed) {
    return res
      .status(403)
      .send({ errorMessage: "Your email has already being confirmed!" })
  }

  return next()
}

export const proceedPasswordReset = (req, res, next) => {
  const passwordResetInSession =
    req.session.password_reset_email_confirmation_state

  if (!passwordResetInSession) {
    return res.status(403).send({ errorMessage: "No ongoing password reset!" })
  }

  const emailIsConfirmed =
    req.session.password_reset_email_confirmation_state.emailConfirmed

  if (!emailIsConfirmed) {
    return res
      .status(403)
      .send({ errorMessage: "Your email has not been confirmed!" })
  }

  return next()
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
