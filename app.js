import express from "express"
import dotenv from "dotenv"
import cors from "cors"

import AuthRoutes from "./routes/public/AuthRoutes.js"
import PostCommentRoutes from "./routes/protected/PostCommentRoutes.js"
import UserPrivateRoutes from "./routes/protected/UserPrivateRoutes.js"
import ChatRoutes from "./routes/protected/ChatRoutes.js"
import UserPublicRoutes from "./routes/public/UserPublicRoutes.js"
import AppRoutes from "./routes/public/AppRoutes.js"


dotenv.config()

const app = express()

app.use(cors({
  origin: "http://localhost:5173",
  credentials: true,
}))

app.use(express.json())

app.use("/api/auth", AuthRoutes)

app.use("/api/post_comment", PostCommentRoutes)

app.use("/api/user_private", UserPrivateRoutes)
app.use("/api/user_public", UserPublicRoutes)

app.use("/api/chat", ChatRoutes)

app.use("/api/app", AppRoutes)

export default app
