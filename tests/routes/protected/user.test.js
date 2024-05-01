import { test, xtest, expect } from "@jest/globals"
import axios from "axios"
import dotenv from "dotenv"

dotenv.config()

const prefixPath = "http://localhost:5000/api/user_private"
const i9xJwt =
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfdXNlcl9pZCI6MywiY2xpZW50X3VzZXJuYW1lIjoiaTl4IiwiaWF0IjoxNzEzOTA0OTUxfQ.f8DfuwetMyjWoipFQw54wkzIaMgrLCeRzTXKPFjQZdU"

const dollypJwt =
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfdXNlcl9pZCI6NCwiY2xpZW50X3VzZXJuYW1lIjoiZG9sbHlwIiwiaWF0IjoxNzEzOTA2MzgwfQ.h_vTw7aXvC3uuTpRExJFqdc8xRkOAfZeC0IbgoXM7nA"

const axiosConfig = (authToken) => ({
  headers: {
    Authorization: `Bearer ${authToken}`,
  },
})

xtest("get session user", async () => {
  const res = await axios.get(prefixPath + "/session_user", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("clientUser")
  console.log(res.data.clientUser)
})

xtest("follow user", async () => {
  const res = await axios.post(
    prefixPath + "/users/4/follow",
    null,
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(200)
})

xtest("unfollow user", async () => {
  const res = await axios.delete(
    prefixPath + "/users/4/unfollow",
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(200)
})

xtest("edit profile", async () => {
  const reqData = { name: "Samuel Ayomide" }

  const res = await axios.put(
    prefixPath + "/update_my_profile",
    reqData,
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("updatedUserData.name")
})

xtest("get posts mentioned in", async () => {
  const res = await axios.get(prefixPath + "/mentioned_posts", axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("mentionedPosts")
})

xtest("get posts reacted to", async () => {
  const res = await axios.get(prefixPath + "/reacted_posts", axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("reactedPosts")
})

test("get posts saved", async () => {
  const res = await axios.get(prefixPath + "/saved_posts", axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("savedPosts")
})
