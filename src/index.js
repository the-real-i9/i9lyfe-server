import { createServer } from "http"
import jwt from "jsonwebtoken"

import app from "./app.js"

const server = createServer(app)

import { Server } from "socket.io"
import { NotificationService } from "./services/notification.service.js"
import { ChatRealtimeService } from "./services/realtime/chat.realtime.service.js"
import { PostCommentRealtimeService } from "./services/realtime/postComment.realtime.service.js"

import { renewJwtToken } from "./services/auth/auth.service.js"

const io = new Server(server)

io.use((socket, next) => {
  const token = socket.handshake.headers.authorization
  jwt.verify(token, process.env.JWT_SECRET, (err, decoded) => {
    if (err) return next(new Error(err.message))
    socket.jwt_payload = decoded
    next()
  })
})

io.on("connection", (socket) => {
  NotificationService.initRTC(io, socket)
  ChatRealtimeService.initRTC(io, socket)
  PostCommentRealtimeService.initRTC(io, socket)
  renewJwtToken(socket)
})

const PORT = process.env.PORT ?? 5000

server.listen(PORT, () => {
  console.log(`Server listening on port ${PORT}`)
})
