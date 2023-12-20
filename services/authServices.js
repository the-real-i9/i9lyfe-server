import { getUserByEmail } from "../models/userModel.js"
import { generateCodeWithExpiration } from "../utils/helpers.js"
import sendMail from "./mailingService.js"

/** @param {string} email */
export const userAlreadyExists = async (email) => {
  const result = await getUserByEmail(email, "1")
  return result.rowCount > 0 ? true : false
}

/** @param {string} email */
export const newAccountRequestService = async (email) => {
  try {
    if (await userAlreadyExists(email))
      return {
        ok: false,
        err: {
          code: 422,
          reason: "A user with this email already exists",
        },
        data: null,
      }

    const [verfCode, verfCodeExpiration] = generateCodeWithExpiration()

    sendMail({
      to: email,
      subject: "i9lyfe - Verify your email",
      html: `<p>Your email verification code is <strong>${verfCode}</strong></p>`,
    })

    return {
      ok: true,
      err: null,
      verfData: {
        email,
        verified: false,
        verfCode,
        verfCodeExpiration,
      },
    }
  } catch (error) {
    console.log(error)
    return {
      ok: false,
      err: {
        code: 500,
        reason: "Internal Server Error",
      },
      verfData: null,
    }
  }
}

/**
 * @param {number} verfCode
 * @param {number} userInputCode
*/
const codesMatch = (verfCode, userInputCode) => verfCode === userInputCode

/** @param {Date} verfCodeExpiration */
const codeLives = (verfCodeExpiration) => Date.now() < new Date(verfCodeExpiration)

/**
 * @param {Object} sessionUserVerfInfo
 * @param {string} sessionUserVerfInfo.email
 * @param {number} sessionUserVerfInfo.verfCode
 * @param {Date} sessionUserVerfInfo.verfCodeExpiration
 * @param {number} userInputCode
 */
export const emailVerificationService = (sessionUserVerfData, userInputCode) => {
  const { email, verfCode, verfCodeExpiration } = sessionUserVerfData

  if (!codesMatch(verfCode, userInputCode)) {
    return {
      ok: false,
      err: {
        code: 422,
        reason:
          "Incorrect Verification Code! Check your email or Resubmit your email.",
      },
      updatedVerfdata: null,
    }
  }

  if (!codeLives(verfCodeExpiration)) {
    return {
      ok: false,
      err: {
        code: 422,
        reason: "Verification Code Expired! Resubmit your email .",
      },
      updatedVerfdata: null,
    }
  }

  sendMail({
    to: email,
    subject: "i9lyfe - Email verification success",
    html: `<p>Your email <strong>${email}</strong> verification has been verified!</p>`,
  })

  return {
    ok: true,
    err: null,
    updatedVerfdata: {
      email,
      verified: true,
      verfCode: null,
      verfCodeExpiration: null,
    }
  }
}
