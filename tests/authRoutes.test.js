import { describe, it, expect } from "@jest/globals"
import request from "supertest"
import app from "../app.js"

import { userAlreadyExists } from "../services/authServices.js"
;(await import("dotenv")).config()

describe("POST /auth/signup/request_new_account", () => {
  it("it should request new account or fail, if user with email already exists", async () => {
    try {
      const testEmail = "oluwarinolasam@gmail.com"

      const res = await request(app)
        .post("/auth/signup/request_new_account")
        .send({
          email: testEmail,
        })

      if (await userAlreadyExists(testEmail)) {
        expect(res.statusCode).toBe(422)
        console.log(res.body.reason)
      } else {
        expect(res.statusCode).toBe(200)
        console.log(res.body.msg)
      }
    } catch (error) {
      expect(error).toMatch("error")
    }
  })
})
