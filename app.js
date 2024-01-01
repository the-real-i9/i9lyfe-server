import express from "express"
import dotenv from "dotenv"
import cors from "cors"

import authRoutes from "./routes/authRoutes.js"
import PostCommentRoutes from "./routes/PostCommentRoutes.js"
import UserRoutes from "./routes/UserRoutes.js"

dotenv.config()

const app = express()

app.use(cors())

app.use(express.json())

app.use("/auth", authRoutes)
app.use(PostCommentRoutes)
app.use(UserRoutes)

export default app
