import { beforeAll, it, expect } from "@jest/globals"
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
  await signUserIn("annak@gmail.com")
  await signUserIn("annie_star@gmail.com")
})

it("should create a post for client", async () => {
  const data = {
    media_blobs: [[97, 98, 99]],
    type: "reel",
    description: `Johnny! Johnny! Yes papa!`,
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

it("should let client react to post, and then undo it", async () => {
  const data = {
    reaction: "😂",
  }

  const res1 = await supertest(app)
    .post(prefixPath + "/users/13/posts/10/react")
    .set("Authorization", getJwt("kendrick"))
    .send(data)

  expect(res1.body).toHaveProperty("msg")

  const res2 = await supertest(app)
    .delete(prefixPath + "/posts/10/remove_reaction")
    .set("Authorization", getJwt("kendrick"))

  expect(res2.body).toHaveProperty("msg")
})

it("should return users who reacted to post", async () => {
  const res = await supertest(app)
    .get(prefixPath + "/posts/4/reactors")
    .set("Authorization", getJwt("johnny"))

  expect(res.body).toBeInstanceOf(Array)
})

it("should filter post reactors by specific reaction", async () => {
  const r = encodeURIComponent("🤣")

  const res = await supertest(app)
    .get(prefixPath + `/posts/4/reactors/${r}`)
    .set("Authorization", getJwt("johnny"))

  expect(res.body).toBeInstanceOf(Array)
})

it("should let client comment on the post", async () => {
  const data = {
    comment_text: `This is a comment on this post from @johnny.`,
    attachment_blob: [99],
  }

  const res = await supertest(app)
    .post(prefixPath + "/users/12/posts/4/comment")
    .set("Authorization", getJwt("johnny"))
    .send(data)

  expect(res.body).toHaveProperty("comment_id")

  // cleanup
  dbQuery({
    text: "DELETE FROM comment_ WHERE id = $1",
    values: [res.body.comment_id],
  })
})

it("should let client comment on (reply to) the comment", async () => {
  const data = {
    comment_text: `This is a reply to this comment from @johnny.`,
    attachment_blob: [99],
  }

  const res = await supertest(app)
    .post(prefixPath + "/users/13/posts/8/comment")
    .set("Authorization", getJwt("johnny"))
    .send(data)

  expect(res.body).toHaveProperty("comment_id")

  // cleanup
  dbQuery({
    text: "DELETE FROM comment_ WHERE id = $1",
    values: [res.body.comment_id],
  })
})

it("should return comments on the post", async () => {
  const res = await supertest(app)
    .get(prefixPath + "/posts/4/comments")
    .set("Authorization", getJwt("johnny"))

  expect(res.body).toBeInstanceOf(Array)
})

it("should return comments on (replies to) the comment", async () => {
  const res = await supertest(app)
    .get(prefixPath + "/comments/3/comments")
    .set("Authorization", getJwt("johnny"))

  expect(res.body).toBeInstanceOf(Array)
})

it("should return comment data in detail", async () => {
  const res = await supertest(app)
    .get(prefixPath + "/comments/3")
    .set("Authorization", getJwt("johnny"))

  expect(res.body).toHaveProperty("comment_id")
})

it("should delete client's comment on post", async () => {
  const res = await supertest(app)
    .delete(prefixPath + "/posts/4/comments/10")
    .set("Authorization", getJwt("johnny"))

  expect(res.body).toHaveProperty("msg")
})

it("should delete client's comment on (reply to) comment", async () => {
  const res = await supertest(app)
    .delete(prefixPath + "/comments/4/comments/25")
    .set("Authorization", getJwt("itz_butcher"))

  expect(res.body).toHaveProperty("msg")
})

it("should let client react to comment, and then undo it", async () => {
  const data = {
    reaction: "🦷",
  }

  const res1 = await supertest(app)
    .post(prefixPath + `/users/12/comments/4/react`)
    .set("Authorization", getJwt("kendrick"))
    .send(data)

  expect(res1.body).toHaveProperty("msg")

  const res2 = await supertest(app)
    .delete(prefixPath + "/comments/4/remove_reaction")
    .set("Authorization", getJwt("kendrick"))

  expect(res2.body).toHaveProperty("msg")
})

it("should return users who reacted to comment", async () => {
  const res = await supertest(app)
    .get(prefixPath + "/comments/3/reactors")
    .set("Authorization", getJwt("johnny"))

  expect(res.body).toBeInstanceOf(Array)
})

it("should filter comment reactors by specific reaction", async () => {
  const r = encodeURIComponent("🎯")

  const res = await supertest(app)
    .get(prefixPath + `/comments/6/reactors/${r}`)
    .set("Authorization", getJwt("johnny"))

  expect(res.body).toBeInstanceOf(Array)
})

it("should let client repost the post, and then undo it", async () => {
  const res1 = await supertest(app)
    .post(prefixPath + "/posts/4/repost")
    .set("Authorization", getJwt("starlight"))

  expect(res1.body).toHaveProperty("msg")

  const res2 = await supertest(app)
    .delete(prefixPath + "/posts/4/unrepost")
    .set("Authorization", getJwt("starlight"))

  expect(res2.body).toHaveProperty("msg")
})

it("should let client save the post, and then undo it", async () => {
  await dbQuery({
    text: "DELETE FROM saved_post WHERE saver_user_id = $1 AND post_id = $2",
    values: [11, 4],
  })

  const res1 = await supertest(app)
    .post(prefixPath + "/posts/4/save")
    .set("Authorization", getJwt("kendrick"))

  expect(res1.body).toHaveProperty("msg")

  const res2 = await supertest(app)
    .delete(prefixPath + "/posts/4/unsave")
    .set("Authorization", getJwt("kendrick"))

  expect(res2.body).toHaveProperty("msg")
})
