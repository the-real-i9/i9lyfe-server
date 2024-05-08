import { test, xtest, expect } from "@jest/globals"
import axios from "axios"
import dotenv from "dotenv"

dotenv.config()


const prefixPath = "http://localhost:5000/api/post_comment"
const i9xJwt =
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfdXNlcl9pZCI6MSwiY2xpZW50X3VzZXJuYW1lIjoiaTl4IiwiaWF0IjoxNzE1MTE5NTM3fQ.SgMAU2aK2A1FABBxOZDkJtTTiDGKSyhHb9516Fo0PsY"

const dollypJwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfdXNlcl9pZCI6MiwiY2xpZW50X3VzZXJuYW1lIjoiZG9sbHlwIiwiaWF0IjoxNzE1MTE5NjAzfQ.3UGpL3sDN5akB-zqpHfsq5qNJrY2snVxtRItESaADrc"

const axiosConfig = (authToken) => ({
  headers: {
    Authorization: `Bearer ${authToken}`,
  },
})

xtest("create post", async () => {
  const reqData = {
    media_blobs: [],
    type: "photo",
    description: "This is a post metioning @dollyp."
  }

  const res = await axios.post(prefixPath + "/new_post", reqData, axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("postData.post_id")
})

xtest("post detail", async () => {
  const res = await axios.get(prefixPath + "/posts/4", axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("post.post_id")

  console.log(res.data.post)
})

xtest("delete post", async () => {
  const res = await axios.delete(prefixPath + "/posts/4", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
})

xtest("react to post", async () => {
  const reqData = {
    reaction: "🤣"
  }

  const res = await axios.post(prefixPath + "/users/1/posts/4/react", reqData, axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
})

xtest("get users who reacted to post", async () => {
  const res = await axios.get(prefixPath + "/posts/4/reactors?limit=20&offset=0", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("postReactors")
})

xtest("get users with this post reaction", async () => {
  const res = await axios.get(prefixPath + "/posts/4/reactors/🤣?limit=20&offset=0", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("reactorsWithReaction")

  // console.log(res.data.reactorsWithReaction)
})

xtest("remove post reaction", async () => {
  const res = await axios.delete(prefixPath + "/posts/4/remove_reaction", axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
})

xtest("comment on post", async () => {
  const reqData = {
    comment_text: "This is another comment on this post from @i9x.",
    attachment_blob: null,
  }

  const res = await axios.post(prefixPath + "/users/1/posts/4/comment", reqData, axiosConfig(i9xJwt))

  expect(res.status).toBe(201)
  expect(res.data).toHaveProperty("commentData.comment_id")

  console.log(res.data.commentData)
})

xtest("comment on comment", async () => {
  const reqData = {
    comment_text: "Now what?",
    attachment_blob: null,
  }

  const res = await axios.post(prefixPath + "/users/2/comments/5/comment", reqData, axiosConfig(i9xJwt))

  expect(res.status).toBe(201)
  expect(res.data).toHaveProperty("commentData.comment_id")

  console.log(res.data.commentData)
})

xtest("get comments on post", async () => {
  const res = await axios.get(prefixPath + "/posts/4/comments", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("commentsOnPost")

  console.log(res.data.commentsOnPost)
})

xtest("get comments on comment", async () => {
  const res = await axios.get(prefixPath + "/comments/5/comments", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("commentsOnComment")

  console.log(res.data.commentsOnComment)
})

xtest("comment detail", async () => {
  const res = await axios.get(prefixPath + "/comments/5", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("comment.comment_id")

  console.log(res.data.comment)
})

xtest("delete comment on post", async () => {
  const res = await axios.delete(prefixPath + "/posts/4/comments/9", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
})

xtest("delete comment on comment", async () => {
  const res = await axios.delete(prefixPath + "/comments/5/comments/8", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
})

xtest("react to comment", async () => {
  const reqData = {
    reaction: "🎯"
  }

  const res = await axios.post(prefixPath + "/users/2/comments/5/react", reqData, axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
})

xtest("get users who reacted to comment", async () => {
  const res = await axios.get(prefixPath + "/comments/5/reactors", axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("commentReactors")

  console.log(res.data.commentReactors)
})

xtest("get users with this comment reaction", async () => {
  const res = await axios.get(prefixPath + "/comments/5/reactors/🎯?limit=20&offset=0", axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("commentReactorsWithReaction")

  console.log(res.data.commentReactorsWithReaction)
})

xtest("remove comment reaction", async () => {
  const res = await axios.delete(prefixPath + "/comments/5/remove_reaction", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
})

xtest("repost post", async () => {
  const res = await axios.post(prefixPath + "/posts/4/repost", null, axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
})

xtest("unrepost post", async () => {
  const res = await axios.delete(prefixPath + "/posts/4/unrepost", axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
})

xtest("save post", async () => {
  const res = await axios.post(prefixPath + "/posts/4/save", null, axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
})

xtest("unsave post", async () => {
  const res = await axios.delete(prefixPath + "/posts/4/unsave", axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
})