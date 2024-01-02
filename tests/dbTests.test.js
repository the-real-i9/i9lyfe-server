import dotenv from "dotenv"
import { test, expect } from "@jest/globals"
import { getAllPostComments, getPost } from "../models/PostCommentModel"

dotenv.config()

test("get a single post", async () => {
  const res = await getPost({ post_id: 17, client_user_id: 4 })

  // console.log(res)
  expect(res).toBeTruthy()
})

test("get all the comments of a post", async () => {
  const res = await getAllPostComments({ post_id: 17, client_user_id: 4 })

  console.log(res)
  expect(res).toBeTruthy()
})
