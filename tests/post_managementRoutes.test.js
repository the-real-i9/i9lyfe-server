import { describe, it, expect } from "@jest/globals"
import request from "supertest"

import app from "../app.js"

describe("POST /create_post", () => {
  it("should create new post along with its mentions and hashtags", async () => {
    const res = await request(app)
      .post("/create_post")
      .send({
        user_id: 3,
        media_urls: ["https://localhost:5000/img/img_1.jpg"],
        type: "photo",
        description: "This is a text with #mommy you #daddy",
      })

    expect(res.status).toBe(200)
    expect(res.body.postData).toHaveProperty("id")
  })
})
