import fs from "node:fs/promises"
import request from "superwstest"
import { io } from "socket.io-client"
import { afterAll, beforeAll, describe, expect, test } from "@jest/globals"

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
  let user1Post1Id = ""

  test("user1 creates a trending post1 | user2 and user3 receives the trending post", async () => {
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

    const recvPostVal2 = await recvPostProm2
    const recvPostVal3 = await recvPostProm3

    expect(recvPostVal2).toBeTruthy()
    expect(recvPostVal3).toBeTruthy()
    expect(recvPostVal2).toHaveProperty(
      "owner_user.username",
      users.user1.username
    )
    expect(recvPostVal3).toHaveProperty(
      "owner_user.username",
      users.user1.username
    )

    user1Post1Id = res.body.id
  })

  test("user2 reacts to user1's post1 | user1 is notified", async () => {
    const recvNotifProm = new Promise((resolve) => {
      users.user1.cliSocket.on("new notification", resolve)
    })

    const res = await request(server)
      .post(`${appPathPriv}/posts/${user1Post1Id}/react`)
      .set("Cookie", users.user2.sessionCookie)
      .send({
        reaction: "🤔",
      })

    expect(res.status).toBe(201)
    expect(res.body).toHaveProperty("msg")

    const recvNotif = await recvNotifProm

    expect(recvNotif).toBeTruthy()
    expect(recvNotif).toHaveProperty("id")
    expect(recvNotif).toHaveProperty("type", "reaction_to_post")
    expect(recvNotif).toHaveProperty("reactor_user[1]", users.user2.username)
  })

  test("user3 reacts to user1's post1 | user1 is notified", async () => {
    const recvNotifProm = new Promise((resolve) => {
      users.user1.cliSocket.on("new notification", resolve)
    })

    const res = await request(server)
      .post(`${appPathPriv}/posts/${user1Post1Id}/react`)
      .set("Cookie", users.user3.sessionCookie)
      .send({
        reaction: "😀",
      })

    expect(res.status).toBe(201)
    expect(res.body).toHaveProperty("msg")

    const recvNotif = await recvNotifProm

    expect(recvNotif).toBeTruthy()
    expect(recvNotif).toHaveProperty("id")
    expect(recvNotif).toHaveProperty("type", "reaction_to_post")
    expect(recvNotif).toHaveProperty("reactor_user[1]", users.user3.username)
  })

  test("user1 checks reactors to her post1", async () => {
    const res = await request(server)
      .get(`${appPathPriv}/posts/${user1Post1Id}/reactors`)
      .set("Cookie", users.user1.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toBeInstanceOf(Array)
    expect(res.body).toHaveLength(2) // two users reacted

    for (const ri of res.body) {
      expect(ri).toHaveProperty("username")

      expect(
        [users.user2.username, users.user3.username].includes(ri.username)
      ).toBe(true)

      if (ri.username === users.user2.username) {
        expect(ri.reaction).toBe("🤔")
      }

      if (ri.username === users.user3.username) {
        expect(ri.reaction).toBe("😀")
      }
    }
  })

  test("user1 filters reactors to her post1 by a certain reaction", async () => {
    const rxn = encodeURIComponent("🤔")

    const res = await request(server)
      .get(`${appPathPriv}/posts/${user1Post1Id}/reactors/${rxn}`)
      .set("Cookie", users.user1.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toBeInstanceOf(Array)
    expect(res.body).toHaveLength(1)

    expect(res.body[0]).toHaveProperty("username", users.user2.username)
  })

  test("user3 removes her reaction from user1's post1", async () => {
    const res = await request(server)
      .delete(`${appPathPriv}/posts/${user1Post1Id}/remove_reaction`)
      .set("Cookie", users.user3.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toHaveProperty("msg")
  })

  test("user1 rechecks reactors to her post1 | user3's reaction gone", async () => {
    const res = await request(server)
      .get(`${appPathPriv}/posts/${user1Post1Id}/reactors`)
      .set("Cookie", users.user1.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toBeInstanceOf(Array)

    expect(res.body.some((v) => v.username === users.user3.username)).toBe(
      false
    )
  })

  let user2Comment1User1Post1Id = ""

  test("user2 comments on user1's post1 | user1 is notified", async () => {
    const recvNotifProm = new Promise((resolve) => {
      users.user1.cliSocket.on("new notification", resolve)
    })

    const res = await request(server)
      .post(`${appPathPriv}/posts/${user1Post1Id}/comment`)
      .set("Cookie", users.user2.sessionCookie)
      .send({
        comment_text: `This is a comment from ${users.user2.username}`,
      })

    expect(res.status).toBe(201)
    expect(res.body).toHaveProperty("id")

    const recvNotif = await recvNotifProm

    expect(recvNotif).toBeTruthy()
    expect(recvNotif).toHaveProperty("id")
    expect(recvNotif).toHaveProperty("type", "comment_on_post")
    expect(recvNotif).toHaveProperty("commenter_user[1]", users.user2.username)

    user2Comment1User1Post1Id = res.body.id
  })

  let user3Comment1User1Post1Id = ""

  test("user3 comments on user1's post1 | user1 is notified", async () => {
    const recvNotifProm = new Promise((resolve) => {
      users.user1.cliSocket.on("new notification", resolve)
    })

    const res = await request(server)
      .post(`${appPathPriv}/posts/${user1Post1Id}/comment`)
      .set("Cookie", users.user3.sessionCookie)
      .send({
        comment_text: `This is a comment from ${users.user3.username}`,
      })

    expect(res.status).toBe(201)
    expect(res.body).toHaveProperty("id")

    const recvNotif = await recvNotifProm

    expect(recvNotif).toBeTruthy()
    expect(recvNotif).toHaveProperty("id")
    expect(recvNotif).toHaveProperty("type", "comment_on_post")
    expect(recvNotif).toHaveProperty("commenter_user[1]", users.user3.username)

    user3Comment1User1Post1Id = res.body.id
  })

  test("user1 checks comments on her post1", async () => {
    const res = await request(server)
      .get(`${appPathPriv}/posts/${user1Post1Id}/comments`)
      .set("Cookie", users.user1.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toBeInstanceOf(Array)
    expect(res.body).toHaveLength(2)

    for (const ci of res.body) {
      expect(ci).toHaveProperty("owner_user.username")

      expect(
        [users.user2.username, users.user3.username].includes(
          ci.owner_user.username
        )
      ).toBe(true)

      if (ci.owner_user.username === users.user2.username) {
        expect(ci.comment_text).toBe(
          `This is a comment from ${users.user2.username}`
        )
      }

      if (ci.owner_user.username === users.user3.username) {
        expect(ci.comment_text).toBe(
          `This is a comment from ${users.user3.username}`
        )
      }
    }
  })

  test("user3 removes her comment on user1's post1", async () => {
    const res = await request(server)
      .delete(
        `${appPathPriv}/posts/${user1Post1Id}/comments/${user3Comment1User1Post1Id}`
      )
      .set("Cookie", users.user3.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toHaveProperty("msg")
  })

  test("user1 rechecks comments on her post1 | user3's comment is gone", async () => {
    const res = await request(server)
      .get(`${appPathPriv}/posts/${user1Post1Id}/comments`)
      .set("Cookie", users.user1.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toBeInstanceOf(Array)

    expect(res.body.some((v) => v.username === users.user3.username)).toBe(
      false
    )
  })

  test("user1 views user2's comment on her post1", async () => {
    const res = await request(server)
      .get(`${appPathPriv}/comments/${user2Comment1User1Post1Id}`)
      .set("Cookie", users.user1.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toHaveProperty("id", user2Comment1User1Post1Id)
  })

  let user1Reply1User2Comment1User1Post1Id = ""

  test("user1 replied to user2's comment on her post1 | user2 is notified", async () => {
    const recvNotifProm = new Promise((resolve) => {
      users.user2.cliSocket.on("new notification", resolve)
    })

    const res = await request(server)
      .post(`${appPathPriv}/comments/${user2Comment1User1Post1Id}/comment`)
      .set("Cookie", users.user1.sessionCookie)
      .send({
        comment_text: `This is a reply from ${users.user1.username}`,
      })

    expect(res.status).toBe(201)
    expect(res.body).toHaveProperty("id")

    const recvNotif = await recvNotifProm

    expect(recvNotif).toBeTruthy()
    expect(recvNotif).toHaveProperty("id")
    expect(recvNotif).toHaveProperty("type", "comment_on_comment")
    expect(recvNotif).toHaveProperty("commenter_user[1]", users.user1.username)

    user1Reply1User2Comment1User1Post1Id = res.body.id
  })

  let user3Reply1User2Comment1User1Post1Id = ""

  test("user3 replied to user2's comment on user1's post1 | user2 is notified", async () => {
    const recvNotifProm = new Promise((resolve) => {
      users.user2.cliSocket.on("new notification", resolve)
    })

    const res = await request(server)
      .post(`${appPathPriv}/comments/${user2Comment1User1Post1Id}/comment`)
      .set("Cookie", users.user3.sessionCookie)
      .send({
        comment_text: `I ${users.user3.username}, second ${users.user1.username} on this!`,
      })

    expect(res.status).toBe(201)
    expect(res.body).toHaveProperty("id")

    const recvNotif = await recvNotifProm

    expect(recvNotif).toBeTruthy()
    expect(recvNotif).toHaveProperty("id")
    expect(recvNotif).toHaveProperty("type", "comment_on_comment")
    expect(recvNotif).toHaveProperty("commenter_user[1]", users.user3.username)

    user3Reply1User2Comment1User1Post1Id = res.body.id
  })

  test("user2 checks replies to her comment1 on user1's post1", async () => {
    const res = await request(server)
      .get(`${appPathPriv}/comments/${user2Comment1User1Post1Id}/comments`)
      .set("Cookie", users.user2.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toBeInstanceOf(Array)
    expect(res.body).toHaveLength(2)

    for (const ci of res.body) {
      expect(ci).toHaveProperty("owner_user.username")

      expect(
        [users.user1.username, users.user3.username].includes(
          ci.owner_user.username
        )
      ).toBe(true)

      if (ci.owner_user.username === users.user1.username) {
        expect(ci.comment_text).toBe(
          `This is a reply from ${users.user1.username}`
        )
      }

      if (ci.owner_user.username === users.user3.username) {
        expect(ci.comment_text).toBe(
          `I ${users.user3.username}, second ${users.user1.username} on this!`
        )
      }
    }
  })

  test("user3 removes her reply to user2's comment1 on user1's post1", async () => {
    const res = await request(server)
      .delete(
        `${appPathPriv}/comments/${user2Comment1User1Post1Id}/comments/${user3Reply1User2Comment1User1Post1Id}`
      )
      .set("Cookie", users.user3.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toHaveProperty("msg")
  })

  test("user2 rechecks replies to her comment1 on user1's post1 | user3's reply is gone", async () => {
    const res = await request(server)
      .get(`${appPathPriv}/comments/${user2Comment1User1Post1Id}/comments`)
      .set("Cookie", users.user2.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toBeInstanceOf(Array)

    expect(res.body.some((v) => v.username === users.user3.username)).toBe(
      false
    )
  })

  test("user2 reacts to user1's reply to her comment1 on user1's post1 | user1 is notified", async () => {
    const recvNotifProm = new Promise((resolve) => {
      users.user1.cliSocket.on("new notification", resolve)
    })

    const res = await request(server)
      .post(
        `${appPathPriv}/comments/${user1Reply1User2Comment1User1Post1Id}/react`
      )
      .set("Cookie", users.user2.sessionCookie)
      .send({
        reaction: "😆",
      })

    expect(res.status).toBe(201)
    expect(res.body).toHaveProperty("msg")

    const recvNotif = await recvNotifProm

    expect(recvNotif).toBeTruthy()
    expect(recvNotif).toHaveProperty("id")
    expect(recvNotif).toHaveProperty("type", "reaction_to_comment")
    expect(recvNotif).toHaveProperty("reactor_user[1]", users.user2.username)
  })

  test("user3 reacts to user1's reply to user2's comment1 on user1's post1 | user1 is notified", async () => {
    const recvNotifProm = new Promise((resolve) => {
      users.user1.cliSocket.on("new notification", resolve)
    })

    const res = await request(server)
      .post(
        `${appPathPriv}/comments/${user1Reply1User2Comment1User1Post1Id}/react`
      )
      .set("Cookie", users.user3.sessionCookie)
      .send({
        reaction: "😂",
      })

    expect(res.status).toBe(201)
    expect(res.body).toHaveProperty("msg")

    const recvNotif = await recvNotifProm

    expect(recvNotif).toBeTruthy()
    expect(recvNotif).toHaveProperty("id")
    expect(recvNotif).toHaveProperty("type", "reaction_to_comment")
    expect(recvNotif).toHaveProperty("reactor_user[1]", users.user3.username)
  })

  test("user1 checks reactors to her reply to user2's comment1 on her post1", async () => {
    const res = await request(server)
      .get(
        `${appPathPriv}/comments/${user1Reply1User2Comment1User1Post1Id}/reactors`
      )
      .set("Cookie", users.user1.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toBeInstanceOf(Array)
    expect(res.body).toHaveLength(2) // two users reacted

    for (const ri of res.body) {
      expect(ri).toHaveProperty("username")

      expect(
        [users.user2.username, users.user3.username].includes(ri.username)
      ).toBe(true)

      if (ri.username === users.user2.username) {
        expect(ri.reaction).toBe("😆")
      }

      if (ri.username === users.user3.username) {
        expect(ri.reaction).toBe("😂")
      }
    }
  })

  test("user1 filters reactors to her reply to user2's comment1 on her post1 by a certain reaction", async () => {
    const rxn = encodeURIComponent("😆")

    const res = await request(server)
      .get(
        `${appPathPriv}/comments/${user1Reply1User2Comment1User1Post1Id}/reactors/${rxn}`
      )
      .set("Cookie", users.user1.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toBeInstanceOf(Array)
    expect(res.body).toHaveLength(1)

    expect(res.body[0]).toHaveProperty("username", users.user2.username)
  })

  test("user3 removes her reaction to user1's reply to user2's comment1 on user1's post1", async () => {
    const res = await request(server)
      .delete(
        `${appPathPriv}/comments/${user1Reply1User2Comment1User1Post1Id}/remove_reaction`
      )
      .set("Cookie", users.user3.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toHaveProperty("msg")
  })

  test("user1 rechecks reactors to her reply to user2's comment1 on her post1 | user3's reaction gone", async () => {
    const res = await request(server)
      .get(
        `${appPathPriv}/comments/${user1Reply1User2Comment1User1Post1Id}/reactors`
      )
      .set("Cookie", users.user1.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toBeInstanceOf(Array)

    expect(res.body.some((v) => v.username === users.user3.username)).toBe(
      false
    )
  })

  let user1Post2Id = ""

  test("user1 creates post2 mentioning user2 | user2 is notified", async () => {
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

  test("user2 views user1's post2 where she's been mentioned", async () => {
    const res = await request(server)
      .get(`${appPathPriv}/posts/${user1Post2Id}`)
      .set("Cookie", users.user2.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toHaveProperty("id", user1Post2Id)
  })

  let user1Post3Id = ""

  test("user1 creates post3", async () => {
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
    expect(res.body).toHaveProperty("id")
    expect(res.body).toHaveProperty("owner_user.username", users.user1.username)

    user1Post3Id = res.body.id
  })

  test("user2 reposts user1's post3 | user1 is notified", async () => {
    const recvNotifProm = new Promise((resolve) => {
      users.user1.cliSocket.on("new notification", resolve)
    })

    const res = await request(server)
      .post(`${appPathPriv}/posts/${user1Post3Id}/repost`)
      .set("Cookie", users.user2.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toHaveProperty("id")
    expect(res.body).toHaveProperty("origin_post_id", user1Post3Id)

    const recvNotif = await recvNotifProm

    expect(recvNotif).toBeTruthy()
    expect(recvNotif).toHaveProperty("id")
    expect(recvNotif).toHaveProperty("type", "repost")
  })

  test("user3 saves user1's post3", async () => {
    const res = await request(server)
      .post(`${appPathPriv}/posts/${user1Post3Id}/save`)
      .set("Cookie", users.user3.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toHaveProperty("msg")
  })

  test("user3 unsaves user1's post3", async () => {
    const res = await request(server)
      .delete(`${appPathPriv}/posts/${user1Post3Id}/unsave`)
      .set("Cookie", users.user3.sessionCookie)

    expect(res.status).toBe(200)
    expect(res.body).toHaveProperty("msg")
  })
})
