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
        description: "This is a text with @gen_i9",
      })

    expect(res.status).toBe(200)
    expect(res.body.postData).toHaveProperty("id")
  })
})

/* describe("POST /react_to_post", () => {
  it("should react to post", async () => {
    const res = await request(app).post("/react_to_post").send({
      reactor_user_id: 3,
      post_id: 10,
      post_owner_user_id: 3,
      reaction_code_point: "😍".codePointAt(),
    })

    expect(res.status).toBe(200)
  })
}) */

/* describe("POST /comment_on_post", () => {
  it("should comment on post", async () => {
    const res = await request(app).post("/comment_on_post").send({
      post_id: 10,
      post_owner_user_id: 3,
      commenter_user_id: 3,
      comment_text: "This is another comment from a comment from i9_gen",
      attachment_url: "https://giphy.com/laughing_baby",
    })

    expect(res.status).toBe(200)
    expect(res.body.commentData).toHaveProperty("id")
  })
}) */

/* describe("POST /react_to_comment", () => {
  it("should react to comment", async () => {
    const res = await request(app).post("/react_to_comment").send({
      reactor_user_id: 3,
      comment_id: 5,
      comment_owner_user_id: 3,
      reaction_code_point: "😁".codePointAt(),
    })

    expect(res.status).toBe(200)
  })
}) */

/* describe("POST /reply_to_comment", () => {
  it("should reply to comment", async () => {
    const res = await request(app).post("/reply_to_comment").send({
      comment_id: 4,
      comment_owner_user_id: 3,
      replier_user_id: 3,
      reply_text: "This is a reply from Kenny boyy!",
      attachment_url: "https://giphy.com/laughing_baby",
    })

    expect(res.status).toBe(200)
    expect(res.body.replyData).toHaveProperty("id")
  })
}) */