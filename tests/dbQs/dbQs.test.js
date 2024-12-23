import { expect, it, xit } from "@jest/globals"
// import dotenv from "dotenv"

import { User } from "../../src/graph_models/user.model.js"
import { Post } from "../../src/graph_models/post.model.js"

// dotenv.config()

xit("should update user connection status in neo4j", async () => {
  const res = await User.create({ email: "oluwarinolasam@gmail.com", username: "i9x", password: "blablabla", name: "Coder", birthday: new Date("2000-07-11").toISOString(), bio: "Testing my graph" })
  /* const res = await User.edit("27a554c8-6dca-41ef-85e4-73ce86b17d49", {
    password: "bulabalu",
    birthday: new Date("2000-07-11"),
  }) */
  // const res = await User.updateConnectionStatus({ client_user_id: "27a554c8-6dca-41ef-85e4-73ce86b17d49", connection_status: "online", last_active: null })
  expect(res).toHaveProperty("id")
})

it("should create post", async () => {
  const res = await Post.create({
    client_user_id: "82731e35-f535-4d7f-bc9e-979c57a65d34",
    client_username: "i9x",
    media_urls: ["https://images.com/photo-1.png"],
    type: "video",
    description: "This is a video where haghtag #testing and #programming",
    mentions: [],
    hashtags: ["testing", "programming"]
  })

  console.log(res)

  expect(res).toHaveProperty("new_post_data")
})
