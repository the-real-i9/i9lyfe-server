import * as mailService from "../mail.service.js"
import * as securityServices from "../security.services.js"
import * as messageBrokerService from "../messageBroker.service.js"
import { User } from "../../graph_models/user.model.js"

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
  if (Number(verificationCode) !== Number(inputCode)) {
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
    return { error: { msg: "Username already taken. Try another." } }

  const passwordHash = await securityServices.hashPassword(info.password)

  const user = await User.create({
    ...info,
    password: passwordHash,
  })

  const jwt = securityServices.generateJwt({
    client_user_id: user.id,
    client_username: user.username,
  })

  messageBrokerService.createTopic(`i9lyfe-user-${user.id}-alerts`)

  return {
    data: {
      msg: "Signup success! You're automatically logged in.",
      user,
      jwt,
    },
  }
}
