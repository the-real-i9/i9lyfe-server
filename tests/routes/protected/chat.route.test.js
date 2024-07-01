import { test, xtest, expect } from "@jest/globals"
import axios from "axios"
import dotenv from "dotenv"

dotenv.config()

const prefixPath = "http://localhost:5000/api/chat"
const i9xJwt =
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfdXNlcl9pZCI6MSwiY2xpZW50X3VzZXJuYW1lIjoiaTl4IiwiaWF0IjoxNzE1MTE5NTM3fQ.SgMAU2aK2A1FABBxOZDkJtTTiDGKSyhHb9516Fo0PsY"

const dollypJwt =
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfdXNlcl9pZCI6MiwiY2xpZW50X3VzZXJuYW1lIjoiZG9sbHlwIiwiaWF0IjoxNzE1MTE5NjAzfQ.3UGpL3sDN5akB-zqpHfsq5qNJrY2snVxtRItESaADrc"

const axiosConfig = (authToken) => ({
  headers: {
    Authorization: `Bearer ${authToken}`,
  },
})


xtest("create conversation", async () => {
  const reqData = {
    partner: {
      user_id: 2,
      username: "dollyp",
    },
    init_message: {
      type: "text",
      text_content: "Hi! How're you?",
    },
  }

  const res = await axios.post(
    prefixPath + "/create_conversation",
    reqData,
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(201)
  expect(res.data).toHaveProperty("conversation_id")

  console.log(res.data)
})

xtest("send message", async () => {
  const reqData = {
    msg_content: {
      type: "text",
      text_content: "Heeeyy! I'm fine!",
    },
  }

  const res = await axios.post(
    prefixPath + "/conversations/5/partner/1/send_message",
    reqData,
    axiosConfig(dollypJwt)
  )

  expect(res.status).toBe(201)
  expect(res.data).toHaveProperty("new_msg_id")

  console.log(res.data)
})

xtest("get my conversations", async () => {
  const res = await axios.get(
    prefixPath + "/my_conversations",
    axiosConfig(dollypJwt)
  )

  expect(res.status).toBe(200)
  expect(res.data).toBeTruthy()

  console.log(res.data)
})

xtest("delete conversation", async () => {
  const res = await axios.delete(
    prefixPath + "/conversations/2",
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(200)
})

xtest("get conversation history", async () => {
  const res = await axios.get(
    prefixPath + "/conversations/5/history",
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(200)
  expect(res.data).toBeTruthy()

  console.log(res.data)
})

xtest("acknowledge message delivered", async () => {
  const reqData = {
    delivery_time: new Date(),
  }

  const res = await axios.put(
    prefixPath + "/conversations/5/partner/1/messages/2/delivered",
    reqData,
    axiosConfig(dollypJwt)
  )

  expect(res.status).toBe(204)
})

xtest("acknowledge message read", async () => {
  const res = await axios.put(
    prefixPath + "/conversations/5/partner/1/messages/2/read",
    null,
    axiosConfig(dollypJwt)
  )

  expect(res.status).toBe(204)
})

xtest("react to message", async () => {
  const reqData = {
    reaction: "🥰",
  }

  const res = await axios.post(
    prefixPath + "/conversations/5/partner/2/messages/2/react",
    reqData,
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(201)
})

xtest("remove reaction to message", async () => {
  const res = await axios.delete(
    prefixPath + "/conversations/5/partner/2/messages/2/remove_reaction",
    axiosConfig(i9xJwt)
  )

  expect(res.status).toBe(200)
})

xtest("delete message", async () => {
  const res = await axios.delete(
    prefixPath + "/conversations/2/messages/2?delete_for=me",
    axiosConfig(dollypJwt)
  )

  expect(res.status).toBe(200)
})
