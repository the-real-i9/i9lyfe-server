import express from "express"
import dotenv from "dotenv"
import cors from "cors"

import AuthRoutes from "./routes/public/auth.routes.js"
import PostCommentRoutes from "./routes/protected/postComment.routes.js"
import UserProtectedRoutes from "./routes/protected/user.protected.routes.js"
import ChatRoutes from "./routes/protected/chat.routes.js"
import UserPublicRoutes from "./routes/public/user.public.routes.js"
import AppRoutes from "./routes/public/app.routes.js"

dotenv.config()

const app = express()

app.use(cors())

app.use(express.json({ limit: "10mb" }))

app.use("/api/auth", AuthRoutes)

app.use("/api/post_comment", PostCommentRoutes)

app.use("/api/user_private", UserProtectedRoutes)
app.use("/api/user_public", UserPublicRoutes)

app.use("/api/chat", ChatRoutes)

app.use("/api/app", AppRoutes)

export default app
