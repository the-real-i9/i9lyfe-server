import * as signinService from "../../services/auth/signin.service.js"

export const signin = async (req, res) => {
  try {
    const { email_or_username, password: inputPassword } = req.body

    const resp = await signinService.signin(
      email_or_username,
      inputPassword
    )

    if (resp.error) return res.status(400).send(resp.error)

    req.session.user = { authJwt: resp.jwt }

    res.status(200).send(resp.data)
  } catch (error) {
    console.error(error)
    res.sendStatus(500)
  }
}
