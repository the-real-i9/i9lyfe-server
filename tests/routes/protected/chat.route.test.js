import { test, xtest, expect } from "@jest/globals"
import axios from "axios"
import dotenv from "dotenv"

dotenv.config()

const prefixPath = "http://localhost:5000/api/chat"
const i9xJwt =
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfdXNlcl9pZCI6MSwiY2xpZW50X3VzZXJuYW1lIjoiaTl4IiwiaWF0IjoxNzE1MTE5NTM3fQ.SgMAU2aK2A1FABBxOZDkJtTTiDGKSyhHb9516Fo0PsY"

const dollypJwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfdXNlcl9pZCI6MiwiY2xpZW50X3VzZXJuYW1lIjoiZG9sbHlwIiwiaWF0IjoxNzE1MTE5NjAzfQ.3UGpL3sDN5akB-zqpHfsq5qNJrY2snVxtRItESaADrc"

const axiosConfig = (authToken) => ({
  headers: {
    Authorization: `Bearer ${authToken}`,
  },
})

xtest("get users to chat with", async () => {
  const res = await axios.get(prefixPath + "/users_to_chat", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("users")
})

xtest("create dm conversation", async () => {
  const reqData = {
    partner: {
      user_id: 4,
      username: "dollyp"
    }
  }

  const res = await axios.post(prefixPath + "/create_dm_conversation", reqData, axiosConfig(i9xJwt))

  expect(res.status).toBe(201)
  expect(res.data).toHaveProperty("dm_conversation_id")
})

xtest("send message", async () => {
  const reqData = {
    msg_content: {
      type: "text",
      text_content: "Hi! How're you?"
    }
  }

  const res = await axios.post(prefixPath + "/conversations/2/send_message", reqData, axiosConfig(i9xJwt))

  expect(res.status).toBe(201)
})

xtest("get my conversations", async () => {
  const res = await axios.get(prefixPath + "/conversations", axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("conversations")
})

xtest("get conversation", async () => {
  const res = await axios.get(prefixPath + "/conversations/2", axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("conversation")
})

xtest("delete conversation", async () => {
  const res = await axios.delete(prefixPath + "/conversations/2", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
})

xtest("get conversation history", async () => {
  const res = await axios.get(prefixPath + "/conversations/2/history", axiosConfig(i9xJwt))

  expect(res.status).toBe(200)
  expect(res.data).toHaveProperty("conversationHistory")

  console.log(res.data)
})

xtest("acknowledge message delivered", async () => {
  const res = await axios.put(prefixPath + "/conversations/2/messages/2/delivered", null, axiosConfig(dollypJwt))

  expect(res.status).toBe(204)
})

xtest("acknowledge message read", async () => {
  const res = await axios.put(prefixPath + "/conversations/2/messages/2/read", null, axiosConfig(dollypJwt))

  expect(res.status).toBe(204)
})

xtest("react to message", async () => {
  const reqData = {
    reaction: "🥰"
  }

  const res = await axios.post(prefixPath + "/conversations/2/messages/2/react", reqData, axiosConfig(dollypJwt))

  expect(res.status).toBe(201)
})

xtest("remove reaction to message", async () => {
  const res = await axios.delete(prefixPath + "/conversations/2/messages/2/remove_reaction", axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
})

test("delete message", async () => {
  const res = await axios.delete(prefixPath + "/conversations/2/messages/2?delete_for=me", axiosConfig(dollypJwt))

  expect(res.status).toBe(200)
})
