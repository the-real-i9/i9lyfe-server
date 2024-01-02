import express from "express"
import dotenv from "dotenv"
import cors from "cors"

import AuthRoutes from "./routes/private/AuthRoutes.js"
import PostCommentRoutes from "./routes/private/PostCommentRoutes.js"
import UserRoutes from "./routes/private/UserPrivateRoutes.js"

dotenv.config()

const app = express()

app.use(cors())

app.use(express.json())

app.use("/auth", AuthRoutes)
app.use(PostCommentRoutes)
app.use(UserRoutes)

export default app
