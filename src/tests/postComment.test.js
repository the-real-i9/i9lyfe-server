import fs from "node:fs/promises"
import request from "superwstest"
import { io } from "socket.io-client"
import { afterAll, beforeAll, describe, expect, test } from "@jest/globals"

import server from "../index.js"
import { neo4jDriver } from "../configs/db.js"

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
}

beforeAll(async () => {
  server.listen(5000, "localhost")

  await neo4jDriver.executeWrite("MATCH (n) DETACH DELETE n")

  describe("signup two users and connect their RTC sockets", () => {
    Object.entries(users).forEach(([user, info], i) => {
      test(`user${i + 1} requests a new account`, async () => {
        const res = await request(server)
          .post(`${signupPath}/request_new_account`)
          .send({ email: info.email })

        expect(res.status).toBe(200)

        users[user].sessionCookie = res.headers["set-cookie"]
      })

      test(`user${i + 1} verifies email`, async () => {
        const verfCode = Number(process.env.DUMMY_VERF_TOKEN)

        const res = await request(server)
          .post(`${signupPath}/verify_email`)
          .set("Cookie", info.sessionCookie)
          .send({ code: verfCode })

        expect(res.status).toBe(200)

        users[user].sessionCookie = res.headers["set-cookie"]
      })

      test(`user${i + 1} submits her credentials`, async () => {
        // eslint-disable-next-line no-unused-vars
        const { email, sessionCookie, ...restInfo } = info

        const res = await request(server)
          .post(`${signupPath}/register_user`)
          .set("Cookie", sessionCookie)
          .send(restInfo)

        expect(res.status).toBe(201)

        users[user].sessionCookie = res.headers["set-cookie"]
      })

      test(`connect user${i + 1} socket`, async () => {
        const sock = io("ws://localhost:5000", {
          extraHeaders: { Cookie: info.sessionCookie },
        })

        expect(sock).toBeTruthy()

        users[user].cliSocket = sock
      })
    })
  })
})

afterAll((done) => {
  users.user1.cliSocket.close()
  users.user2.cliSocket.close()

  server.close(done)
})

describe("test posting and related functions", () => {
  /* Test every functionality associated with an endpoint before moving to the next */

  describe("test post creation", () => {
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
      expect(res.body).toHaveProperty(
        "owner_user.username",
        users.user1.username
      )
    })

    test("user1 creates a trending post received by user2", async () => {
      const recvPostProm = new Promise((resolve) => {
        users.user2.cliSocket.once("new post", resolve)
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

      const recvPost = await recvPostProm

      expect(recvPost).toBeTruthy()
      expect(recvPost).toHaveProperty(
        "owner_user.username",
        users.user1.username
      )
    })

    test("user1 creates a post mentioning user2", async () => {
      const recvNotifProm = new Promise((resolve) => {
        users.user2.cliSocket.once("new notification", resolve)
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

      const recvNotif = await recvNotifProm

      expect(recvNotif).toBeTruthy()
      expect(recvNotif).toHaveProperty("id")
      expect(recvNotif).toHaveProperty("type", "mention_in_post")
    })
  })
})
