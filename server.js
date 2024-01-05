import { createServer } from 'http'
import jwt from 'jsonwebtoken'

import app from "./app.js"

const server = createServer(app)

import { Server } from 'socket.io'
import { NotificationService } from './services/NotificationService.js'

export const io = new Server(server)

io.use((socket, next) => {
  const token = socket.handshake.headers.authorization
  jwt.verify(token, process.env.JWT_SECRET, (err, decoded) => {
    if (err) return next(new Error(err.message));
    socket.jwt_payload = decoded;
    next();
  });
})

io.on("connection", (socket) => {
  NotificationService.initWebSocket(io, socket)
})

server.listen(5000, 'localhost', () => {
  console.log(`Server running at http://localhost:${process.env.PORT || 5000}`)
})
