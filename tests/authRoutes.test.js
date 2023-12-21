import { describe, it, expect } from "@jest/globals"
import request from "supertest"
import app from "../app.js"
;(await import("dotenv")).config()

describe("POST /auth/signup/request_new_account", () => {
  it("it should request new account or fail, if user with email already exists", async () => {
    const testEmail = "oluwarinolasa@gmail.com"

    const res = await request(app)
      .post("/auth/signup/request_new_account")
      .send({
        email: testEmail,
      })

    expect(res.statusCode).toBe(200)
    expect(res.body).toHaveProperty("msg")
  })
})

describe("POST /auth/signin", () => {
  it("it should pass on correct credentials or fail otherwise", async () => {
    // const [testEmail, testPassword] = ["samuel123@gmail.com", "sammyken"] // incorrect credentials
    // const [testEmail, testPassword] = ["oluwarinolasam@gmail.com", "incfhunmytor"] // incorrect password credential
    const [testEmail, testPassword] = ["oluwarinolasam@gmail.com", "fhunmytor"] // correct credentials

    const res = await request(app).post("/auth/signin").send({
      email: testEmail,
      password: testPassword,
    })

    expect(res.statusCode).toBe(200)
    expect(res.body).toHaveProperty("jwtToken")
  })
})
