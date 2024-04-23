import { test, xtest, expect } from "@jest/globals"
import axios from "axios"
import dotenv from "dotenv"

dotenv.config()

const email = "oluwarinolasam@gmail.com"
const prefixPath = "http://localhost:5000/api/auth"
const signupCookie =
  "connect.sid=s%3A_7fE748pXm1V2X-JSKdE7Dhhj4ITLtgi.vKTYVGgqkYFS89rCBqtpr8B50acqD%2Bm5sef6zoOsUl4; Path=/api/auth/signup; Expires=Tue, 23 Apr 2024 20:51:36 GMT; HttpOnly"

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
  const code = 569475
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
    username: "i9x",
    password: "fhunmytor",
    name: "Kenny Samuel",
    birthday: new Date(2000, 10, 7),
    bio: "Testing testing!",
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