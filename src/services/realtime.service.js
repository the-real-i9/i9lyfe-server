import { EventEmitter } from "node:events"
import { consumeTopics } from "./messageBroker.service.js"
import { updateConnectionStatus } from "./user.service.js"
import { getPost } from "./contentRecommendation.service.js"
import * as chatService from "./chat.service.js"

/** @type import("socket.io").Server */
let sio = null

/** @param {import("socket.io").Server} io */
export const initRTC = (io) => {
  sio = io
}

export const newPostEventEmitter = new EventEmitter()

/** @param {import("socket.io").Socket} socket */
export const initSocketRTC = async (socket) => {
  const { client_username } = socket.auth

  updateConnectionStatus({ client_username, connection_status: "online" })

  // CONSUME EVENTS IN TOPICS
  const consumer = await consumeTopics([
    { topic: `i9lyfe-user-${client_username}-alerts` },
  ])

  consumer.on("message", (message) => {
    const { event, data } = JSON.parse(message.value.toString())

    socket.emit(event, data)
  })

  socket.on("disconnect", async () => {
    updateConnectionStatus({
      client_username,
      connection_status: "offline",
      last_active: new Date(),
    })

    consumer.close((err) => err && console.error(err))
  })

  consumer.on("error", (err) => err && console.error(err))

  consumer.on("offsetOutOfRange", (err) => err && console.error(err))

  // REALTIME POST AND COMMENT UPDATES
  socket.on("start receiving post updates", (post_id) => {
    socket.join(`post-${post_id}-updates`)
  })

  socket.on("stop receiving post updates", (post_id) => {
    socket.leave(`post-${post_id}-updates`)
  })

  socket.on("start receiving comment updates", (comment_id) => {
    socket.join(`comment-${comment_id}-updates`)
  })

  socket.on("stop receiving comment updates", (comment_id) => {
    socket.leave(`comment-${comment_id}-updates`)
  })

  // NEW POST PUBLISHING
  newPostEventEmitter.on("new post", async (post_id, owner_username) => {
    if (owner_username === client_username) {
      return
    }

    // get post based on "post recommendation algorithm"
    const post = await getPost(post_id, client_username)

    if (post) {
      socket.emit("new post", post)
    }
  })

  // CLIENT USER ACTIONS
  socket.on("send message", async (data) => {
    try {
      const resp = await chatService.sendMessage({ client_username, ...data })

      socket.emit("server reply", { toEvent: "send message", reply: resp.data })
    } catch (error) {
      socket.emit("server error", { onEvent: "send message", error })
    }
  })

  socket.on("get chat history", async (data) => {
    try {
      const resp = await chatService.getChatHistory({
        client_username,
        ...data,
      })

      socket.emit("server reply", {
        toEvent: "get chat history",
        reply: resp.data,
      })
    } catch (error) {
      socket.emit("server error", { onEvent: "get chat history", error })
    }
  })

  socket.on("ack message delivered", async (data) => {
    try {
      await chatService.ackMessageDelivered({ client_username, ...data })
    } catch (error) {
      socket.emit("server error", { onEvent: "ack message delivered", error })
    }
  })

  socket.on("ack message read", async (data) => {
    try {
      await chatService.ackMessageRead({ client_username, ...data })
    } catch (error) {
      socket.emit("server error", { onEvent: "ack message read", error })
    }
  })

  socket.on("react to message", async (data) => {
    try {
      await chatService.reactToMessage({ client_username, ...data })
    } catch (error) {
      socket.emit("server error", { onEvent: "react to message", error })
    }
  })

  socket.on("remove reaction to message", async (data) => {
    try {
      await chatService.removeReactionToMessage({ client_username, ...data })
    } catch (error) {
      socket.emit("server error", {
        onEvent: "remove reaction to message",
        error,
      })
    }
  })

  socket.on("delete message", async (data) => {
    try {
      await chatService.deleteMessage({ client_username, ...data })
    } catch (error) {
      socket.emit("server error", { onEvent: "delete message", error })
    }
  })
}

export const sendPostUpdate = (post_id, data) => {
  sio.to(`post-${post_id}-updates`).emit("latest post update", data)
}

export const sendCommentUpdate = (comment_id, data) => {
  sio.to(`comment-${comment_id}-updates`).emit("latest comment update", data)
}

export const publishNewPost = (post_id, owner_username) => {
  newPostEventEmitter.emit("new post", post_id, owner_username)
}
