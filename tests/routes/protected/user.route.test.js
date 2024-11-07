import fs from "fs/promises"
import { beforeAll, xit, it, expect } from "@jest/globals"
import dotenv from "dotenv"
import supertest from "supertest"

import app from "../../../src/app.js"

dotenv.config()

const prefixPath = "/api/user_private"

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

  await signUserIn("johnny@gmail.com")
  /* await signUserIn("butcher@gmail.com")
  await signUserIn("annak@gmail.com")
  await signUserIn("annie_star@gmail.com") */
})

it("should change user profile picture", async () => {
  const file = await fs.readFile("../../../profile_pic.png")
  
  const data = {
    picture_data: [...file],
  }

  const res = await supertest(app)
  .put(prefixPath + "/change_profile_picture")
  .set("Authorization", getJwt("johnny"))
  .send(data)

  expect(res.body).toHaveProperty("msg")
})

xit("should get the user session info via session jwt", async () => {
  const res = await supertest(app)
    .get(prefixPath + "/session_user")
    .set("Authorization", getJwt("johnny"))

  expect(res.body).toHaveProperty("sessionUser")
})

xit("should let client follow the user, and undo xit", async () => {
  const res1 = await supertest(app)
    .post(prefixPath + "/users/12/follow")
    .set("Authorization", getJwt("johnny"))

  expect(res1.body).toHaveProperty("msg")

  const res2 = await supertest(app)
    .delete(prefixPath + "/users/12/unfollow")
    .set("Authorization", getJwt("johnny"))

  expect(res2.body).toHaveProperty("msg")
})

xit("should edit client's profile", async () => {
  const data = { name: "Samuel Ayomide" }

  const res = await supertest(app)
    .patch(prefixPath + "/edit_profile")
    .set("Authorization", getJwt("johnny"))
    .send(data)

  expect(res.body).toHaveProperty("msg")
})

xit("should switch client's connection status between online and offline", async () => {
  const data1 = {
    connection_status: "online",
  }

  const res1 = await supertest(app)
    .patch(prefixPath + "/update_connection_status")
    .set("Authorization", getJwt("johnny"))
    .send(data1)

  expect(res1.body).toHaveProperty("msg")

  const data2 = {
    connection_status: "offline",
    last_active: new Date(),
  }

  const res2 = await supertest(app)
    .patch(prefixPath + "/update_connection_status")
    .set("Authorization", getJwt("johnny"))
    .send(data2)

  expect(res2.body).toHaveProperty("msg")
})

xit("should return client's home feed posts", async () => {
  const res = await supertest(app)
    .get(prefixPath + "/home_feed")
    .set("Authorization", getJwt("johnny"))

  expect(res.body).toBeInstanceOf(Array)
})

xit("should return posts client is mentioned in", async () => {
  const res = await supertest(app)
    .get(prefixPath + "/mentioned_posts")
    .set("Authorization", getJwt("johnny"))

  expect(res.body).toBeInstanceOf(Array)
})

xit("should return posts client reacted to", async () => {
  const res = await supertest(app)
    .get(prefixPath + "/reacted_posts")
    .set("Authorization", getJwt("johnny"))

  expect(res.body).toBeInstanceOf(Array)
})

xit("should return posts saved by client", async () => {
  const res = await supertest(app)
    .get(prefixPath + "/saved_posts")
    .set("Authorization", getJwt("johnny"))

  expect(res.body).toBeInstanceOf(Array)
})

xit("should return client's notifications", async () => {
  const res = await supertest(app)
    .get(prefixPath + "/my_notifications?from=2024-04-30")
    .set("Authorization", getJwt("johnny"))

  expect(res.body).toBeInstanceOf(Array)
})
