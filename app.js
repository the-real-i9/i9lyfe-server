import express from "express"
import dotenv from "dotenv"
import cors from "cors"
import path, { dirname } from "path"
import { fileURLToPath } from 'url';

import authRoutes from "./routes/public/auth.routes.js"
import PostCommentRoutes from "./routes/protected/PostCommentRoutes.js"
import UserProtectedRoutes from "./routes/protected/UserProtectedRoutes.js"
import ChatRoutes from "./routes/protected/ChatRoutes.js"
import UserPublicRoutes from "./routes/public/UserPublicRoutes.js"
import AppRoutes from "./routes/public/AppRoutes.js"


dotenv.config()

const app = express()

app.use(cors({
  origin: "http://localhost:5173",
  credentials: true,
}))

const __dirname = dirname(fileURLToPath(import.meta.url));
app.use(express.static(path.join(__dirname, "static")))

app.use(express.json({ limit: "10mb" }))


app.use("/api/auth", authRoutes)

app.use("/api/post_comment", PostCommentRoutes)

app.use("/api/user_private", UserProtectedRoutes)
app.use("/api/user_public", UserPublicRoutes)

app.use("/api/chat", ChatRoutes)

app.use("/api/app", AppRoutes)

export default app
