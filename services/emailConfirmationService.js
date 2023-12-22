import { generateTokenWithExpiration } from "../utils/helpers"

/**
 * @param {import('express').Request} req
 * @param {string} confirmationStage
 * @param {import('./mailingService.js').PrimaryMailSender} primaryMailSender
 * @param {import('./mailingService.js').TokenMailSender} tokenMailSender
 */
export const emailConfirmationService = async (
  req,
  confirmationStage,
  primaryMailSender,
  tokenMailSender,
) => {
  try {
    /* Abstract here */
    if (confirmationStage === "email submission") {
      const { email } = req.body
      const [token, tokenExpires] = generateTokenWithExpiration()

      // users of this service each have varying message styles which includes this token,
      // therefore we employed patterns and principles
      // OO Principles: Dependency inversion, Dependency injection
      tokenMailSender.sendToken(email, token)

      req.session.email_confirmation_data = {
        email,
        confirmed: false,
        confirmationToken: token,
        confirmationTokenExpires: tokenExpires,
        confirmationStage: "token validation",
      }

      return {
        ok: true,
        err: null,
      }
    }

    /* Abstract here */
    if (confirmationStage === "token validation") {
      const { email, confirmationToken, confirmationTokenExpires } =
        req.session.email_confirmation_data
      const { token: userInputToken } = req.body

      if (!tokensMatch(confirmationToken, userInputToken)) {
        return {
          ok: false,
          err: {
            code: 422,
            reason: "Incorrect confirmation token! Check or Re-submit your email.",
          },
        }
      }

      if (!tokenLives(confirmationTokenExpires)) {
        return {
          ok: false,
          err: {
            code: 422,
            reason: "Confirmation token expired! Re-submit your email.",
          },
        }
      }

      primaryMailSender.send(email)

      req.session.email_confirmation_data = {
        email,
        confirmed: true,
        confirmationToken: null,
        confirmationTokenExpires: null,
        confirmationStage: "email confirmed",
      }

      return {
        ok: true,
        err: null,
      }
    }
  } catch (error) {
    console.log(error)
    return {
      ok: false,
      err: {
        code: 500,
        reason: "Internal Server Error",
      },
    }
  }
}

/**
 * @param {number} confirmationToken
 * @param {number} userInputToken
 */
const tokensMatch = (confirmationToken, userInputToken) =>
  confirmationToken === userInputToken

/** @param {Date} confirmationTokenExpiration */
const tokenLives = (confirmationTokenExpiration) =>
  Date.now() < new Date(confirmationTokenExpiration)
