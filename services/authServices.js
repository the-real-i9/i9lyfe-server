import bcrypt from "bcrypt"

import { createNewUser, getUserByEmail } from "../models/userModel.js"
import {
  generateCodeWithExpiration,
  generateJwtToken,
} from "../utils/helpers.js"
import sendMail from "./mailingService.js"

/** @param {string} email */
export const userExists = async (email) => {
  const result = await getUserByEmail(email, "1")
  return result.rowCount > 0 ? true : false
}

/** @param {string} email */
export const newAccountRequestService = async (email) => {
  try {
    if (await userExists(email))
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
      data: {
        verfData: {
          email,
          verified: false,
          verfCode,
          verfCodeExpiration,
        },
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
      data: null,
    }
  }
}

/**
 * @param {number} verfCode
 * @param {number} userInputCode
 */
const codesMatch = (verfCode, userInputCode) => verfCode === userInputCode

/** @param {Date} verfCodeExpiration */
const codeLives = (verfCodeExpiration) =>
  Date.now() < new Date(verfCodeExpiration)

/**
 * @param {Object} sessionUserVerfInfo
 * @param {string} sessionUserVerfInfo.email
 * @param {number} sessionUserVerfInfo.verfCode
 * @param {Date} sessionUserVerfInfo.verfCodeExpiration
 * @param {number} userInputCode
 */
export const emailVerificationService = (
  sessionUserVerfData,
  userInputCode
) => {
  const { email, verfCode, verfCodeExpiration } = sessionUserVerfData

  if (!codesMatch(verfCode, userInputCode)) {
    return {
      ok: false,
      err: {
        code: 422,
        reason:
          "Incorrect Verification Code! Check your email or Resubmit your email.",
      },
      data: null,
    }
  }

  if (!codeLives(verfCodeExpiration)) {
    return {
      ok: false,
      err: {
        code: 422,
        reason: "Verification Code Expired! Resubmit your email .",
      },
      data: null,
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
    data: {
      updatedVerfdata: {
        email,
        verified: true,
        verfCode: null,
        verfCodeExpiration: null,
      },
    },
  }
}

/** @param {string} password */
export const hashPassword = async (password) => {
  return await bcrypt.hash(password, 10)
}

/**
 * @param {Object} userDataInput
 * @param {string} userDataInput.email
 * @param {string} userDataInput.username
 * @param {string} userDataInput.password
 * @param {string} userDataInput.name
 * @param {Date} userDataInput.birthday
 * @param {string} userDataInput.bio
 */
export const userRegistrationService = async (userDataInput) => {
  try {
    const passwordHash = await hashPassword(userDataInput.password)

    const result = await createNewUser({
      ...userDataInput,
      password: passwordHash,
      birthday: new Date(userDataInput.birthday),
    })

    const userData = result.rows[0]

    const jwtToken = generateJwtToken({
      user_id: userData.id,
      email: userData.email,
    })

    return { ok: true, err: null, data: { userData, jwtToken } }
  } catch (error) {
    console.log(error)
    return {
      ok: false,
      err: { code: 500, reason: "Internal Server Error" },
      data: null,
    }
  }
}

/**
 * @param {string} passwordInput Password supplied by user
 * @param {string} passwordHash The hashed version stored in database
 * @returns The compare result; true if both match and false otherwise
 */
const passwordMatch = async (passwordInput, passwordHash) => {
  return await bcrypt.compare(passwordInput, passwordHash)
}

/**
 * @param {string} email
 * @param {string} passwordInput
 */
export const userSigninService = async (email, passwordInput) => {
  try {
    const result = await getUserByEmail(
      email,
      "id email username password name profile_pic_url"
    )

    if (result.rowCount === 0) {
      return {
        ok: false,
        err: { code: 422, reason: "Incorrect email or password" },
        data: null,
      }
    }

    const { password: passwordHash, ...userData } = result.rows[0]
    if (!(await passwordMatch(passwordInput, passwordHash))) {
      return {
        ok: false,
        err: { code: 422, reason: "Incorrect email or password" },
        data: null,
      }
    }

    const jwtToken = generateJwtToken({
      user_id: userData.id,
      email: userData.email,
    })
    return {
      ok: true,
      err: null,
      data: {
        userData, // observe that password has been excluded above
        jwtToken,
      },
    }
  } catch (error) {
    console.log(error)
    return {
      ok: false,
      err: { code: 500, reason: "Internal Server Error" },
      data: null,
    }
  }
}
