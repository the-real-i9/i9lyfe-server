import dotenv from "dotenv"
import { test, expect } from "@jest/globals"
import { PostService } from "../services/PostService.js"
import { Post, PostCommentService } from "../services/PostCommentService.js"

dotenv.config()

test("create new post", async () => {
  try {
    const data = await new PostService().createPost({
      client_user_id: 4,
      media_urls: ["https://localhost:5000/img/img_1.jpg"],
      type: "photo",
      description:
        "This is a text with #ayo you #yemisi and @gen_i9 with @mckenney",
    })

    console.log(data)
  } catch (error) {
    console.error(error)
    expect(error).toBeUndefined()
  }
})

test.skip("react to post", async () => {
  try {
    const data = await new PostCommentService(new Post(8, 4)).addReaction({
      reactor_user_id: 4,
      reaction_code_point: "👨‍👩‍👧".codePointAt(),
    })

    console.log(data)
  } catch (error) {
    console.error(error)
    expect(error).toBeUndefined()
  }
})
