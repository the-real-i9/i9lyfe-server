import * as authServices from "../../services/auth.services.js"
import * as mailService from "../../services/mail.service.js"
import * as messageBrokerService from "../../services/messageBroker.service.js"
import { User } from "../../models/user.model.js"

export const requestNewAccount = async (req, res) => {
  const { email } = req.body

  try {
    if (await User.exists(email))
      return res
        .status(422)
        .send({ msg: "A user with this email already exists." })

    const [code, codeExpires] = authServices.generateCodeWithExpiration()

    mailService.sendMail({
      to: email,
      subject: "i9lyfe - Verify your email",
      html: `<p>Your email verification code is <strong>${code}</strong></p>`,
    })

    req.session.email_verification_state = {
      email,
      verified: false,
      verificationCode: code,
      verificationCodeExpires: codeExpires,
    }

    res
      .status(200)
      .send({
        msg: `Enter the 6-digit code sent to ${email} to verify your email`,
      })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const verifyEmail = async (req, res) => {
  const { code } = req.body

  try {
    const { email, verificationCode, verificationCodeExpires } =
      req.session.email_verification_state

    if (Number(verificationCode) !== Number(code)) {
      return res
        .status(422)
        .send({
          msg: "Incorrect verification code! Check or Re-submit your email.",
        })
    }

    if (!authServices.isTokenAlive(verificationCodeExpires)) {
      return res
        .status(422)
        .send({ msg: "Verification code expired! Re-submit your email." })
    }

    mailService.sendMail({
      to: email,
      subject: "i9lyfe - Email verification success",
      html: `<p>Your email <strong>${email}</strong> has been verified!</p>`,
    })

    req.session.email_verification_state = {
      email,
      verified: true,
      verificationCode: null,
      verificationCodeExpires: null,
    }

    res.status(200).send({ msg: `Your email ${email} has been verified!` })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const registerUser = async (req, res) => {
  try {
    const { email } = req.session.email_verification_state

    const info = { email, ...req.body }

    if (await User.exists(info.username)) {
      return res
        .status(422)
        .send({ msg: "Username already taken. Try another." })
    }

    const passwordHash = await authServices.hashPassword(info.password)

    const user = await User.create({
      ...info,
      password: passwordHash,
      birthday: new Date(info.birthday),
    })

    const jwt = authServices.generateJwt({
      client_user_id: user.id,
      client_username: user.username,
    })

    req.session.destroy()

    messageBrokerService.createTopic(`user-${user.id}-alerts`)

    res.status(201).send({
      msg: "Registration success! You're automatically logged in.",
      user,
      jwt,
    })
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
