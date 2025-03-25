import fs from "node:fs/promises"
import request from "superwstest"
import { io } from "socket.io-client"
import { afterAll, beforeAll, describe, expect, test, xtest } from "@jest/globals"

import server from "../index.js"
import { neo4jDriver } from "../initializers/db.js"

const signupPath = "/api/auth/signup"
const appPathPriv = "/api/app/private"

/**
 * @typedef {Object} User
 * @property {string} email
 * @property {string} username
 * @property {string} name
 * @property {string} password
 * @property {string} [bio]
 * @property {string[]} [sessionCookie]
 * @property {import("socket.io-client").Socket} [cliSocket]
 */

/** @type {Object<string, User>} */
const users = {
  user1: {
    email: "harveyspecter@gmail.com",
    username: "harvey",
    name: "Harvey Specter",
    password: "harvey_psl",
    birthday: "1993-11-07",
    bio: "Whatever!",
  },
  user2: {
    email: "mikeross@gmail.com",
    username: "mikeross",
    name: "Mike Ross",
    password: "mikeross_psl",
    birthday: "1999-11-07",
    bio: "Whatever!",
  },
  user3: {
    email: "alexwilliams@gmail.com",
    username: "alex",
    name: "Alex Williams",
    password: "williams_psl",
    birthday: "1999-11-07",
    bio: "Whatever!",
  },
}

beforeAll(async () => {
  await neo4jDriver.executeWrite("MATCH (n) DETACH DELETE n")

  await new Promise((res) => {
    server.listen(5000, "localhost", res)
  })

  await Promise.all(
    Object.entries(users).map(async ([user, info]) => {
      {
        const res = await request(server)
          .post(`${signupPath}/request_new_account`)
          .send({ email: info.email })

        expect(res.status).toBe(200)

        users[user].sessionCookie = res.headers["set-cookie"]
      }

      {
        const verfCode = process.env.DUMMY_VERF_TOKEN

        const res = await request(server)
          .post(`${signupPath}/verify_email`)
          .set("Cookie", info.sessionCookie)
          .send({ code: verfCode })

        expect(res.status).toBe(200)

        users[user].sessionCookie = res.headers["set-cookie"]
      }

      {
        // eslint-disable-next-line no-unused-vars
        const { email, sessionCookie, ...restInfo } = info

        const res = await request(server)
          .post(`${signupPath}/register_user`)
          .set("Cookie", sessionCookie)
          .send(restInfo)

        expect(res.status).toBe(201)

        users[user].sessionCookie = res.headers["set-cookie"]
      }

      {
        const sock = io("ws://localhost:5000", {
          extraHeaders: { Cookie: info.sessionCookie },
        })

        expect(sock).toBeTruthy()

        users[user].cliSocket = sock
      }
    })
  )
})

afterAll((done) => {
  users.user1.cliSocket.close()
  users.user2.cliSocket.close()
  users.user3.cliSocket.close()

  server.close(done)
})

describe("test content sharing and interaction: a story between 3 users", () => {
  test("user1 creates post", async () => {
    const photo1 = await fs.readFile(
      new URL("./test_files/photo_1.png", import.meta.url)
    )
    expect(photo1).toBeTruthy()

    const res = await request(server)
      .post(`${appPathPriv}/new_post`)
      .set("Cookie", users.user1.sessionCookie)
      .send({
        media_data_list: [[...photo1]],
        type: "photo",
        description: "I'm beautiful",
      })

    expect(res.status).toBe(201)
    expect(res.body).toHaveProperty("owner_user.username", users.user1.username)
  })

  

  let user1Post2Id = ""

  test("user1 creates a post mentioning user2 | user2 is notified", async () => {
    const recvNotifProm = new Promise((resolve) => {
      users.user2.cliSocket.on("new notification", resolve)
    })

    const photo1 = await fs.readFile(
      new URL("./test_files/photo_1.png", import.meta.url)
    )
    expect(photo1).toBeTruthy()

    const res = await request(server)
      .post(`${appPathPriv}/new_post`)
      .set("Cookie", users.user1.sessionCookie)
      .send({
        media_data_list: [[...photo1]],
        type: "photo",
        description: `This is a post mentioning @${users.user2.username}`,
      })

    expect(res.status).toBe(201)
    expect(res.body).toHaveProperty("id")

    const recvNotif = await recvNotifProm

    expect(recvNotif).toBeTruthy()
    expect(recvNotif).toHaveProperty("id")
    expect(recvNotif).toHaveProperty("type", "mention_in_post")

    user1Post2Id = res.body.id
  })

  test("user2 views post2 in which she was mentioned by user1", async () => {
    const res = await request(server)
    .get(`${appPathPriv}/posts/${user1Post2Id}`)
    .set("Cookie", users.user2.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toHaveProperty("id", user1Post2Id)
  })

  let user1Post3Id = ""

  test("user1 creates a trending post | user2 and user3 receives the trending post", async () => {
    const recvPostProm2 = new Promise((resolve) => {
      users.user2.cliSocket.on("new post", resolve)
    })
    const recvPostProm3 = new Promise((resolve) => {
      users.user3.cliSocket.on("new post", resolve)
    })

    const photo1 = await fs.readFile(
      new URL("./test_files/photo_1.png", import.meta.url)
    )
    expect(photo1).toBeTruthy()

    const res = await request(server)
      .post(`${appPathPriv}/new_post`)
      .set("Cookie", users.user1.sessionCookie)
      .send({
        media_data_list: [[...photo1]],
        type: "photo",
        description: "This is No.1 #trending",
      })

    expect(res.status).toBe(201)
    expect(res.body).toHaveProperty("id")

    const recvPost2 = await recvPostProm2
    const recvPost3 = await recvPostProm3

    expect(recvPost2).toBeTruthy()
    expect(recvPost3).toBeTruthy()
    expect(recvPost2).toHaveProperty("owner_user.username", users.user1.username)
    expect(recvPost3).toHaveProperty("owner_user.username", users.user1.username)

    user1Post3Id = res.body.id
  })

  test("user2 reacts to post3 from user1 | user1 is notified", async () => {
    const recvNotifProm = new Promise((resolve) => {
      users.user1.cliSocket.on("new notification", resolve)
    })

    const res = await request(server)
    .post(`${appPathPriv}/posts/${user1Post3Id}/react`)
    .set("Cookie", users.user2.sessionCookie)
    .send({
      reaction: "🤔"
    })

    expect(res.status).toBe(201)
    expect(res.body).toHaveProperty("msg")

    const recvNotif = await recvNotifProm

    expect(recvNotif).toBeTruthy()
    expect(recvNotif).toHaveProperty("id")
    expect(recvNotif).toHaveProperty("type", "reaction_to_post")
    expect(recvNotif).toHaveProperty("reactor_user[1]", users.user2.username)
  })

  test("user3 reacts to post3 from user1 | user1 is notified", async () => {
    const recvNotifProm = new Promise((resolve) => {
      users.user1.cliSocket.on("new notification", resolve)
    })

    const res = await request(server)
    .post(`${appPathPriv}/posts/${user1Post3Id}/react`)
    .set("Cookie", users.user3.sessionCookie)
    .send({
      reaction: "🤔"
    })

    expect(res.status).toBe(201)
    expect(res.body).toHaveProperty("msg")

    const recvNotif = await recvNotifProm

    expect(recvNotif).toBeTruthy()
    expect(recvNotif).toHaveProperty("id")
    expect(recvNotif).toHaveProperty("type", "reaction_to_post")
    expect(recvNotif).toHaveProperty("reactor_user[1]", users.user3.username)
  })
})
