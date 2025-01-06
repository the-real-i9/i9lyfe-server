import { EventEmitter } from "node:events"
import { consumeTopics } from "./messageBroker.service.js"
import { updateConnectionStatus } from "./user.service.js"
import { getPost } from "./contentRecommendation.service.js"

/** @type import("socket.io").Server */
let sio = null

/** @param {import("socket.io").Server} io */
export const initRTC = (io) => {
  sio = io
}

export const newPostEventEmitter = new EventEmitter()

/** @param {import("socket.io").Socket} socket */
export const initSocketRTC = (socket) => {
  const { client_user_id } = socket.jwt_payload

  updateConnectionStatus({ client_user_id, connection_status: "online" })

  const consumer = consumeTopics([
    { topic: `i9lyfe-user-${client_user_id}-alerts` },
  ])

  consumer.on("message", (message) => {
    const { event, data } = JSON.parse(message.value.toString())

    socket.emit(event, data)
  })

  socket.on("disconnect", () => {
    updateConnectionStatus({
      client_user_id,
      connection_status: "offline",
      last_active: new Date(),
    })
    consumer.close((err) => err && console.error(err))
  })

  consumer.on("error", (err) => console.error(err))

  consumer.on("offsetOutOfRange", (err) => console.error(err))

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

  newPostEventEmitter.on("new post", async (post_id, owner_user_id) => {
    if (owner_user_id === client_user_id) {
      return
    }
    
    // get post based on "post recommendation algorithm"
    const post = await getPost(post_id, client_user_id)

    if (post) {
      socket.emit("new post", post)
    }
  })
}

export const sendPostUpdate = (post_id, data) => {
  sio.to(`post-${post_id}-updates`).emit("latest post update", data)
}

export const sendCommentUpdate = (comment_id, data) => {
  sio.to(`comment-${comment_id}-updates`).emit("latest comment update", data)
}

export const publishNewPost = (post_id, owner_user_id) => {
  newPostEventEmitter.emit("new post", post_id, owner_user_id)
}
