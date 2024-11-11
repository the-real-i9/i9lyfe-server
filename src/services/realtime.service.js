import { EventEmitter } from "node:events"
import { Consumer, KafkaClient } from "kafka-node"
import { Post } from "../models/post.model"


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

  const kafkaClient = new KafkaClient({ kafkaHost: process.env.KAFKA_HOST })

  const consumer = new Consumer(kafkaClient, [
    { topic: `user-${client_user_id}-alerts` },
  ])

  consumer.on("message", (message) => {
    const { event, data } = JSON.parse(message.value.toString())

    socket.emit(event, data)
  })

  socket.on("disconnect", () => {
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

  newPostEventEmitter.on("new post", async (post_id) => {
    // get post based on "post recommendation algorithm"
    const post = await Post.find(post_id, client_user_id, true)

    if (post) {
      socket.send("new post", post)
    }
  })
}

export const sendPostUpdate = (post_id, data) => {
  sio.to(`post-${post_id}-updates`).emit("latest post update", data)
}

export const sendCommentUpdate = (comment_id, data) => {
  sio.to(`comment-${comment_id}-updates`).emit("latest comment update", data)
}
