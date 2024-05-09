import { test, xtest, expect } from "@jest/globals"
import axios from "axios"
import dotenv from "dotenv"

dotenv.config()

const prefixPath = "http://localhost:5000/api/user_public"

xtest("get user profile", async () => {
  const res = await axios.get(prefixPath + "/dollyp")

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("profileData.user_id")

  console.log(res.data.profileData)
})

xtest("get user followers", async () => {
  const res = await axios.get(prefixPath + "/dollyp/followers")

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("userFollowers")

  console.log(res.data)
})

xtest("get user following", async () => {
  const res = await axios.get(prefixPath + "/i9x/following")

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("userFollowing")

  console.log(res.data)
})

xtest("get user posts", async () => {
  const res = await axios.get(prefixPath + "/i9x/posts")

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("userPosts")

  console.log(res.data)
})