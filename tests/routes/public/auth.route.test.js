import { it, expect } from "@jest/globals"
import dotenv from "dotenv"
import { dbQuery } from "../../../src/models/db.js"
import supertest from "supertest"

dotenv.config()

import app from "../../../src/app.js"

const prefixPath = "/api/auth"

it("should signup user", async () => {
  const email = "oluwarinolasam@gmail.com"

  try {
    // step 1
    const step1Body = { email }

    const step1Res = await supertest(app)
      .post(prefixPath + "/signup/request_new_account")
      .send(step1Body)

    expect(step1Res.body).toHaveProperty(
      "msg",
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
    const step2Res = await supertest(app)
      .post(prefixPath + "/signup/verify_email")
      .set("Cookie", [signupCookie])
      .send(step2Body)

    expect(step2Res.body).toHaveProperty("msg", `Your email ${email} has been verified!`)

    // step3
    const step3Body = {
      username: "i9x",
      password: "fhunmytor",
      name: "Samuel Oluwarinola",
      birthday: new Date(2000, 10, 7),
      bio: "#nerdIsLife",
    }

    const step3Res = await supertest(app)
      .post(prefixPath + "/signup/register_user")
      .set("Cookie", [signupCookie])
      .send(step3Body)

    expect(step3Res.body).toHaveProperty("jwt")
  } finally {
    // cleanup
    dbQuery({
      text: `DELETE FROM i9l_user WHERE email = $1`,
      values: [email],
    })
  }
}, 5000)

it("should signin user", async () => {
  const email_or_username = "johnny@gmail.com"

  const body = {
    email_or_username,
    password: "fhunmytor",
  }
  const res = await supertest(app)
  .post(prefixPath + "/signin")
  .send(body)

  expect(res.body).toHaveProperty("jwt")
}, 5000)
