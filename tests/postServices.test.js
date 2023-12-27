import dotenv from "dotenv"
import { test, expect } from "@jest/globals"
import { postCreationService, postReactionService } from "../services/postServices"

dotenv.config()

test("create new post", async () => {
  const res = await postCreationService({
    user_id: 3,
    media_urls: ["https://localhost:5000/img/img_1.jpg"],
    type: "photo",
    description: "This is a text with #ayo you #yemisi and @kenny with @samuel"
  })

  expect(res.ok).toBe(true)
})

test("react to post", async () => {
  const res = await postReactionService({
    user_id: 3,
    post_id: 31,
    reaction_code_point: "😴".codePointAt()
  })

  expect(res.ok).toBe(true)
})