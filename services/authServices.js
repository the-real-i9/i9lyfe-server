import bcrypt from "bcrypt"
import {
  changeUserPassword,
  createNewUser,
  getUserByEmail,
} from "../models/userModel.js"
import { generateJwtToken } from "../utils/helpers.js"
import sendMail from "./mailingService.js"

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

    return {
      ok: true,
      error: null,
      data: { userData, jwtToken },
    }
  } catch (error) {
    console.log(error)
    return {
      ok: false,
      error: { code: 500, reason: "Internal Server Error" },
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

export const passwordResetService = async (userEmail, newPassword) => {
  try {
    const passwordHash = await hashPassword(newPassword)

    await changeUserPassword(userEmail, passwordHash)

    sendMail({
      to: userEmail,
      subject: "i9lyfe - Password reset successful",
      html: `<p>${userEmail}, your password has been changed successfully!</p>`,
    })

    return {
      ok: true,
      err: null,
      data: null,
    }
  } catch (error) {
    console.log(error)
    return {
      ok: false,
      err: {
        code: 500,
        reason: "Internal server error",
      },
      data: null,
    }
  }
}
