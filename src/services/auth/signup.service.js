import * as mailService from "../mail.service.js"
import * as securityServices from "../security.services.js"
import { User } from "../../models/user.model.js"

export const requestNewAccount = async (email) => {
  if (await User.exists(email))
    return {
      error: { msg: "A user with this email already exists." },
    }

  const { token: verificationCode, expires: verificationCodeExpires } =
    securityServices.generateTokenWithExpiration()

  mailService.sendMail({
    to: email,
    subject: "i9lyfe - Verify your email",
    html: `<p>Your email verification code is <strong>${verificationCode}</strong></p>`,
  })

  return {
    verificationCode,
    verificationCodeExpires,
    data: {
      msg: `Enter the 6-digit code sent to ${email} to verify your email`,
    },
  }
}

export const verifyEmail = ({
  email,
  inputCode,
  verificationCode,
  verificationCodeExpires,
}) => {
  if (verificationCode !== inputCode) {
    return {
      error: {
        msg: "Incorrect verification code! Check or Re-submit your email.",
      },
    }
  }

  if (!securityServices.isTokenAlive(verificationCodeExpires)) {
    return {
      error: {
        msg: "Verification code expired! Re-submit your email.",
      },
    }
  }

  mailService.sendMail({
    to: email,
    subject: "i9lyfe - Email verification success",
    html: `<p>Your email <strong>${email}</strong> has been verified!</p>`,
  })

  return {
    data: { msg: `Your email ${email} has been verified!` },
  }
}

export const registerUser = async (info) => {
  if (await User.exists(info.username))
    return { error: { msg: "Username not available" } }

  const passwordHash = await securityServices.hashPassword(info.password)

  const user = await User.create({
    ...info,
    password: passwordHash,
  })

  const jwt = securityServices.signJwt({
    client_username: user.username,
  })

  return {
    data: {
      msg: "Signup success! You're automatically logged in.",
      user,
    },
    jwt,
  }
}
