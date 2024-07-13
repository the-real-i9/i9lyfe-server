import { it, xtest, expect } from "@jest/globals"
import fs from "fs"
import axios from "axios"
import dotenv from "dotenv"
import supertest from "supertest"

import app from "../../../src/app.js"
import os from "os"
import path from "path"

dotenv.config()

const prefixPath = "/api/user_private"

const userJwts = {}

function getJwt(username) {
  if (!userJwts[username]) {
    userJwts[username] =
      "Bearer " +
      fs.readFileSync(path.join(os.tmpdir(), "i9lyfe", `${username}.txt`), {
        encoding: "utf8",
      })
  }

  return userJwts[username]
}

it("should get session user", async () => {
  const res = await supertest(app)
    .get(prefixPath + "/session_user")
    .set("Authorization", getJwt("johnny"))

  expect(res.body).toHaveProperty("sessionUser")
})

it("should follow user", async () => {
  const res = await supertest(app)
    .post(prefixPath + "/users/12/follow")
    .set("Authorization", getJwt("johnny"))

    expect(res.body).not.toHaveProperty("error")
})

it("should unfollow user", async () => {
  const res = await supertest(app)
    .delete(prefixPath + "/users/12/unfollow")
    .set("Authorization", getJwt("johnny"))

    expect(res.body).not.toHaveProperty("error")
})

it("should edit profile", async () => {
  const data = { name: "Samuel Ayomide" }

  const res = await supertest(app)
    .patch(prefixPath + "/edit_profile")
    .set("Authorization", getJwt("johnny"))
    .send(data)

  expect(res.body).not.toHaveProperty("error")
})

it("update connection status", async () => {
  const data = {
    connection_status: "offline",
    last_active: new Date(),
  }

  const res = await supertest(app)
    .patch(prefixPath + "/update_connection_status")
    .set("Authorization", getJwt("johnny"))
    .send(data)

    expect(res.body).not.toHaveProperty("error")
})

xtest("get home feed posts", async () => {
  const res = await axios.get(prefixPath + "/home_feed", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toBeTruthy()
  console.log(res.data)
})

xtest("get posts mentioned in", async () => {
  const res = await axios.get(
    prefixPath + "/mentioned_posts",
    axiosConfig(dollypJwt)
  )

  expect(res.status).toBe(200)
  expect(res.data).toBeTruthy()
  console.log(res.data)
})

xtest("get posts reacted to", async () => {
  const res = await axios.get(
    prefixPath + "/reacted_posts",
    axiosConfig(dollypJwt)
  )

  expect(res.status).toBe(200)
  expect(res.data).toBeTruthy()
  console.log(res.data)
})

xtest("get posts saved", async () => {
  const res = await axios.get(
    prefixPath + "/saved_posts",
    axiosConfig(dollypJwt)
  )

  expect(res.status).toBe(200)
  expect(res.data).toBeTruthy()
  console.log(res.data)
})

xtest("get my notifications", async () => {
  const res = await axios.get(
    prefixPath + "/my_notifications?from=2024-04-30",
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(200)
  expect(res.data).toBeTruthy()
  console.log(res.data)
})
