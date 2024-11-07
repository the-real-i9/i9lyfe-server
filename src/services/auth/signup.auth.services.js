import { User } from "../../models/user.model"
import { generateCodeWithExpiration, generateJwt, hashPassword, tokenLives } from "../../utils/helpers"
import sendMail from "../mail.service"

export const requestNewAccount = async (email) => {
  if (await User.exists(email))
    return {
      ok: false,
      error: {
        code: 422,
        msg: "A user with this email already exists.",
      },
      data: null,
    }

  const [code, codeExpires] = generateCodeWithExpiration()

  sendMail({
    to: email,
    subject: "i9lyfe - Verify your email",
    html: `<p>Your email verification code is <strong>${code}</strong></p>`,
  })

  return {
    ok: true,
    error: null,
    data: {
      msg: `Enter the 6-digit code sent to ${email} to verify your email`,
      sessionData: {
        email,
        verified: false,
        verificationCode: code,
        verificationCodeExpires: codeExpires,
      },
    },
  }
}

export const verifyEmail = async (inputCode, sessionData) => {
  const { email, verificationCode, verificationCodeExpires } = sessionData

    if (Number(verificationCode) !== Number(inputCode)) {
      return {
        ok: false,
        error: {
          code: 422,
          msg: "Incorrect verification code! Check or Re-submit your email.",
        },
        data: null,
      }
    }

    if (!tokenLives(verificationCodeExpires)) {
      return {
        ok: false,
        error: {
          code: 422,
          msg: "Verification code expired! Re-submit your email.",
        },
        data: null,
      }
    }

    sendMail({
      to: email,
      subject: "i9lyfe - Email verification success",
      html: `<p>Your email <strong>${email}</strong> has been verified!</p>`,
    })

    return {
      ok: true,
      error: null,
      data: {
        msg: `Your email ${email} has been verified!`,
        sessionData: {
          email,
          verified: true,
          verificationCode: null,
          verificationCodeExpires: null,
        },
      },
    }
}

/**
 * @param {object} info
 * @param {string} info.email
 * @param {string} info.username
 * @param {string} info.password
 * @param {string} info.name
 * @param {Date} info.birthday
 * @param {string} info.bio
 */
export const registerUser = async (info) => {
  if (await User.exists(info.username)) {
    return {
      ok: false,
      error: {
        code: 422,
        msg: "Username already taken. Try another.",
      },
      data: null,
    }
  }

  const passwordHash = await hashPassword(info.password)

  const user = await User.create({
    ...info,
    password: passwordHash,
    birthday: new Date(info.birthday),
  })

  const jwt = generateJwt({
    client_user_id: user.id,
    client_username: user.username,
  })

  return {
    ok: true,
    error: null,
    data: {
      msg: "Registration success! You're automatically logged in.",
      user,
      jwt,
    },
  }
}