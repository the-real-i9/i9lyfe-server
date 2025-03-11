import * as signupService from "../../services/auth/signup.service.js"

/**
 * @param {import("express").Request} req
 * @param {import("express").Response} res
 */
export const requestNewAccount = async (req, res) => {
  const { email } = req.body

  try {
    const resp = await signupService.requestNewAccount(email)

    if (resp.error) return res.status(400).send(resp.error)

    req.session.signup = {
        email,
        verified: false,
        verificationCode: resp.verificationCode,
        verificationCodeExpires: resp.verificationCodeExpires,
    }

    req.sessionOptions.maxAge = 60 * 60 * 1000
    req.sessionOptions.path = "/api/auth/signup/verify_email"

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const verifyEmail = async (req, res) => {
  const { code: inputCode } = req.body

  try {
    if (!req.session?.signup) return res.sendStatus(401)

    const signupSessionData = req.session.signup

    const resp = signupService.verifyEmail({ inputCode, ...signupSessionData })

    if (resp.error) return res.status(400).send(resp.error)

    req.session.signup = {
        email: signupSessionData.email,
        verified: true,
    }

    req.sessionOptions.maxAge = 60 * 60 * 1000
    req.sessionOptions.path = "/api/auth/signup/register_user"

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}

export const registerUser = async (req, res) => {
  try {
    if (!req.session?.signup) return res.sendStatus(401)

    const { email } = req.session.signup

    const resp = await signupService.registerUser({
      email,
      ...req.body,
    })

    if (resp.error) return res.status(400).send(resp.error)

    req.session.signup = undefined
    
    req.session.user = { authJwt: resp.jwt }

    req.sessionOptions.path = "/"
    req.sessionOptions.maxAge = 10 * 24 * 60 * 60 * 1000

    

    res.status(201).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
