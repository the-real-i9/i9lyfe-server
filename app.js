import express from "express"
import dotenv from "dotenv"
// import cors from "cors"

import AuthRoutes from "./routes/public/AuthRoutes.js"
import PostCommentRoutes from "./routes/private/PostCommentRoutes.js"
import UserPrivateRoutes from "./routes/private/UserPrivateRoutes.js"
import UserPublicRoutes from "./routes/public/UserPublicRoutes.js"

dotenv.config()

const app = express()

// app.use(cors())

app.use(express.json())

app.use("/api/auth", AuthRoutes)

app.use("/api", PostCommentRoutes)
app.use("/api", UserPrivateRoutes)

app.use("/api", UserPublicRoutes)

export default app
