import { beforeAll, it, expect } from "@jest/globals"
import dotenv from "dotenv"
import supertest from "supertest"

import app from "../../../src/app.js"
import { dbQuery } from "../../../src/models/db.js"

dotenv.config()

const prefixPath = "/api/chat"

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
  // await signUserIn("annie_star@gmail.com")
})

it("should create conversation between client and partner", async () => {
  const data = {
    partner: {
      user_id: 11,
      username: "kendrick",
    },
    init_message: {
      type: "text",
      props: {
        textContent: "Hi! How're you?",
      },
    },
  }

  const res = await supertest(app)
    .post(prefixPath + "/create_conversation")
    .set("Authorization", getJwt("johnny"))
    .send(data)

  expect(res.body).toHaveProperty("conversation_id")

  //cleanup
  dbQuery({
    text: "DELETE FROM conversation WHERE id = $1",
    values: [res.body.conversation_id],
  })
})

it("should let client send a message", async () => {
  const data = {
    msg_content: {
      type: "image",
      props: {
        caption: "This is a dummy image!",
        mimeType: "image/jpg",
        size: 1024,
        data: [97, 98, 99, 100],
      },
    },
  }

  const res = await supertest(app)
    .post(prefixPath + `/conversations/1/partner/12/send_message`)
    .set("Authorization", getJwt("johnny"))
    .send(data)

  expect(res.body).toHaveProperty("new_msg_id")

  // clean up
  dbQuery({
    text: "DELETE FROM message_ WHERE id = $1",
    values: [res.body.new_msg_id],
  })
})

it("should return client's conversations", async () => {
  const res = await supertest(app)
    .get(prefixPath + "/my_conversations")
    .set("Authorization", getJwt("itz_butcher"))

  expect(res.body).toBeInstanceOf(Array)
})

it("should delete client's conversation", async () => {
  const res = await supertest(app)
    .delete(prefixPath + "/conversations/1")
    .set("Authorization", getJwt("johnny"))

  expect(res.body).toHaveProperty("msg")

  // cleanup
  await dbQuery({
    text: "UPDATE user_conversation SET deleted = false WHERE conversation_id = $1 AND user_id = $2",
    values: [1, 10],
  })
})

it("should return client's conversation's history", async () => {
  const res = await supertest(app)
    .get(prefixPath + "/conversations/1/history")
    .set("Authorization", getJwt("johnny"))

  expect(res.body).toBeInstanceOf(Array)
})

it("should let client acknowledge that the message has delivered", async () => {
  const data = {
    delivery_time: new Date(),
  }

  const res = await supertest(app)
    .put(prefixPath + "/conversations/1/partner/10/messages/4/delivered")
    .set("Authorization", getJwt("itz_butcher"))
    .send(data)

  expect(res.body).toHaveProperty("msg")
})

it("should let client acknowledge that they've read the message", async () => {
  const res = await supertest(app)
    .put(prefixPath + "/conversations/1/partner/12/messages/3/read")
    .set("Authorization", getJwt("johnny"))

  expect(res.body).toHaveProperty("msg")
})

it("should let client react to message, and then undo it", async () => {
  const data = {
    reaction: "🥰",
  }

  const res1 = await supertest(app)
    .post(prefixPath + "/conversations/1/partner/12/messages/6/react")
    .set("Authorization", getJwt("itz_butcher"))
    .send(data)

  expect(res1.body).toHaveProperty("msg")

  const res2 = await supertest(app)
    .delete(
      prefixPath + "/conversations/1/partner/12/messages/6/remove_reaction"
    )
    .set("Authorization", getJwt("itz_butcher"))

  expect(res2.body).toHaveProperty("msg")
})

it("should let client delete a message", async () => {
  const res = await supertest(app)
    .delete(prefixPath + "/conversations/1/partner/10/messages/3?delete_for=me")
    .set("Authorization", getJwt("itz_butcher"))

  expect(res.body).toHaveProperty("msg")
})
