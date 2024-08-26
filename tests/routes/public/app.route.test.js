import { it, expect, beforeAll, describe } from "@jest/globals"


import app from "../../../src/app.js"
import supertest from "supertest"

const prefixPath = "/api/app"

const userJwts = {}

function getJwt(username) {
  return "Bearer " + userJwts[username]
}

beforeAll(async () => {
  async function signUserIn(email_or_username) {
    const data = {
      email_or_username,
      password: process.env.TEST_USER_PASSWORD,
    }
    const res = await supertest(app).post("/api/auth/signin").send(data)

    expect(res.body).toHaveProperty("jwt")

    userJwts[res.body.user.username] = res.body.jwt
  }

  // await signUserIn("johnny@gmail.com")
  await signUserIn("butcher@gmail.com")
  // await signUserIn("annak@gmail.com")
  // await signUserIn("annie_star@gmail.com")
})

it("should search for users to chat with", async () => {
  const res = await supertest(app)
    .get(prefixPath + "/users/search?term=john")
    .set("Authorization", getJwt("itz_butcher"))

  expect(res.body).toBeInstanceOf(Array)
})

it("should return all explore content", async () => {
  const res = await supertest(app).get(prefixPath + "/explore")

  expect(res.body).toBeInstanceOf(Array)
})

describe("search and filter", () => {
  it('should search for app\'s posts(default) including the term: "cunt"', async () => {
    const res = await supertest(app).get(
      prefixPath + "/explore/search?term=cunt"
    )

    expect(res.body).toBeInstanceOf(Array)
    expect(res.body.length).toBeGreaterThan(0)
  })

  it('should search for app\'s posts(default) of type: video including the term: "mention"', async () => {
    const res = await supertest(app).get(
      prefixPath + "/explore/search?term=mention&filter=video"
    )

    expect(res.body).toBeInstanceOf(Array)
    expect(res.body.length).toBeGreaterThan(0)
  })
})

it('should return all posts with hashtag: "willy"', async () => {
  const res = await supertest(app).get(prefixPath + "/hashtags/willy")

  expect(res.body).toBeInstanceOf(Array)
})
