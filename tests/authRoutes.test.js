import { describe, it, expect } from "@jest/globals"
import request from "supertest"
import app from "../app.js"

import { userAlreadyExists } from "../services/authServices.js"
;(await import("dotenv")).config()


describe("POST /auth/signup/request_new_account", () => {
  it("it should be successful (status 200), except user already exists (status 422)", async () => {
    try {
      const testEmail = "oluwarinolasam@gmail.com"

      const res = await request(app)
        .post("/auth/signup/request_new_account")
        .send({
          email: testEmail,
        })

      if (await userAlreadyExists(testEmail)) expect(res.statusCode).toBe(422)
      else expect(res.statusCode).toBe(200)
    } catch (error) {
      expect(error).toMatch("error")
    }
  })
})
