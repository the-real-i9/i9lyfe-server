import { it, xtest, expect } from "@jest/globals"
import axios from "axios"
import dotenv from "dotenv"
import supertest from "supertest"

import app from "../../../src/app.js"

dotenv.config()

const prefixPath = "/api/user_public"

it("should return user profile data", async () => {
  const res = await supertest(app).get(prefixPath + "/johnny")

  expect(res.body).toHaveProperty("user_id")
}, 5000)

xtest("should return user followers", async () => {
  const res = await axios.get(prefixPath + "/johnny/followers")

  expect(res.status).toBe(200)
  expect(res.data).toBeTruthy()

  console.log(res.data)
})

xtest("get user following", async () => {
  const res = await axios.get(prefixPath + "/i9x/following")

  expect(res.status).toBe(200)
  expect(res.data).toBeTruthy()

  console.log(res.data)
})

xtest("get user posts", async () => {
  const res = await axios.get(prefixPath + "/i9x/posts")

  expect(res.status).toBe(200)
  expect(res.data).toBeTruthy()

  console.log(res.data)
})
