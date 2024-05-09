import { test, xtest, expect } from "@jest/globals"
import axios from "axios"
import dotenv from "dotenv"

dotenv.config()

const prefixPath = "http://localhost:5000/api/app"

const i9xJwt =
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfdXNlcl9pZCI6MSwiY2xpZW50X3VzZXJuYW1lIjoiaTl4IiwiaWF0IjoxNzE1MTE5NTM3fQ.SgMAU2aK2A1FABBxOZDkJtTTiDGKSyhHb9516Fo0PsY"

const dollypJwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfdXNlcl9pZCI6MiwiY2xpZW50X3VzZXJuYW1lIjoiZG9sbHlwIiwiaWF0IjoxNzE1MTE5NjAzfQ.3UGpL3sDN5akB-zqpHfsq5qNJrY2snVxtRItESaADrc"

const axiosConfig = (authToken) => ({
  headers: {
    Authorization: `Bearer ${authToken}`,
  },
})

xtest("explore", async () => {
  const res = await axios.get(prefixPath + "/explore", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("explorePosts")

  console.log(res.data.explorePosts)
})

test("explore: search & filter", async () => {
  const res = await axios.get(prefixPath + "/explore/search?search=mention", axiosConfig(dollypJwt))
  // const res = await axios.get(prefixPath + "/explore/search?search=post&filter=video")
  // const res = await axios.get(prefixPath + "/explore/search?search=samuel&filter=user")
  // const res = await axios.get(prefixPath + "/explore/search?search=genius&filter=hashtag")

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("results")

  console.log(res.data.results)
})

xtest("get hashtag posts", async () => {
  const res = await axios.get(prefixPath + "/hashtags/genius", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("hashtagPosts")

  console.log(res.data.hashtagPosts)
})