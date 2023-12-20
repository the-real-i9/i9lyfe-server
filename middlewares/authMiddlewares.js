/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 * @param {import('express').NextFunction} next
 */
export const confirmOngoingRegistration = (req, res, next) => {
  if (!req.session.potential_user_verification_data) {
    return res.status(403).send({ errorMessage: "No ongoing registration!" })
  }

  next()
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 * @param {import('express').NextFunction} next
 */
export const rejectVerifiedEmail = (req, res, next) => {
  if (req.session.potential_user_verification_data.verified) {
    return res
      .status(403)
      .send({ errorMessage: "Your email has already being verified!" })
  }

  next()
}

/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 * @param {import('express').NextFunction} next
 */
export const rejectUnverifiedEmail = (req, res, next) => {
  if (!req.session.potential_user_verification_data.verified) {
    return res
      .status(403)
      .send({ errorMessage: "Your email has not been verified!" })
  }

  next()
}