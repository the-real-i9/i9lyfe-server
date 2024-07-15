import { beforeAll, it, expect } from "@jest/globals"
import dotenv from "dotenv"
import supertest from "supertest"

import app from "../../../src/app.js"

dotenv.config()

const prefixPath = "/api/user_public"

const userJwts = {}

function getJwt(username) {
  return "Bearer " + userJwts[username]
}

beforeAll(async () => {
  async function signUserIn(email_or_username) {
    const data = {
      email_or_username,
      password: "fhunmytor",
    }
    const res = await supertest(app).post("/api/auth/signin").send(data)

    expect(res.body).toHaveProperty("jwt")

    userJwts[res.body.user.username] = res.body.jwt
  }

  // await signUserIn("johnny@gmail.com")
  await signUserIn("butcher@gmail.com")
  await signUserIn("annak@gmail.com")
  await signUserIn("annie_star@gmail.com")
})

it("should return client's profile data", async () => {
  const res = await supertest(app).get(prefixPath + "/johnny")

  expect(res.body).toHaveProperty("user_id")
})

it("should return client's followers", async () => {
  const res = await supertest(app)
    .get(prefixPath + "/johnny/followers")
    .set("Authorization", getJwt("kendrick"))

  expect(res.body).toBeInstanceOf(Array)
})

it("should return user following client", async () => {
  const res = await supertest(app)
    .get(prefixPath + "/johnny/following")
    .set("Authorization", getJwt("starlight"))

  expect(res.body).toBeInstanceOf(Array)
})

it("should return posts published by client", async () => {
  const res = await supertest(app)
  .get(prefixPath + "/johnny/posts")
  .set("Authorization", getJwt("itz_butcher"))

  expect(res.body).toBeInstanceOf(Array)
})
