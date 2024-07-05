import { test, xtest, expect } from "@jest/globals"
import axios from "axios"
import dotenv from "dotenv"
import { dbQuery } from "../../../src/models/db.js"

dotenv.config()

const prefixPath = "http://localhost:5000/api/auth"

test("signup", async () => {
  const email = "oluwarinolasam@gmail.com"
  try {
    // step 1
    const step1Body = { email }
    const step1Res = await axios.post(
      prefixPath + "/signup/request_new_account",
      step1Body
    )

    expect(step1Res.status).toBe(200)
    expect(step1Res.data.msg).toBe(
      `Enter the 6-digit code sent to ${email} to verify your email`
    )

    // step 2
    const signupCookie = step1Res.headers["set-cookie"]
    const vcode = (
      await dbQuery({
        text: `
      SELECT sess -> 'email_verification_state' -> 'verificationCode' AS vcode
      FROM ongoing_registration 
      WHERE sess -> 'email_verification_state' ->> 'email' = $1`,
        values: [email],
      })
    ).rows[0].vcode

    const step2Body = { code: vcode }
    const step2Res = await axios.post(
      prefixPath + "/signup/verify_email",
      step2Body,
      {
        headers: {
          Cookie: signupCookie,
        },
      }
    )

    expect(step2Res.status).toBe(200)
    expect(step2Res.data.msg).toBe(`Your email ${email} has been verified!`)

    // step3
    const step3Body = {
      username: "i9",
      password: "fhunmytor",
      name: "Samuel Oluwarinola",
      birthday: new Date(2000, 11, 7),
      bio: "#nerdIsLife",
    }

    const step3Res = await axios.post(
      prefixPath + "/signup/register_user",
      step3Body,
      {
        headers: {
          Cookie: signupCookie,
        },
      }
    )

    expect(step3Res.status).toBe(201)
    expect(step3Res.data).toHaveProperty("jwt")
  } finally {
    // cleanup
    dbQuery({
      text: `DELETE FROM i9l_user WHERE email = $1`,
      values: [email],
    })
  }
})

xtest("signin", async () => {
  const email = "oluwarinolasam@gmail.com"

  const reqData = {
    email,
    password: "fhunmytor",
  }
  const res = await axios.post(prefixPath + "/signin", reqData)

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("jwt")
})
