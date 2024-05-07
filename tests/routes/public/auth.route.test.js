import { test, xtest, expect } from "@jest/globals"
import axios from "axios"
import dotenv from "dotenv"

dotenv.config()

const email = "oluwarinolasam@gmail.com"
const prefixPath = "http://localhost:5000/api/auth"
const signupCookie =
  "connect.sid=s%3ABOJNeD7RkIC9ALDHhhg_BrWL0FmqKXfy.mE0iAUCNtO7ZbWe9j1EOm7I0k%2BuI721BEJtjAAVZBl8; Path=/api/auth/signup; Expires=Tue, 07 May 2024 22:10:15 GMT; HttpOnly"

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
  const reqData = { code: 536029 }
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
    username: "i9x",
    password: "fhunmytor",
    name: "Kehinde Ogunrinola",
    birthday: new Date(2000, 10, 7),
    bio: "Ingenious!",
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

test("signin", async () => {
  const reqData = { email: "oluwarinolasam@gmail.com", password: "fhunmytor" }
  const res = await axios.post(prefixPath + "/signin", reqData)

  if (res.status === 200) {
    console.log(res.data)
  }

  expect(res.status).toBe(200)
  expect(res.data.msg).toBe("Signin success!")
})
