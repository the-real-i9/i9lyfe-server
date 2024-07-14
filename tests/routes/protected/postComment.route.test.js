import { beforeAll, it, xtest, expect } from "@jest/globals"
import axios from "axios"
import dotenv from "dotenv"
import supertest from "supertest"

import app from "../../../src/app.js"
import { dbQuery } from "../../../src/models/db.js"

dotenv.config()

const prefixPath = "/api/post_comment"

const userJwts = {}

function getJwt(username) {
  return "Bearer " + userJwts[username]
}

beforeAll(async () => {
  async function signUserIn(email_or_username) {
    const data = {
      email_or_username,
      password: "fhunmytor",
    }
    const res = await supertest(app).post("/api/auth/signin").send(data)

    expect(res.body).toHaveProperty("jwt")

    userJwts[res.body.user.username] = res.body.jwt
  }

  await signUserIn("johnny@gmail.com")
  await signUserIn("butcher@gmail.com")
  /* await signUserIn("annak@gmail.com")
  await signUserIn("annie_star@gmail.com") */
})

it("should create a post for client", async () => {
  const data = {
    media_blobs: [[97, 98, 88]],
    type: "photo",
    description: `William Butcher likes to call people "cunt" `,
  }

  const res = await supertest(app)
    .post(prefixPath + "/new_post")
    .set("Authorization", getJwt("johnny"))
    .send(data)

  expect(res.body).toHaveProperty("post_id")

  // cleanup
  dbQuery({
    text: "DELETE FROM post WHERE id = $1",
    values: [res.body.post_id],
  })
})

it("should return the post data in detail", async () => {
  const res = await supertest(app)
    .get(prefixPath + "/posts/4")
    .set("Authorization", getJwt("itz_butcher"))

  expect(res.body).toHaveProperty("post_id")
})

it("should delete the client's post", async () => {
  const res = await supertest(app)
  .delete(prefixPath + "/posts/3")
  .set("Authorization", getJwt("johnny"))

  expect(res.body).toHaveProperty("msg")
})

xtest("react to post", async () => {
  const reqData = {
    reaction: "🤣",
  }

  const res = await axios.post(
    prefixPath + "/users/1/posts/4/react",
    reqData,
    axiosConfig(dollypJwt)
  )

  expect(res.status).toBe(200)
})

xtest("get users who reacted to post", async () => {
  const res = await axios.get(
    prefixPath + "/posts/4/reactors?limit=20&offset=0",
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(200)
  expect(res.data).toBeTruthy()
})

xtest("get users of certain post reaction", async () => {
  const res = await axios.get(
    prefixPath + "/posts/4/reactors/🤣?limit=20&offset=0",
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(200)
  expect(res.data).toBeTruthy()
})

xtest("remove post reaction", async () => {
  const res = await axios.delete(
    prefixPath + "/posts/4/remove_reaction",
    axiosConfig(dollypJwt)
  )

  expect(res.status).toBe(200)
})

xtest("comment on post", async () => {
  const reqData = {
    comment_text: "This is another comment on this post from @i9x.",
    attachment_blob: null,
  }

  const res = await axios.post(
    prefixPath + "/users/1/posts/4/comment",
    reqData,
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(201)
  expect(res.data).toHaveProperty("comment_id")

  console.log(res.data)
})

xtest("comment on comment", async () => {
  const reqData = {
    comment_text: "Now what?",
    attachment_blob: null,
  }

  const res = await axios.post(
    prefixPath + "/users/2/comments/5/comment",
    reqData,
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(201)
  expect(res.data).toHaveProperty("comment_id")

  console.log(res.data)
})

xtest("get comments on post", async () => {
  const res = await axios.get(
    prefixPath + "/posts/4/comments",
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(200)
  expect(res.data).toBeTruthy()

  console.log(res.data)
})

xtest("get comments on comment", async () => {
  const res = await axios.get(
    prefixPath + "/comments/5/comments",
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(200)
  expect(res.data).toBeTruthy()

  console.log(res.data)
})

xtest("comment detail", async () => {
  const res = await axios.get(prefixPath + "/comments/5", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("comment_id")

  console.log(res.data)
})

xtest("delete comment on post", async () => {
  const res = await axios.delete(
    prefixPath + "/posts/4/comments/9",
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(200)
})

xtest("delete comment on comment", async () => {
  const res = await axios.delete(
    prefixPath + "/comments/5/comments/8",
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(200)
})

xtest("react to comment", async () => {
  const reqData = {
    reaction: "🎯",
  }

  const res = await axios.post(
    prefixPath + "/users/2/comments/5/react",
    reqData,
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(200)
})

xtest("get users who reacted to comment", async () => {
  const res = await axios.get(
    prefixPath + "/comments/5/reactors",
    axiosConfig(dollypJwt)
  )

  expect(res.status).toBe(200)
  expect(res.data).toBeTruthy()

  console.log(res.data)
})

xtest("get users with this comment reaction", async () => {
  const res = await axios.get(
    prefixPath + "/comments/5/reactors/🎯?limit=20&offset=0",
    axiosConfig(dollypJwt)
  )

  expect(res.status).toBe(200)
  expect(res.data).toBeTruthy()

  console.log(res.data)
})

xtest("remove comment reaction", async () => {
  const res = await axios.delete(
    prefixPath + "/comments/5/remove_reaction",
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(200)
})

xtest("repost post", async () => {
  const res = await axios.post(
    prefixPath + "/posts/4/repost",
    null,
    axiosConfig(dollypJwt)
  )

  expect(res.status).toBe(200)
})

xtest("unrepost post", async () => {
  const res = await axios.delete(
    prefixPath + "/posts/4/unrepost",
    axiosConfig(dollypJwt)
  )

  expect(res.status).toBe(200)
})

xtest("save post", async () => {
  const res = await axios.post(
    prefixPath + "/posts/4/save",
    null,
    axiosConfig(dollypJwt)
  )

  expect(res.status).toBe(200)
})

xtest("unsave post", async () => {
  const res = await axios.delete(
    prefixPath + "/posts/4/unsave",
    axiosConfig(dollypJwt)
  )

  expect(res.status).toBe(200)
})
