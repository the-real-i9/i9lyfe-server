import dotenv from "dotenv"
import { test, expect } from "@jest/globals"
import { createNewPost } from "../models/postModel"

dotenv.config()

test("create new post", async () => {
  const res = await createNewPost({
    user_id: 3,
    media_urls: ["https://localhost:5000/img/img_1.jpg"],
    type: "photo",
    description: "This is a text with #ayo you #yemisi"
  })

  expect(res.rowCount).toBe(1)
})