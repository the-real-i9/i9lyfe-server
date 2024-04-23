import { test, xtest, expect } from "@jest/globals"
import axios from "axios"
import dotenv from "dotenv"

dotenv.config()

const prefixPath = "http://localhost:5000/api/user_private"
const i9xJwtToken =
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfdXNlcl9pZCI6MywiY2xpZW50X3VzZXJuYW1lIjoiaTl4IiwiaWF0IjoxNzEzOTA0OTUxfQ.f8DfuwetMyjWoipFQw54wkzIaMgrLCeRzTXKPFjQZdU"

const axiosConfig = {
  headers: {
    Authorization: `Bearer ${i9xJwtToken}`,
  },
}

xtest("get session user", async () => {
  const res = await axios.get(prefixPath + "/session_user", axiosConfig)

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("clientUser")
  console.log(res.data.clientUser)
})

xtest("follow user", async () => {
  const res = await axios.post(
    prefixPath + "/follow_user",
    { to_follow_user_id: 4 },
    axiosConfig
  )

  expect(res.status).toBe(200)
})

xtest("unfollow user", async () => {
  const res = await axios.delete(prefixPath + "/followings/4", axiosConfig)

  expect(res.status).toBe(200)
})

test("edit profile", async () => {
  const res = await axios.put(
    prefixPath + "/update_my_profile",
    { name: "Samuel Ayomide" },
    axiosConfig
  )

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("updatedUserData.name")
})


// xtest()
