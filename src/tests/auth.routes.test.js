import request from "superwstest"
import { afterAll, beforeAll, describe, expect, it } from "@jest/globals"

import server from ".."
import { neo4jDriver } from "../configs/db.js"
import { registerUser } from "../services/auth/signup.service.js"

beforeAll((done) => {
  server.listen(0, "localhost", done)
})

afterAll(async (done) => {
  await neo4jDriver.executeWrite(`MATCH (n) DETACH DELETE n`)

  server.close(done)
})

const baseURL = "/api/public/auth"

describe("user signup", () => {
  let signupSessionCookie = ""

  it("should request new account", async () => {
    const res = await request(server)
      .post(`${baseURL}/signup/request_new_account`)
      .send({ email: "suberu@gmail.com" })

    expect(res.status).toBe(200)
    expect(res.body).toHaveProperty("msg")

    signupSessionCookie = res.headers["set-cookie"][0]
  })

  it("should verify email", async () => {
    const verfCode = Number(process.env.DUMMY_VERF_TOKEN)

    const res = await request(server)
      .post(`${baseURL}/signup/verify_email`)
      .set("Cookie", [signupSessionCookie])
      .send({ code: verfCode })

    expect(res.status).toBe(200)
    expect(res.body).toHaveProperty("msg")
  })

  it("should register user", async () => {
    const res = await request(server)
      .post(`${baseURL}/signup/register_user`)
      .set("Cookie", [signupSessionCookie])
      .send({
        username: "mike",
        name: "Mike Ross",
        password: "blablabla",
        birthday: "2000-11-07",
        bio: "I'm a genius lawyer with no degree",
      })

    expect(res.status).toBe(201)
    expect(res.body).toHaveProperty("jwt")
  })

  it("should sign in user", async () => {
    const pre_res = await registerUser({
      username: "mike",
      name: "Mike Ross",
      password: "blablabla",
      birthday: "2000-11-07",
      bio: "I'm a genius lawyer with no degree",
    })
  
    expect(pre_res).toHaveProperty("data.msg")

    const res = await request(server)
    .post(`${baseURL}/signin`)
    .send({
      email_or_username: "mike",
      password: "blablabla",
    })

    expect(res.status).toBe(200)
    expect(res.body).toHaveProperty("jwt")
  })
})

