import { test, xtest, expect } from "@jest/globals"
import axios from "axios"
import dotenv from "dotenv"

dotenv.config()


const prefixPath = "http://localhost:5000/api/post_comment"
const i9xJwt =
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfdXNlcl9pZCI6MywiY2xpZW50X3VzZXJuYW1lIjoiaTl4IiwiaWF0IjoxNzEzOTA0OTUxfQ.f8DfuwetMyjWoipFQw54wkzIaMgrLCeRzTXKPFjQZdU"

const dollypJwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfdXNlcl9pZCI6NCwiY2xpZW50X3VzZXJuYW1lIjoiZG9sbHlwIiwiaWF0IjoxNzEzOTA2MzgwfQ.h_vTw7aXvC3uuTpRExJFqdc8xRkOAfZeC0IbgoXM7nA"

const axiosConfig = (authToken) => ({
  headers: {
    Authorization: `Bearer ${authToken}`,
  },
})

xtest("create post", async () => {
  const reqData = {
    media_blobs: [],
    type: "photo",
    description: "This is a post."
  }

  const res = await axios.post(prefixPath + "/new_post", reqData, axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("postData.post_id")
})

xtest("home feed", async () => {
  const res = await axios.get(prefixPath + "/home_feed?limit=20&offset=0", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("homeFeedPosts")
  expect(res.data.homeFeedPosts[0]).toHaveProperty("post_id")
})

xtest("post detail", async () => {
  const res = await axios.get(prefixPath + "/posts/14", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("post")
  expect(res.data.post).toHaveProperty("post_id")
})

xtest("delete post", async () => {
  const res = await axios.delete(prefixPath + "/posts/14", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
})

test("react to post", async () => {
  const reqData = {
    reaction: "🤣"
  }

  const res = await axios.post(prefixPath + "/users/3/posts/15/react", reqData, axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
})