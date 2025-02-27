import fs from "node:fs/promises"
import request from "superwstest"
import { afterAll, beforeAll, describe, expect, test } from "@jest/globals"

import server from "../index.js"
import { neo4jDriver } from "../configs/db.js"

beforeAll(async () => {
  server.listen(0, "localhost")

  await neo4jDriver.executeWrite("MATCH (n) DETACH DELETE n")
})

afterAll((done) => {
  server.close(done)
})

const signupPath = "/api/auth/signup"
const appPathPriv = "/api/app/private"

describe("test Post and associated activities", () => {
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

  describe("signup two users", () => {
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
    })
  })

  test("user1 creates post", async () => {
    const photo1 = await fs.readFile("./test_files/photo_1.png")
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
  })
})
