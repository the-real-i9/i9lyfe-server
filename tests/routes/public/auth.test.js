import { test, xtest, expect } from "@jest/globals"
import axios from "axios"
import dotenv from "dotenv"

dotenv.config()

const email = "ogunrinola.kehinde@yahoo.com"
const prefixPath = "http://localhost:5000/api/auth"
const signupCookie =
'connect.sid=s%3AhPgd1EINQ52C_HAsQ3yGJQ20TQprO7tE.QiIQormLNi4mlmtVhHTiaaldah3JlqQ3ByWjCyeJcMY; Path=/api/auth/signup; Expires=Tue, 23 Apr 2024 22:03:50 GMT; HttpOnly'

xtest("signup: request new account", async () => {
  const res = await axios.post(prefixPath + "/signup/request_new_account", {
    email,
  })

  if (res.status === 200) {
    console.log(res.headers["set-cookie"])
  }

  expect(res.status).toBe(200)
  expect(res.data.msg).toBe(
    `Enter the 6-digit code sent to ${email} to verify your email`
  )
})

xtest("signup: verify email", async () => {
  const code = 359900
  const res = await axios.post(
    prefixPath + "/signup/verify_email",
    { code },
    {
      headers: {
        Cookie: signupCookie,
      },
    }
  )

  expect(res.status).toBe(200)
  expect(res.data.msg).toBe(`Your email ${email} has been verified!`)
})

xtest("signup: register user", async () => {
  const userInfo = {
    username: "dollyp",
    password: "fhunmytor",
    name: "Dolapo Olaleye",
    birthday: new Date(1999, 10, 7),
    bio: "Testing testing dollyp!",
  }

  const res = await axios.post(
    prefixPath + "/signup/register_user",
    { ...userInfo },
    {
      headers: {
        Cookie: signupCookie,
      },
    }
  )

  if (res.status === 201) {
    console.log(res.data)
  }

  expect(res.status).toBe(201)
  expect(res.data.msg).toBe(
    "Registration success! You're automatically logged in."
  )
})

xtest("signin", async () => {
  const res = await axios.post(prefixPath + "/signin", {
    email: "oluwarinolasam@gmail.com",
    password: "fhunmytor",
  })

  if (res.status === 200) {
    console.log(res.data)
  }

  expect(res.status).toBe(200)
  expect(res.data.msg).toBe("Signin success!")
})

test()