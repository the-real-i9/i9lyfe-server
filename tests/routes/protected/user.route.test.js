import { test, xtest, expect } from "@jest/globals"
import axios from "axios"
import dotenv from "dotenv"

dotenv.config()

const prefixPath = "http://localhost:5000/api/user_private"
const i9xJwt =
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfdXNlcl9pZCI6MSwiY2xpZW50X3VzZXJuYW1lIjoiaTl4IiwiaWF0IjoxNzE1MTE5NTM3fQ.SgMAU2aK2A1FABBxOZDkJtTTiDGKSyhHb9516Fo0PsY"

const dollypJwt =
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfdXNlcl9pZCI6MiwiY2xpZW50X3VzZXJuYW1lIjoiZG9sbHlwIiwiaWF0IjoxNzE1MTE5NjAzfQ.3UGpL3sDN5akB-zqpHfsq5qNJrY2snVxtRItESaADrc"

const axiosConfig = (authToken) => ({
  headers: {
    Authorization: `Bearer ${authToken}`,
  },
})

xtest("get session user", async () => {
  const res = await axios.get(prefixPath + "/session_user", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("sessionUser")
  console.log(res.data.sessionUser)
})

xtest("follow user", async () => {
  const res = await axios.post(
    prefixPath + "/users/2/follow",
    null,
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(200)
})

xtest("unfollow user", async () => {
  const res = await axios.delete(
    prefixPath + "/users/1/unfollow",
    axiosConfig(dollypJwt)
  )

  expect(res.status).toBe(200)
})

xtest("edit profile", async () => {
  const reqData = { name: "Samuel Ayomide" }

  const res = await axios.patch(
    prefixPath + "/edit_my_profile",
    reqData,
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(200)
})

xtest("update connection status", async () => {
  const reqData = {
    connection_status: "online",
    last_active: null,
  }

  const res = await axios.patch(
    prefixPath + "/update_my_connection_status",
    reqData,
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(200)
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

test("get my notifications", async () => {
  const res = await axios.get(
    prefixPath + "/my_notifications?from=2024-04-30",
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(200)
  expect(res.data).toBeTruthy()
  console.log(res.data)
})
