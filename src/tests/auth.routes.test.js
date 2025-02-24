import request from "superwstest"
import { afterAll, beforeAll, describe, expect, it } from "@jest/globals"

import server from ".."
import { neo4jDriver } from "../configs/db.js"

beforeAll((done) => {
  server.listen(0, "localhost", done)
})

afterAll(async (done) => {
  await neo4jDriver.executeWrite(`MATCH (n) DETACH DELETE n`)

  server.close(done)
})

const signupPath = "/api/public/auth/signup"
const signinPath = "/api/public/auth/signin"
const signoutPath = "/api/private/app/signout"

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
        username: "mike",
        name: "Mike Ross",
        password: "blablabla",
        birthday: "2000-11-07",
        bio: "I'm a genius lawyer with no degree",
      })

    expect(res.status).toBe(201)
  })

  it("User1 signs out", async () => {
    const res = await request(server)
    .get(signoutPath)

    expect(res.status).toBe(200)
  })

  it("User1 signs in with incorrect credentials", async () => {
    const res = await request(server)
    .post(signinPath)
    .send({
      email_or_username: "mike",
      password: "blablabla",
    })

    expect(res.status).toBe(402)
  })
})

