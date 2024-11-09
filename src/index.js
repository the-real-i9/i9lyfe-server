import { createServer } from "http"
import jwt from "jsonwebtoken"

import app from "./app.js"

const server = createServer(app)

import { Server } from "socket.io"
import * as realtimeService from "./services/realtime.service.js"

import { renewJwtToken } from "./services/auth.services.js"

const io = new Server(server)

io.use((socket, next) => {
  const token = socket.handshake.headers.authorization
  jwt.verify(token, process.env.JWT_SECRET, (err, decoded) => {
    if (err) return next(new Error(err.message))
    socket.jwt_payload = decoded
    next()
  })
})

realtimeService.initRTC(io)

io.on("connection", (socket) => {
  realtimeService.initSocketRTC(socket)
  renewJwtToken(socket)
})

const PORT = process.env.PORT ?? 5000

server.listen(PORT, () => {
  console.log(`Server listening on port ${PORT}`)
})
