import { createServer } from "http"

import dotenv from "dotenv"

dotenv.config()

import app from "./app.js"

import { Server as WSServer } from "socket.io"
import * as realtimeService from "./services/realtime.service.js"

import { renewJwtToken, verifyJwt } from "./services/security.services.js"
import { expressSession } from "./middlewares/auth.middlewares.js"
// import { neo4jDriver } from "./configs/db.js"
import { kafkaProducer } from "./configs/broker.js"

const httpServer = createServer(app)

const io = new WSServer(httpServer)

io.engine.use(expressSession())

io.engine.use((req, res, next) => {
  if (!req.session?.user) {
    return next(new Error("authentication required"))
  }

  const { authJwt } = req.session.user

  try {
    req.auth = verifyJwt(authJwt)
  } catch (error) {
    return next(error)
  }

  next()
})

io.use(async (socket, next) => {
  socket.auth = socket.request.auth

  realtimeService.initRTC(io)
  await realtimeService.initSocketRTC(socket)
  renewJwtToken(socket)

  next()
})

httpServer.on("close", () => {
  // neo4jDriver.close()
  kafkaProducer.close()
})

if (process.env.NODE_ENV !== "test") {
  const PORT = process.env.PORT ?? 5000

  httpServer.listen(PORT, () => {
    console.log(`Server listening on port ${PORT}`)
  })
}

export default httpServer
