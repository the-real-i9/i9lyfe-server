import { test, xtest, expect } from "@jest/globals"
import axios from "axios"
import dotenv from "dotenv"

dotenv.config()

const email = "ogunrinola.kehinde@yahoo.com"
const prefixPath = "http://localhost:5000/api/auth"
const signupCookie =
  "connect.sid=s%3AnXZ6Bt7lkIPx77CESXjgBPtwKasZz0Nw.JsZgFutixH6l%2BKEB9sDCNswKoA4BuNsb9ogjzgU5pq4; Path=/api/auth/signup; Expires=Tue, 07 May 2024 22:18:18 GMT; HttpOnly"

xtest("signup: request new account", async () => {
  const reqData = { email }
  const res = await axios.post(
    prefixPath + "/signup/request_new_account",
    reqData
  )

  if (res.status === 200) {
    console.log(res.headers["set-cookie"])
  }

  expect(res.status).toBe(200)
  expect(res.data.msg).toBe(
    `Enter the 6-digit code sent to ${email} to verify your email`
  )
})

xtest("signup: verify email", async () => {
  const reqData = { code: 310718 }
  const res = await axios.post(prefixPath + "/signup/verify_email", reqData, {
    headers: {
      Cookie: signupCookie,
    },
  })

  expect(res.status).toBe(200)
  expect(res.data.msg).toBe(`Your email ${email} has been verified!`)
})

xtest("signup: register user", async () => {
  const reqData = {
    username: "dollyp",
    password: "fhunmytor",
    name: "Dolapo Olaleye",
    birthday: new Date(1999, 10, 7),
    bio: "Nerdy!",
  }

  const res = await axios.post(prefixPath + "/signup/register_user", reqData, {
    headers: {
      Cookie: signupCookie,
    },
  })

  if (res.status === 201) {
    console.log(res.data)
  }

  expect(res.status).toBe(201)
  expect(res.data.msg).toBe(
    "Registration success! You're automatically logged in."
  )
})

xtest("signin", async () => {
  const reqData = { email: "ogunrinola.kehinde@yahoo.com", password: "fhunmytor" }
  const res = await axios.post(prefixPath + "/signin", reqData)

  if (res.status === 200) {
    console.log(res.data)
  }

  expect(res.status).toBe(200)
  expect(res.data.msg).toBe("Signin success!")
})
