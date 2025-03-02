import { createServer } from "http"

import dotenv from "dotenv"

dotenv.config()

import app from "./app.js"

import { Server as WSServer } from "socket.io"
import * as realtimeService from "./services/realtime.service.js"

import { renewJwtToken, verifyJwt } from "./services/security.services.js"
import { expressSessionMiddleware } from "./middlewares/auth.middlewares.js"
import { neo4jDriver } from "./configs/db.js"
import { kafkaProducer } from "./configs/broker.js"

const httpServer = createServer(app)

const io = new WSServer(httpServer)

io.engine.use(
  expressSessionMiddleware("session_store", process.env.SESSION_COOKIE_SECRET)
)

io.engine.use((req, res, next) => {
  if (!req.session?.user) {
    return next(new Error("authentication required"))
  }

  const { authJwt } = req.session.user

  try {
    req.auth = verifyJwt(authJwt, process.env.AUTH_JWT_SECRET)
  } catch (error) {
    return next(error)
  }

  next()
})

io.use((socket, next) => {
  socket.auth = socket.request.auth

  next()
})

realtimeService.initRTC(io)

io.on("connection", (socket) => {
  realtimeService.initSocketRTC(socket)
  renewJwtToken(socket)
})

httpServer.on("close", () => {
  neo4jDriver.close()
  kafkaProducer.disconnect()
})

if (process.env.NODE_ENV !== "test") {
  const PORT = process.env.PORT ?? 5000

  httpServer.listen(PORT, () => {
    console.log(`Server listening on port ${PORT}`)
  })
}

export default httpServer
