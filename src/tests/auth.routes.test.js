import request from "superwstest"
import { afterAll, beforeAll, describe, expect, it } from "@jest/globals"

import server from ".."
import { neo4jDriver } from "../configs/graph_db.js"

beforeAll((done) => {
  server.listen(0, "localhost", done)
})

afterAll((done) => {
  server.close(done)
})

const baseURL = "/api/public/auth"

describe("user signup", () => {
  afterAll(async () => {
    await neo4jDriver.executeWrite(`MATCH (n) DETACH DELETE n`)
  })

  let cookie = ""
  let sess_id = "sess:"

  it("should request new account", async () => {
    const res = await request(server)
      .post(`${baseURL}/signup/request_new_account`)
      .send({ email: "suberu@gmail.com" })

    expect(res.status).toBe(200)
    expect(res.body).toHaveProperty("msg")

    cookie = res.headers["set-cookie"][0]
    sess_id += cookie.match(/(?<=connect\.sid=s%3A)[^.]+(?=\.)/)[0]
  })

  it("should verify email", async () => {
    const { records } = await neo4jDriver.executeRead(
      `MATCH (s:ongoing_signup{ sid: $sid }) RETURN s.data AS sess_data`,
      { sid: sess_id }
    )

    const sessData = JSON.parse(records[0].get("sess_data"))
    expect(sessData).toHaveProperty("signup")

    const verfCode = sessData.signup.data.verificationCode

    const res = await request(server)
      .post(`${baseURL}/signup/verify_email`)
      .set("Cookie", [cookie])
      .send({ code: verfCode })

    expect(res.status).toBe(200)
    expect(res.body).toHaveProperty("msg")
  })

  it("should register user", async () => {
    const res = await request(server)
      .post(`${baseURL}/signup/register_user`)
      .set("Cookie", [cookie])
      .send({
        username: "mike",
        name: "Mike Ross",
        password: "blablabla",
        birthday: "2000-11-07",
        bio: "I'm a genius lawyer with no degree",
      })

    expect(res.body).toHaveProperty("jwt")
  })
})
