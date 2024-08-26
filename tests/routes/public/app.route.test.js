import { xtest, expect } from "@jest/globals"
import dotenv from "dotenv"

dotenv.config()

const prefixPath = "http://localhost:5000/api/app"


xtest("search users to chat with", async () => {
  // prefixPath + "/users/search?search=dolapo",
  
})

xtest("explore", async () => {
  // prefixPath + "/explore"
  
})

xtest("explore: search & filter", async () => {
  // prefixPath + "/explore/search?search=mention"
  
})

xtest("get hashtag posts", async () => {
  // prefixPath + "/hashtags/genius"
  
})