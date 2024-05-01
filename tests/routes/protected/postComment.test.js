import { test, xtest, expect } from "@jest/globals"
import axios from "axios"
import dotenv from "dotenv"

dotenv.config()


const prefixPath = "http://localhost:5000/api/post_comment"
const i9xJwtToken =
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfdXNlcl9pZCI6MywiY2xpZW50X3VzZXJuYW1lIjoiaTl4IiwiaWF0IjoxNzEzOTA0OTUxfQ.f8DfuwetMyjWoipFQw54wkzIaMgrLCeRzTXKPFjQZdU"

const axiosConfig = {
  headers: {
    Authorization: `Bearer ${i9xJwtToken}`,
  },
}

xtest("create post", async () => {
  const reqData = {
    media_blobs: [],
    type: "photo",
    description: "This is a post."
  }

  const res = await axios.post(prefixPath + "/new_post", reqData, axiosConfig)

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("postData.post_id")
})

xtest("home feed", async () => {
  const res = await axios.get(prefixPath + "/home_feed?limit=20&offset=0", axiosConfig)

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("homeFeedPosts")
  expect(res.data.homeFeedPosts[0]).toHaveProperty("post_id")
})

test("post detail", async () => {
  const res = await axios.get(prefixPath + "/posts/14", axiosConfig)

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("post")
  expect(res.data.post).toHaveProperty("post_id")
})

