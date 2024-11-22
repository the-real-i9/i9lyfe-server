import * as signupService from "../../services/auth/signup.service.js"

export const requestNewAccount = async (req, res) => {
  const { email } = req.body

  try {
    const resp = await signupService.requestNewAccount(email)

    if (resp.error) return res.status(400).send(resp.error)

    req.session.signup = {
      step: "verify email",
      data: {
        email,
        verified: false,
        verificationCode: resp.verificationCode,
        verificationCodeExpires: resp.verificationCodeExpires,
      },
    }

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const verifyEmail = async (req, res) => {
  const { code: inputCode } = req.body

  try {
    if (req.session?.signup.step != "verify email")
      return res.status(400).send({ msg: "Invalid cookie at endpoint" })

    const signupSessionData = req.session.signup.data

    const resp = signupService.verifyEmail({ inputCode, ...signupSessionData })

    if (resp.error) return res.status(400).send(resp.error)

    req.session.signup = {
      step: "register user",
      data: {
        email: signupSessionData.email,
        verified: true,
      },
    }

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const registerUser = async (req, res) => {
  try {
    if (req.session?.signup.step != "verify email")
      return res.status(400).send({ msg: "Invalid cookie at endpoint" })

    const { email } = req.session.signup.data

    const resp = await signupService.registerUser({ email, ...req.body })

    if (resp.error) return res.status(400).send(resp.error)

    req.session.destroy()

    res.status(201).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
