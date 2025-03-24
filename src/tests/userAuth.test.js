import request from "superwstest"
import { afterAll, beforeAll, describe, expect, test } from "@jest/globals"

import server from "../index.js"
import { neo4jDriver } from "../initializers/db.js"

beforeAll(async () => {
  await neo4jDriver.executeWrite("MATCH (n) DETACH DELETE n")

  await new Promise((res) => {
    server.listen(5000, "localhost", res)
  })

})

afterAll((done) => {
  server.close(done)
})

const signupPath = "/api/auth/signup"
const signinPath = "/api/auth/signin"
const forgotPasswordPath = "/api/auth/forgot_password"
const signoutPath = "/api/app/private/signout"

describe("test user authentication", () => {
  let sessionCookie = []

  test("User1 requests a new account", async () => {
    const res = await request(server)
      .post(`${signupPath}/request_new_account`)
      .send({ email: "suberu@gmail.com" })

    expect(res.status).toBe(200)

    sessionCookie = res.headers["set-cookie"]
  })

  test("User1 sends an incorrect email verf code", async () => {
    const verfCode = Number(process.env.DUMMY_VERF_TOKEN) + 1

    const res = await request(server)
      .post(`${signupPath}/verify_email`)
      .set("Cookie", sessionCookie)
      .send({ code: verfCode })

    expect(res.status).toBe(400)
    expect(res.body).toHaveProperty("msg")
  })

  test("User1 sends the correct email verf code", async () => {
    const verfCode = Number(process.env.DUMMY_VERF_TOKEN)

    const res = await request(server)
      .post(`${signupPath}/verify_email`)
      .set("Cookie", sessionCookie)
      .send({ code: verfCode })

    expect(res.status).toBe(200)

    sessionCookie = res.headers["set-cookie"]
  })

  test("User1 submits her credentials", async () => {
    const res = await request(server)
      .post(`${signupPath}/register_user`)
      .set("Cookie", sessionCookie)
      .send({
        username: "suberu",
        name: "Suberu Garuda",
        password: "sketeppy",
        birthday: "1993-11-07",
        bio: "Whatever!",
      })

    expect(res.status).toBe(201)

    sessionCookie = res.headers["set-cookie"]
  })

  test("User1 signs out", async () => {
    const res = await request(server)
      .get(signoutPath)
      .set("Cookie", sessionCookie)

    expect(res.status).toBe(200)
  })

  test("User1 signs in with incorrect credentials", async () => {
    const res = await request(server).post(signinPath).send({
      email_or_username: "suberu@gmail.com",
      password: "millini",
    })

    expect(res.status).toBe(404)
  })

  test("User1 signs in with correct credentials", async () => {
    const res = await request(server).post(signinPath).send({
      email_or_username: "suberu@gmail.com",
      password: "sketeppy",
    })

    expect(res.status).toBe(200)

    sessionCookie = res.headers["set-cookie"]
  })

  test("User1 signs out again", async () => {
    const res = await request(server)
      .get(signoutPath)
      .set("Cookie", sessionCookie)

    expect(res.status).toBe(200)
  })

  test("User1 requests password reset", async () => {
    const res = await request(server)
      .post(`${forgotPasswordPath}/request_password_reset`)
      .send({ email: "suberu@gmail.com" })

    expect(res.status).toBe(200)

    sessionCookie = res.headers["set-cookie"]
  })

  test("User1 sends an incorrect email confirmation token", async () => {
    const token = Number(process.env.DUMMY_VERF_TOKEN) + 1

    const res = await request(server)
      .post(`${forgotPasswordPath}/confirm_email`)
      .set("Cookie", sessionCookie)
      .send({ token })

    expect(res.status).toBe(400)
    expect(res.body).toHaveProperty("msg")
  })

  test("User1 sends the correct email confirmation token", async () => {
    const token = Number(process.env.DUMMY_VERF_TOKEN)

    const res = await request(server)
      .post(`${forgotPasswordPath}/confirm_email`)
      .set("Cookie", sessionCookie)
      .send({ token })

    expect(res.status).toBe(200)

    sessionCookie = res.headers["set-cookie"]
  })

  test("User1 changes her password", async () => {
    const res = await request(server)
      .post(`${forgotPasswordPath}/reset_password`)
      .set("Cookie", sessionCookie)
      .send({ newPassword: "millinie", confirmNewPassword: "millinie" })

    expect(res.status).toBe(200)
  })

  test("User1 signs in with new password", async () => {
    const res = await request(server).post(signinPath).send({
      email_or_username: "suberu",
      password: "millinie",
    })

    expect(res.status).toBe(200)
  })

  test("User2 requests a new account with already existing email", async () => {
    const res = await request(server)
      .post(`${signupPath}/request_new_account`)
      .send({ email: "suberu@gmail.com" })

    expect(res.status).toBe(400)
  })
})
