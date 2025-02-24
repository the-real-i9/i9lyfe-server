import request from "superwstest"
import { afterAll, beforeAll, describe, expect, it } from "@jest/globals"

import server from ".."
import { neo4jDriver } from "../configs/db.js"

beforeAll(async () => {
  server.listen(0, "localhost")

  await neo4jDriver.executeWrite("MATCH (n) DETACH DELETE n")
})

afterAll((done) => {
  server.close(done)
})

const signupPath = "/api/auth/signup"
const signinPath = "/api/auth/signin"
const signoutPath = "/api/app/private/signout"

describe("test user authentication", () => {
  let signupSessionCookie = ""
  let userSessionCookie = ""

  it("User1 requests a new account", async () => {
    const res = await request(server)
      .post(`${signupPath}/request_new_account`)
      .send({ email: "suberu@gmail.com" })

    expect(res.status).toBe(200)

    signupSessionCookie = res.headers["set-cookie"][0]
  })

  it("User1 sends an incorrect email verf code", async () => {
    const verfCode = Number(process.env.DUMMY_VERF_TOKEN)+1

    const res = await request(server)
      .post(`${signupPath}/verify_email`)
      .set("Cookie", [signupSessionCookie])
      .send({ code: verfCode })

    expect(res.status).toBe(400)
    expect(res.body).toHaveProperty("msg")
  })

  it("User1 sends the correct email verf code", async () => {
    const verfCode = Number(process.env.DUMMY_VERF_TOKEN)

    const res = await request(server)
      .post(`${signupPath}/verify_email`)
      .set("Cookie", [signupSessionCookie])
      .send({ code: verfCode })

    expect(res.status).toBe(200)
  })

  it("User1 submits her credentials", async () => {
    const res = await request(server)
      .post(`${signupPath}/register_user`)
      .set("Cookie", [signupSessionCookie])
      .send({
        username: "suberu",
        name: "Suberu Garuda",
        password: "sketeppy",
        birthday: "1993-11-07",
        bio: "Whatever!",
      })

    expect(res.status).toBe(201)

    userSessionCookie = res.headers["set-cookie"][0]
  })

  it("User1 signs out", async () => {
    const res = await request(server)
    .get(signoutPath)
    .set("Cookie", [userSessionCookie])

    expect(res.status).toBe(200)
  })

  it("User1 signs in with incorrect credentials", async () => {
    const res = await request(server)
    .post(signinPath)
    .send({
      email_or_username: "suberu@gmail.com",
      password: "millini",
    })

    expect(res.status).toBe(404)
  })

  it("User1 signs in with correct credentials", async () => {
    const res = await request(server)
    .post(signinPath)
    .send({
      email_or_username: "suberu@gmail.com",
      password: "sketeppy",
    })

    expect(res.status).toBe(200)
  })

  it("User2 requests a new account with already existing email", async () => {
    const res = await request(server)
      .post(`${signupPath}/request_new_account`)
      .send({ email: "suberu@gmail.com" })

    expect(res.status).toBe(400)
  })
})

