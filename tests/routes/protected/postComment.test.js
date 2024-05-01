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

test("create post", async () => {
  const reqData = {
    media_blobs: [],
    type: "photo",
    description: "This is a post metioning @dollyp."
  }

  const res = await axios.post(prefixPath + "/new_post", reqData, axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("postData.post_id")
})

xtest("home feed", async () => {
  const res = await axios.get(prefixPath + "/home_feed?limit=20&offset=0", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("homeFeedPosts")
})

xtest("post detail", async () => {
  const res = await axios.get(prefixPath + "/posts/15", axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("post.post_id")
})

xtest("delete post", async () => {
  const res = await axios.delete(prefixPath + "/posts/14", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
})

xtest("react to post", async () => {
  const reqData = {
    reaction: "🤣"
  }

  const res = await axios.post(prefixPath + "/users/3/posts/15/react", reqData, axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
})

xtest("get users who reacted to post", async () => {
  const res = await axios.get(prefixPath + "/posts/15/reactors?limit=20&offset=0", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("postReactors")
})

xtest("get users with this post reaction", async () => {
  const res = await axios.get(prefixPath + "/posts/15/reactors/🤣?limit=20&offset=0", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("reactorsWithReaction")
})

xtest("remove post reaction", async () => {
  const res = await axios.delete(prefixPath + "/posts/15/remove_reaction", axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
})

xtest("comment on post", async () => {
  const reqData = {
    comment_text: "This is another comment on this post.",
    attachment_blob: null,
  }

  const res = await axios.post(prefixPath + "/users/3/posts/15/comment", reqData, axiosConfig(dollypJwt))

  expect(res.status).toBe(201)
  expect(res.data).toHaveProperty("commentData.comment_id")
})

xtest("comment on comment", async () => {
  const reqData = {
    comment_text: "Now what?",
    attachment_blob: null,
  }

  const res = await axios.post(prefixPath + "/users/4/comments/6/comment", reqData, axiosConfig(i9xJwt))

  expect(res.status).toBe(201)
  expect(res.data).toHaveProperty("commentData.comment_id")
})

xtest("get comments on post", async () => {
  const res = await axios.get(prefixPath + "/posts/15/comments", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("commentsOnPost")
})

xtest("get comments on comment", async () => {
  const res = await axios.get(prefixPath + "/comments/5/comments", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("commentsOnComment")
})

xtest("comment detail", async () => {
  const res = await axios.get(prefixPath + "/comments/8", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("comment.comment_id")
})

xtest("delete comment on post", async () => {
  const res = await axios.delete(prefixPath + "/posts/15/comments/5", axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
})

xtest("delete comment on comment", async () => {
  const res = await axios.delete(prefixPath + "/comments/6/comments/10", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
})

xtest("react to comment", async () => {
  const reqData = {
    reaction: "🎯"
  }

  const res = await axios.post(prefixPath + "/users/4/comments/6/react", reqData, axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
})

xtest("get users who reacted to comment", async () => {
  const res = await axios.get(prefixPath + "/comments/6/reactors", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("commentReactors")
})

xtest("get users with this comment reaction", async () => {
  const res = await axios.get(prefixPath + "/comments/6/reactors/🎯?limit=20&offset=0", axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("commentReactorsWithReaction")
})

xtest("remove comment reaction", async () => {
  const res = await axios.delete(prefixPath + "/comments/6/remove_reaction", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
})

xtest("repost post", async () => {
  const res = await axios.post(prefixPath + "/posts/15/repost", null, axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
})

xtest("unrepost post", async () => {
  const res = await axios.delete(prefixPath + "/posts/15/unrepost", axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
})

xtest("save post", async () => {
  const res = await axios.post(prefixPath + "/posts/15/save", null, axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
})

xtest("unsave post", async () => {
  const res = await axios.delete(prefixPath + "/posts/15/unsave", axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
})