import express from "express"
import authRoutes from "./routes/auth_routes.js"

const app = express()

app.use("/auth", authRoutes)

export default app
