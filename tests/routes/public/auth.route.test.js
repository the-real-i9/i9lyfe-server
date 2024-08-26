import { it, expect } from "@jest/globals"
import dotenv from "dotenv"
import supertest from "supertest"

import { dbQuery } from "../../../src/models/db.js"
import app from "../../../src/app.js"

dotenv.config()

const prefixPath = "/api/auth"

it("should signup user", async () => {
  const email = "sample_email@gmail.com"

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

    expect(step2Res.body).toHaveProperty(
      "msg",
      `Your email ${email} has been verified!`
    )

    // step3
    const step3Body = {
      username: "i9x",
      password: process.env.TEST_USER_PASSWORD,
      name: "Samuel Oluwarinola",
      birthday: "2000-10-07",
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
    dbQuery({
      text: `
    DELETE FROM ongoing_registration 
    WHERE sess -> 'email_verification_state' ->> 'email' = $1`,
      values: [email],
    })
  }
})

it("should signin user", async () => {
  const email_or_username = "butcher@gmail.com"

  const body = {
    email_or_username,
    password: process.env.TEST_USER_PASSWORD,
  }
  const res = await supertest(app)
    .post(prefixPath + "/signin")
    .send(body)

  expect(res.body).toHaveProperty("jwt")
})
