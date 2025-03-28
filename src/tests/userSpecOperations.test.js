import fs from "node:fs/promises"
import request from "superwstest"
import { io } from "socket.io-client"
import { afterAll, beforeAll, describe, expect, test, xdescribe } from "@jest/globals"

import server from "../index.js"
import { neo4jDriver } from "../initializers/db.js"

const signupPath = "/api/auth/signup"
const appPathPriv = "/api/app/private"
const appPathPubl = "/api/app/public"

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

  // Create 3 users
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

xdescribe("test user specific operations", () => {
  test("edit user1 profile", async () => {
    // -- Edit user profile
  })

  test("change user1 profile picture", async () => {
    // -- Change user profile picture
  })

  test("view user2 profile", async () => {
    // -- View user profile
  })

  test("user1, user3, and user3 all follow each other", async () => {
    // -- Follow a user(s)
  })

  test("user2 checks user1 followers", async () => {})

  test("user2 checks user3 followings", async () => {})

  describe("user2 checks posts in which she's been mentioned | posts reacted to | posts saved", () => {
    let user1PostId = ""
    let user3PostId = ""

    beforeAll(() => {
      // user1 creates a post mentioning user2
      {
        //
      }
      // user3 creates post mentioning user2
      {
        //
      }
      // user2 reacts to user1 and user2 posts
      {
        //
      }
      // user2 saves user1 and user2 posts
      {
        //
      }
    })

    test("user2 checks posts in which she's been mentioned", async () => {})
    test("user2 checks posts reacted to", async () => {})
    test("user2 checks posts saved", async () => {})
  })

  test("user1 checks her notifications", async () => {})

  test("user2 checks her notifications", async () => {})

  test("user3 checks her notifications", async () => {})
})

// Story
// -- Check user followers and followings

// -- Create posts, mentioning users
// -- React to posts
// -- Save posts
// -- Check posts in which you were mentioned
// -- Check posts you saved
// -- Check posts you reacted to
// -- Check user notifications
// -- Read user notifications
